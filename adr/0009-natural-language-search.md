# ADR 0009: 自然言語による日記検索機能

## ステータス
Proposed

## コンテキスト

### モチベーション

- 「去年の夏に食べたもの」「友達と行った旅行」のように、ユーザが自然な言葉で過去の日記を検索したい
- 既存のキーワード検索は完全一致・部分一致のみで、意味的に近い日記を見つけられない
- ベクトル埋め込みによる意味的検索（RAG）を用いることで、文脈・ニュアンスを考慮した検索を実現する

### 要件

- 既存のLLM機能と同様に、ユーザ自身のGemini APIトークンを使用
- 自然言語のクエリを入力すると、意味的に関連する日記エントリを上位N件返す
- 日記エントリの追加・更新時にベクトルを自動生成・更新する
- 検索結果には日付・タイトル・関連スニペットを含む
- ベクトル生成は非同期（Redis Pub/Sub経由）で行い、即座のレスポンスを維持する

## 決定

### アーキテクチャ概要

```
日記作成/更新
     ↓
Backend gRPC → Redis Pub/Sub → Subscriber → Gemini Embedding API
                                                      ↓
                                              diary_embeddings テーブル (pgvector)

自然言語検索クエリ
     ↓
Frontend → Backend gRPC → Gemini Embedding API (クエリのベクトル化)
                                   ↓
                          pgvector ANN検索 (diary_embeddings)
                                   ↓
                          上位N件の日記エントリを返却
```

### データフロー

#### 埋め込みベクトルの生成

1. **日記作成・更新時**: `diary_embedding` メッセージをRedis Pub/Subに送信
2. **Subscriber**: メッセージを受信し、日記本文をGemini Embedding APIに送信
3. **Gemini API**: テキストの埋め込みベクトル（1536次元）を返却
4. **Database**: `diary_embeddings` テーブルにUPSERTで保存

#### 検索時

1. **Frontend**: 検索ボックスに自然言語クエリを入力し、`SearchDiaryEntriesSemantic` RPCを呼び出し
2. **Backend**: クエリテキストをGemini Embedding APIでベクトル化（同期処理）
3. **Database**: pgvectorのコサイン類似度でANN検索、上位N件を取得
4. **Frontend**: 検索結果を日付・スニペット付きで表示

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
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    embedding vector(1536) NOT NULL, -- Gemini text-embedding-004 の次元数
    model_version TEXT NOT NULL DEFAULT 'text-embedding-004',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(diary_id)
);

-- コサイン類似度でのANN検索インデックス
CREATE INDEX idx_diary_embeddings_embedding ON diary_embeddings
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE INDEX idx_diary_embeddings_user_id ON diary_embeddings(user_id);
```

### API設計

#### gRPC Service

```protobuf
message SearchDiaryEntriesSemanticRequest {
  string query = 1;        // 自然言語クエリ
  int32 limit = 2;         // 上位何件返すか (default: 10, max: 50)
  // user_id は認証情報から取得
}

message SemanticSearchResult {
  string diary_id = 1;
  google.protobuf.Timestamp date = 2;
  string title = 3;
  string snippet = 4;      // クエリに関連する抜粋（最大200文字）
  float similarity = 5;    // コサイン類似度スコア (0.0〜1.0)
}

message SearchDiaryEntriesSemanticResponse {
  repeated SemanticSearchResult results = 1;
}

