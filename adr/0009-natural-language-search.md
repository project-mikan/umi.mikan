# ADR 0009: 自然言語による日記検索機能

## ステータス
Accepted

## コンテキスト

### モチベーション

- 「去年の夏に食べたもの」「友達と行った旅行」のように、ユーザが自然な言葉で過去の日記を検索したい
- 既存のキーワード検索は完全一致・部分一致のみで、意味的に近い日記を見つけられない
- ベクトル埋め込みによる意味的検索（RAG）を用いることで、文脈・ニュアンスを考慮した検索を実現する

### 要件

- 既存のLLM機能と同様に、ユーザ自身のGemini APIトークンを使用
- `user_llms.semantic_search_enabled` フラグで機能を有効化したユーザのみ利用可能
- 自然言語のクエリを入力すると、意味的に関連する日記エントリを上位N件返す
- 日記エントリの追加・更新時にベクトルを自動生成・更新する（当日分は翌朝処理）
- 検索結果にはチャンクサマリー・スニペット・類似度スコアを含む
- ベクトル生成は非同期（Redis Pub/Sub経由）で行い、即座のレスポンスを維持する

## 決定

### アーキテクチャ概要

```
日記作成/更新
     ↓
Backend gRPC → (今日の日記はスキップ) → Redis Pub/Sub → Subscriber → Gemini Embedding API
                                              ↑                              ↓
                              DiaryEmbeddingJob (翌朝4:30 JST)   diary_embeddings テーブル (pgvector halfvec)

自然言語検索クエリ
     ↓
Frontend → Backend gRPC → Gemini Embedding API (クエリのベクトル化)
                                   ↓
                          pgvector HNSW ANN検索 (diary_embeddings) + キーワードLIKE検索（ハイブリッド）
                                   ↓
                          上位N件の日記エントリを返却
```

### データフロー

#### 埋め込みベクトルの生成

1. **日記作成・更新時**: 当日（JST）の日記はスキップ。過去日記は `diary_embedding` メッセージをRedis Pub/Subに送信
2. **DiaryEmbeddingJob（スケジューラ）**: 毎朝4:30 JST に前日分の日記をRedis Pub/Sub経由でキューに追加
3. **Subscriber**: メッセージを受信し、`semantic_search_enabled = true` のユーザのみ処理
4. **チャンク分割**: `gemini-2.5-flash-lite` で日記を話題ごとのチャンクに分割（失敗時は全文を1チャンクにフォールバック）
6. **Gemini API**: 各チャンクに日付コンテキスト（`YYYY年MM月DD日の日記:\n{chunk}`）を付与してembedding生成
7. **Database**: `diary_embeddings` テーブルにチャンク単位でUPSERT（既存チャンクを削除してから再挿入）

#### 検索時

1. **Frontend**: 検索ボックスに自然言語クエリを入力し、`SearchDiaryEntriesSemantic` RPCを呼び出し
2. **Backend**: クエリテキストをGemini Embedding APIでベクトル化（同期処理・クエリ用タスクタイプ）
3. **Database**: pgvectorのHNSW（ef_search=100）でコサイン類似度ANN検索、各日記で最も類似度の高いチャンク1件を取得（日記単位に集約）
4. **ハイブリッド検索**: ベクトル検索の結果にキーワードLIKE検索結果をマージ（固有名詞・専門語のカバー）
5. **スニペット**: マッチしたチャンクの `chunk_content` を先頭200文字で切り取り
6. **Frontend**: 検索結果をスニペット・チャンクサマリー・スコア付きで表示

### データベース設計

#### 新規拡張: pgvector

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

#### 新規テーブル: `diary_embeddings`

```sql
CREATE TABLE diary_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    diary_id UUID NOT NULL REFERENCES diaries(id) ON DELETE CASCADE,
    -- diaries.user_id から導出可能だが、JOIN なしのユーザースコープフィルタのための意図的な非正規化
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL DEFAULT 0,                          -- 日記内チャンクインデックス（0始まり）
    chunk_content TEXT NOT NULL DEFAULT '',                      -- チャンクテキスト（スニペット表示用）
    chunk_summary TEXT NOT NULL DEFAULT '',                      -- チャンク概要（1〜2文）
    embedding halfvec(3072) NOT NULL,                            -- Gemini gemini-embedding-001 のネイティブ次元数
    model_version TEXT NOT NULL DEFAULT 'gemini-embedding-001', -- embedding生成モデル
    chunk_model_version TEXT NOT NULL DEFAULT 'gemini-2.5-flash-lite', -- チャンク分割モデル
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(diary_id, chunk_index)
);

-- コサイン類似度でのANN検索インデックス（HNSWはivfflatより行数制限がなく安定）
CREATE INDEX idx_diary_embeddings_embedding ON diary_embeddings
    USING hnsw (embedding halfvec_cosine_ops)
    WITH (m = 16, ef_construction = 64);

CREATE INDEX idx_diary_embeddings_user_id ON diary_embeddings(user_id);
```

#### 新規テーブル: `semantic_search_logs`

```sql
CREATE TABLE semantic_search_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

意味的検索リクエストのメトリクス集計用。

#### 既存テーブル変更: `user_llms`

```sql
ALTER TABLE user_llms ADD COLUMN semantic_search_enabled BOOLEAN NOT NULL DEFAULT FALSE;
```

このフラグが `true` のユーザのみ、embedding生成・意味的検索が実行される。

### API設計

#### gRPC Service

```protobuf
message SearchDiaryEntriesSemanticRequest {
  string query = 1;  // 自然言語クエリ
  int32 limit = 2;   // 上位何件返すか (default: 10, max: 50)
  // user_id は認証情報から取得
}

message SemanticSearchResult {
  string diary_id = 1;
  YMD date = 2;
  string snippet = 3;       // チャンクの先頭200文字
  float similarity = 4;     // コサイン類似度スコア (0.0〜1.0)
  string chunk_summary = 5; // マッチしたチャンクの概要（1〜2文）
  int32 chunk_count = 6;    // 日記内のチャンク総数
}

message SearchDiaryEntriesSemanticResponse {
  repeated SemanticSearchResult results = 1;
  string embedding_model = 2; // embedding生成に使用したモデル
  string chunk_model = 3;     // チャンク分割に使用したモデル
}

// バックフィル用
message RegenerateAllEmbeddingsRequest {}
message RegenerateAllEmbeddingsResponse {
  bool success = 1;
  int32 queued_count = 2; // キューに追加した日記数
}

// インデックス状態確認用（デバッグ・設定画面）
message GetDiaryEmbeddingStatusRequest {
  string diary_id = 1;
}
message GetDiaryEmbeddingStatusResponse {
  bool indexed = 1;
  string model_version = 2;
  int64 created_at = 3;
  int64 updated_at = 4;
  repeated float embedding_values = 5; // ベクトル値プレビュー（先頭10件のみ）
  // ...
}

service DiaryService {
  // 既存のRPCは省略
  rpc SearchDiaryEntriesSemantic(SearchDiaryEntriesSemanticRequest) returns (SearchDiaryEntriesSemanticResponse);
  rpc RegenerateAllEmbeddings(RegenerateAllEmbeddingsRequest) returns (RegenerateAllEmbeddingsResponse);
  rpc GetDiaryEmbeddingStatus(GetDiaryEmbeddingStatusRequest) returns (GetDiaryEmbeddingStatusResponse);
}
```

#### Redis Pub/Sub メッセージフォーマット

チャネル: `diary_events`

```json
{
  "type": "diary_embedding",
  "user_id": "uuid",
  "diary_id": "uuid"
}
```

### フロントエンド実装

#### 検索UI

- 既存の検索フォームに「意味的検索」トグルを追加
- 通常検索（キーワード）と意味的検索を切り替え可能
- 検索結果カードにチャンクサマリーと類似度スコアを表示
- 使用モデル（embedding_model / chunk_model）をフッターに表示

### バックエンド実装

#### SearchDiaryEntriesSemantic RPC

1. `user_llms.semantic_search_enabled` を確認（falseなら `FailedPrecondition` エラー）
2. クエリテキストをGemini Embedding API（RETRIEVAL_QUERY）でベクトル化
3. ReadOnlyトランザクション内で `SET LOCAL hnsw.ef_search = 100` を設定
4. `diary_embeddings` テーブルでコサイン類似度ANN検索（閾値 0.4 以上）
5. ウィンドウ関数（`ROW_NUMBER()`）で日記単位に集約し、最良チャンクを1件選択
6. キーワードLIKE検索の結果をマージ（ハイブリッド検索）
7. `chunk_content` の先頭200文字をスニペットとして返却
8. `semantic_search_logs` にリクエストを記録

#### Subscriber ハンドラ

メッセージタイプ: `diary_embedding`

処理内容:
1. `user_llms` から APIキーと `semantic_search_enabled` を確認（falseならスキップ）
2. 対象日記の本文・日付を取得
3. `SplitDiaryIntoChunks` で話題ごとに分割（失敗時は全文1チャンクにフォールバック）
4. 各チャンクに日付コンテキスト（`YYYY年MM月DD日の日記:\n`）を付与してembedding生成
5. `diary_embeddings` テーブルにチャンク単位でUPSERT（既存チャンク削除→再挿入トランザクション）

#### DiaryEmbeddingJob（スケジューラ）

- 毎朝4:30 JST に実行
- `semantic_search_enabled = true` の全ユーザの前日分日記を対象
- embedding未生成または日記更新後に再生成が必要なものをキューに追加

#### 今日の日記のDeferral

日記作成・更新時（`publishDiaryEmbeddingMessage`）は、当日（JST）の日記に対してはPub/Subへの送信をスキップ。翌朝4:30 JSTのスケジューラジョブが処理することで、編集中の日記に対して無駄なembedding生成を避ける。

### LLMモデル設計

#### 埋め込みモデル

```
モデル: gemini-embedding-001 (Gemini)
入力: "{YYYY}年{MM}月{DD}日の日記:\n{chunk_content}"  -- 日付コンテキストを付与
次元数: 3072 (ネイティブ次元、MRL削減なし)
型: halfvec (pgvector、HNSWで4000次元まで対応)
タスクタイプ: RETRIEVAL_DOCUMENT (ドキュメント側)
             RETRIEVAL_QUERY (クエリ側)