service DiaryService {
  // 既存のRPCは省略
  rpc SearchDiaryEntriesSemantic(SearchDiaryEntriesSemanticRequest) returns (SearchDiaryEntriesSemanticResponse);
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
- 検索結果カードに類似度スコアバッジを表示
- 関連スニペット部分をハイライト表示

#### 状態管理

- 検索中のローディング状態
- ベクトル未生成の日記がある場合の警告表示
- 検索履歴のローカルストレージ保存

### バックエンド実装

#### SearchDiaryEntriesSemantic RPC

1. クエリテキストをGemini Embedding APIでベクトル化
2. `diary_embeddings` テーブルでコサイン類似度ANN検索
3. 類似度スコア閾値（0.5以上）でフィルタリング
4. 対応する日記エントリの詳細を取得
5. スニペット生成（クエリに最も関連する部分を抽出）

#### Subscriber ハンドラ

新規ハンドラ: `DiaryEmbeddingHandler`

- メッセージタイプ: `diary_embedding`
- 処理内容:
  1. 対象日記エントリのタイトルと本文を取得
  2. ユーザのLLM設定（APIキー）を確認
  3. Gemini Embedding APIにテキストを送信
  4. 取得したベクトルを `diary_embeddings` テーブルにUPSERT
  5. メトリクスを記録

### LLMプロンプト設計

#### 埋め込みモデル

```
モデル: text-embedding-004 (Gemini)
入力: "{title}\n\n{content}"  -- タイトルと本文を結合してコンテキストを最大化
次元数: 1536
タスクタイプ: RETRIEVAL_DOCUMENT (ドキュメント側)
             RETRIEVAL_QUERY (クエリ側)
```

#### スニペット生成

- 日記本文をセンテンス単位で分割
- 各センテンスのベクトルとクエリベクトルのコサイン類似度を計算
- 最も類似度の高いセンテンスを最大200文字で抜粋

### モニタリング

#### Prometheus メトリクス

```
diary_embedding_processing_total{status="success|failure"}
diary_embedding_processing_duration_seconds
diary_semantic_search_total
diary_semantic_search_duration_seconds
diary_embedding_api_calls_total{provider="gemini"}
diary_embedding_api_errors_total{provider="gemini",error_type="..."}
```

## 結果

### メリット

1. **直感的な検索体験**: キーワードを思い出せなくても、ニュアンスで検索できる
2. **既存アーキテクチャの活用**: Redis Pub/SubパターンとGemini APIの再利用
3. **非同期ベクトル生成**: 日記投稿時のレスポンスに影響しない
4. **pgvectorによる高速検索**: PostgreSQL内でベクトル検索が完結し、外部サービス不要

### デメリット

1. **pgvector依存**: PostgreSQL拡張のインストールが必要
2. **埋め込み生成コスト**: 日記作成・更新ごとにAPI呼び出しが発生
3. **検索レイテンシ**: クエリのベクトル化に同期APIコールが必要（数百ms）
4. **過去日記の初回インデックス**: 既存日記のバックフィル処理が必要

### トレードオフ

- **ベクトル生成タイミング**: 非同期（Pub/Sub経由）を選択
  - 代替案: 同期処理（シンプルだが日記投稿レスポンスが遅延）
- **検索処理**: 同期処理を選択
  - 代替案: 非同期＋ポーリング（UI複雑化を避けるため同期を採用）
- **ベクトルDB**: pgvectorを選択
  - 代替案: Pinecone/Weaviateなど専用ベクトルDB（運用コスト・複雑性の増加を避ける）
- **埋め込みモデル**: Gemini text-embedding-004 を選択
  - 代替案: OpenAI text-embedding-3-small（ユーザは既にGeminiのAPIキーを持つため統一）

## 参考資料

- ADR 0003: LLM統合
- ADR 0004: Redis Pub/Sub実装
- ADR 0008: 日記エントリのLLMハイライト機能

## 実装チェックリスト

### データベース

- [ ] pgvector拡張の有効化（マイグレーション）
- [ ] `diary_embeddings` テーブル作成
- [ ] ivfflatインデックス作成
- [ ] xoコード生成

### バックエンド

- [ ] `SearchDiaryEntriesSemantic` RPC実装
- [ ] Gemini Embedding APIクライアント実装
- [ ] Subscriber: `DiaryEmbeddingHandler` 実装
- [ ] スニペット生成ロジック実装
- [ ] 既存日記のバックフィルスクリプト作成
- [ ] Prometheus メトリクス追加
- [ ] テスト作成

### フロントエンド

- [ ] 意味的検索トグル追加
- [ ] 検索結果カードコンポーネント更新（スニペット・スコア表示）
- [ ] ローディング状態の実装
- [ ] i18n対応（ja.json, en.json）
- [ ] エラーハンドリング実装

### インフラ

- [ ] compose.yml: pgvector対応のPostgreSQLイメージへ変更
- [ ] Grafana ダッシュボード更新

### ドキュメント

- [ ] CLAUDE.md 更新