```

#### チャンク分割モデル

```
モデル: gemini-2.5-flash-lite
処理: 日記を話題ごとのチャンクに分割し、各チャンクに概要（1〜2文）を付与
失敗時フォールバック: 日記全文を1チャンクとして扱う
```

#### スニペット生成

- マッチしたチャンクの `chunk_content` をそのまま先頭200文字で切り取り
- チャンク分割の時点で話題単位に分かれているため、チャンク内容がそのままスニペットとして有用

### コード生成

xoが `diaryembedding.dbtpl.go` を生成（`Halfvec` 型でembeddingカラムを保持）。
複雑なベクトル検索・チャンクUPSERTは `diary_embeddings.go` にカスタムクエリとして実装し、xo生成コードと共存。

### モニタリング

#### Prometheusメトリクス（バックエンドサービス）

```
backend_semantic_search_requests_total{status="success|failure"}
backend_semantic_search_duration_seconds{status="success|failure"}
backend_semantic_search_results_count
```

#### Prometheusメトリクス（Subscriber）

```
messages_processed_total{type="diary_embedding", status="success|error"}
processing_duration_seconds{type="diary_embedding"}
summaries_generated_total{type="diary_embedding"}
```

## コスト見積もり

> Gemini API料金（2025年時点、Pay-as-you-go）を基準に算出。
> - `gemini-embedding-001`: $0.00015 / 1K tokens（入力のみ）
> - `gemini-2.5-flash-lite`: チャンク分割用（入力 + 出力）

### 前提

| 項目 | 想定値 |
|---|---|
| 日記の平均文字数 | 1,000文字（日本語）|
| 日本語のトークン換算 | 1文字 ≈ 1.5トークン → **1,500トークン/件** |
| 月間日記投稿数 | 20件/月（週5日ペース）|
| 月間意味的検索回数 | 20回/月 |
| チャンク数（平均） | 3チャンク/日記 |

### 月間合計コスト試算

| シナリオ | 月間投稿数 | 月間検索数 | 月額概算 |
|---|---|---|---|
| ライト | 10件 | 5回 | **$0.001〜** |
| スタンダード | 20件 | 20回 | **$0.002〜** |
| ヘビー | 30件 | 50回 | **$0.003〜** |

### 初回バックフィルコスト（過去日記の一括埋め込み生成）

`RegenerateAllEmbeddings` RPCで過去日記をキューに追加して処理。
過去1年分（365件 × 平均3チャンク）を一括インデックス化した場合のコストは微小。

## 結果

### メリット

1. **直感的な検索体験**: キーワードを思い出せなくても、ニュアンスで検索できる
2. **チャンク分割による精度向上**: 長い日記でも話題単位で検索でき、無関係な話題のノイズが低減
3. **ハイブリッド検索**: ベクトル検索 + キーワードLIKE検索で固有名詞・専門語もカバー
4. **既存アーキテクチャの活用**: Redis Pub/SubパターンとGemini APIの再利用
5. **非同期ベクトル生成**: 日記投稿時のレスポンスに影響しない
6. **pgvectorによる高速検索**: PostgreSQL内でベクトル検索が完結し、外部サービス不要

### デメリット

1. **pgvector依存**: PostgreSQL拡張のインストールが必要
2. **embedding生成コスト**: 日記作成・更新ごとにAPI呼び出しが発生（チャンク分割も含む）
3. **検索レイテンシ**: クエリのベクトル化に同期APIコールが必要（数百ms）
4. **過去日記の初回インデックス**: 既存日記のバックフィル処理が必要（`RegenerateAllEmbeddings` RPCで実施）

### トレードオフ

- **ベクトル生成タイミング**: 当日はスキップ・翌朝処理を選択
  - 代替案: 即時同期処理（シンプルだが編集中の日記に無駄なAPI呼び出しが発生）
- **チャンク分割**: LLMによる話題ベース分割を選択
  - 代替案: 固定長チャンク分割（シンプルだが話題の境界を無視する）
- **インデックス**: HNSWを選択（ivfflatより行数制限がなく安定）
  - 代替案: ivfflat（検索精度・速度のトレードオフが異なる）
- **ハイブリッド検索**: ベクトル + キーワードLIKEを選択
  - 代替案: ベクトルのみ（固有名詞・専門語の見落としリスクがある）
- **ベクトルDB**: pgvectorを選択
  - 代替案: Pinecone/Weaviateなど専用ベクトルDB（運用コスト・複雑性の増加を避ける）
- **埋め込みモデル**: Gemini gemini-embedding-001 を選択（3072次元、MRL削減なし）
  - 代替案: OpenAI text-embedding-3-small（ユーザは既にGeminiのAPIキーを持つため統一）

## 参考資料

- ADR 0003: LLM統合
- ADR 0004: Redis Pub/Sub実装
- ADR 0008: 日記エントリのLLMハイライト機能

## 実装チェックリスト

### データベース

- [x] pgvector拡張の有効化（schema/5000_diary_embeddings.sql）
- [x] `diary_embeddings` テーブル作成（チャンク対応: `chunk_index`, `chunk_content`, `chunk_summary`, `chunk_model_version`）
- [x] HNSWインデックス作成（`halfvec_cosine_ops`, m=16, ef_construction=64）
- [x] xoコード生成（`diaryembedding.dbtpl.go`、`Halfvec` 型でembeddingカラムを保持）
- [x] カスタムクエリ実装（`diary_embeddings.go`：UPSERT・ベクトル検索・ステータス取得）
- [x] `user_llms.semantic_search_enabled` カラム追加
- [x] `semantic_search_logs` テーブル作成

### バックエンド

- [x] `SearchDiaryEntriesSemantic` RPC実装（ハイブリッド検索、閾値0.4、HNSW ef_search=100）
- [x] `RegenerateAllEmbeddings` RPC実装（バックフィル用、分散ロック付き）
- [x] `GetDiaryEmbeddingStatus` RPC実装（インデックス状態確認）
- [x] Gemini Embedding APIクライアント実装（`GenerateEmbedding` メソッド、3072次元）
- [x] チャンク分割クライアント実装（`SplitDiaryIntoChunks` メソッド、`gemini-2.5-flash-lite`）
- [x] Subscriber: `generateDiaryEmbedding` 実装（チャンク分割・日付付与・マークダウン除去）
- [x] 今日の日記のDeferral実装（`isTodayJST` / `publishDiaryEmbeddingMessage`）
- [x] スニペット生成ロジック実装（`chunk_content` 先頭200文字）
- [x] `DiaryEmbeddingJob` 実装（毎朝4:30 JST、前日分を処理）
- [x] Prometheusメトリクス追加
- [x] テスト作成

### フロントエンド

- [x] 意味的検索トグル追加（キーワード/意味的切り替えボタン）
- [x] 検索結果カードコンポーネント更新（スニペット・チャンクサマリー・スコア・使用モデル表示）
- [x] i18n対応（ja.json, en.json）
- [x] エラーハンドリング実装
- [x] ローディング状態の実装（`navigating` ストアでボタンにスピナー表示）

### インフラ

- [x] compose.yml: pgvector対応のPostgreSQLイメージへ変更（`pgvector/pgvector:pg17`）
- [x] Grafana ダッシュボード作成（`monitoring/grafana/dashboards/umi-mikan-rag.json`）

### ドキュメント

- [x] CLAUDE.md 更新
