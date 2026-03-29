# ADR 0011: 人間関係の可視化機能

## ステータス
Proposed

## コンテキスト

### モチベーション

- 日記に登場する人物（家族・友人・同僚など）との関わりを振り返りたい
- 「最近、誰と一番多く過ごしているか」「ある人との関係が変化しているか」を把握したい
- 人物名の手動管理なしに、LLMが日記から自動で人物を抽出・分類する
- 人間関係をグラフ（ノード・エッジ）で可視化することで、直感的な理解を促す

### 要件

- 既存のLLM機能と同様に、ユーザ自身のGemini APIトークンを使用
- 日記エントリから人物名・関係性・感情を自動抽出（非同期）
- 抽出結果を基に「人間関係グラフ」をフロントエンドで可視化
- 期間フィルタリングが可能（直近30日・直近90日・全期間など）
- 同一人物の名前揺れ（「田中さん」「田中」「Tanaka」）を可能な範囲で統合
- プライバシーを考慮し、人物データはユーザ自身のみが閲覧可能

## 決定

### アーキテクチャ概要

```
日記作成/更新
     ↓
Backend gRPC → Redis Pub/Sub → Subscriber → Gemini API
                                                  ↓
                                   person_mentions テーブル
                                   person_relationships テーブル

人間関係グラフ表示
Frontend (/people) → Backend gRPC → person_mentions / person_relationships テーブル
                                                ↓
                                    グラフデータ（JSON）を返却
                                                ↓
                                    Frontend (D3.js / Cytoscape.js でレンダリング)
```

### データフロー

#### 人物抽出（日記作成・更新時）

1. **Backend**: 日記作成・更新時に `person_extraction` メッセージをRedis Pub/Subに送信
2. **Subscriber**: メッセージを受信し、日記本文をGemini APIで解析
3. **Gemini API**: 人物名・関係性カテゴリ・感情トーンを返却
4. **Database**: `person_mentions` テーブルにUPSERTで保存

#### グラフ表示

1. **Frontend**: 人間関係グラフページを開く
2. **Backend**: `GetRelationshipGraph` RPCで期間・フィルタ条件を受け取る
3. **Database**: 期間内の `person_mentions` を集計してグラフデータを生成
4. **Frontend**: グラフをD3.jsでインタラクティブに表示

### データベース設計

#### 新規テーブル: `person_mentions`

```sql
CREATE TABLE person_mentions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    diary_id UUID NOT NULL REFERENCES diaries(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    person_name TEXT NOT NULL,             -- 日記中の人物名（正規化済み）
    person_alias TEXT NOT NULL,            -- 日記中の実際の表記
    relationship_category TEXT NOT NULL,   -- 'family' | 'friend' | 'colleague' | 'romantic' | 'other'
    emotional_tone TEXT NOT NULL,          -- 'positive' | 'neutral' | 'negative' | 'mixed'
    context_snippet TEXT,                  -- 言及箇所の抜粋（最大150文字）
    diary_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_person_mentions_user_id ON person_mentions(user_id);
CREATE INDEX idx_person_mentions_diary_id ON person_mentions(diary_id);
CREATE INDEX idx_person_mentions_person_name ON person_mentions(user_id, person_name);
CREATE INDEX idx_person_mentions_diary_date ON person_mentions(user_id, diary_date);
```

#### 新規テーブル: `person_aliases`

```sql
-- 名前揺れ統合のためのエイリアス管理テーブル
CREATE TABLE person_aliases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    canonical_name TEXT NOT NULL,   -- 正規名（ユーザが設定 or LLMが推定）
    alias TEXT NOT NULL,            -- 別名・揺れ表記
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, alias)
);

CREATE INDEX idx_person_aliases_user_id ON person_aliases(user_id);
```

### API設計

#### gRPC Service

```protobuf
message PersonNode {
  string name = 1;
  string relationship_category = 2;
  int32 mention_count = 3;
  float positive_ratio = 4;    // ポジティブな言及の割合
  float negative_ratio = 5;
  google.protobuf.Timestamp first_mentioned = 6;
  google.protobuf.Timestamp last_mentioned = 7;
}

message PersonEdge {
  // ユーザ自身を中心ノードとする星型グラフのためエッジはユーザ→人物のみ
  string person_name = 1;
  int32 weight = 2;            // 言及回数（エッジの太さに対応）
  string dominant_emotion = 3; // 全体的な感情傾向
}

message GetRelationshipGraphRequest {
  google.protobuf.Timestamp period_start = 1;
  google.protobuf.Timestamp period_end = 2;
  string relationship_category = 3; // フィルタ（空文字の場合は全カテゴリ）
  int32 min_mention_count = 4;      // 最低言及回数（デフォルト: 1）
}

message GetRelationshipGraphResponse {
  repeated PersonNode nodes = 1;
  repeated PersonEdge edges = 2;
  int32 total_diary_count = 3;
  int32 total_person_count = 4;
}

message GetPersonDetailRequest {
  string person_name = 1;
  google.protobuf.Timestamp period_start = 2;
  google.protobuf.Timestamp period_end = 3;
}

message PersonMentionEntry {
  string diary_id = 1;
  google.protobuf.Timestamp diary_date = 2;
  string context_snippet = 3;
  string emotional_tone = 4;
}

message GetPersonDetailResponse {
  string person_name = 1;
  string relationship_category = 2;
  int32 mention_count = 3;
  repeated string aliases = 4;
  repeated PersonMentionEntry recent_mentions = 5; // 直近10件
  float positive_ratio = 6;
  float negative_ratio = 7;
  float neutral_ratio = 8;
  string emotion_trend = 9; // 'improving' | 'declining' | 'stable'
}

message UpdatePersonAliasRequest {
  string alias = 1;
  string canonical_name = 2; // 空文字の場合は統合を解除
}

message UpdatePersonAliasResponse {
  bool success = 1;
}

service DiaryService {
  // 既存のRPCは省略
  rpc GetRelationshipGraph(GetRelationshipGraphRequest) returns (GetRelationshipGraphResponse);
  rpc GetPersonDetail(GetPersonDetailRequest) returns (GetPersonDetailResponse);
  rpc UpdatePersonAlias(UpdatePersonAliasRequest) returns (UpdatePersonAliasResponse);
}
```

#### Redis Pub/Sub メッセージフォーマット

チャネル: `diary_events`

```json
{
  "type": "person_extraction",
  "user_id": "uuid",
  "diary_id": "uuid"
}
```

### フロントエンド実装

#### 人間関係グラフページ（新規ルート: `/people`）

- **グラフビュー**: D3.js Force-Directed Graph
  - 中心ノード: ユーザ自身
  - 周辺ノード: 登場人物（言及回数に比例したサイズ）
  - エッジ: 言及回数に比例した太さ
  - ノード色: 感情傾向（ポジティブ=青、ネガティブ=赤、ニュートラル=グレー）
  - ノードクリック: 人物詳細パネルを表示

- **フィルタパネル**:
  - 期間スライダー（直近30日 / 直近90日 / 直近1年 / 全期間）
  - カテゴリフィルタ（家族 / 友人 / 同僚 / その他）
  - 最低言及回数スライダー（ノイズ除去）

- **人物詳細パネル**（サイドパネル）:
  - 人物名・カテゴリバッジ
  - 感情トレンドグラフ（時系列）
  - 直近の言及コンテキスト一覧
  - 別名管理（名前揺れ統合UI）

#### PWA対応

- グラフデータのオフラインキャッシュ
- 最終取得日時の表示

### バックエンド実装

#### Subscriber ハンドラ

新規ハンドラ: `PersonExtractionHandler`

- メッセージタイプ: `person_extraction`
- 処理内容:
  1. 対象日記エントリのタイトルと本文を取得
  2. ユーザのLLM設定（APIキー）を確認
  3. Gemini APIに人物抽出リクエストを送信
  4. レスポンスをパース
  5. `person_aliases` テーブルを参照して名前揺れを統合
  6. `person_mentions` テーブルにUPSERT（diary_id単位で既存データを削除後に再挿入）
  7. メトリクスを記録

#### GetRelationshipGraph RPC

1. 期間内の `person_mentions` を `person_name` でGROUP BY集計
2. 言及回数・感情割合を計算
3. グラフのノード・エッジデータに変換して返却

### LLMプロンプト設計

```
以下の日記から、登場する人物を抽出してください。

【日記本文】
{diary_content}

【抽出ルール】
- 実在の人物のみ抽出（架空の人物・著者は除く）
- 日記の書き手自身は抽出しない
- 固有名詞として識別できる場合のみ抽出（「誰か」「その人」などは除く）

【出力形式】
JSON配列形式で、以下の項目を返してください:
- person_name: 人物名（姓名、または呼称）
- relationship_category: 'family' | 'friend' | 'colleague' | 'romantic' | 'other'
- emotional_tone: その人物への言及全体の感情 'positive' | 'neutral' | 'negative' | 'mixed'
- context_snippet: その人物が登場する代表的な文章（最大150文字）

人物が登場しない場合は空配列 [] を返してください。
出力はJSON形式のみとし、説明文は不要です。
```

#### 期待される出力例

```json
[
  {
    "person_name": "田中さん",
    "relationship_category": "colleague",
    "emotional_tone": "positive",
    "context_snippet": "田中さんに今日のプレゼンを褒めてもらって素直に嬉しかった。"
  },
  {
    "person_name": "母",
    "relationship_category": "family",
    "emotional_tone": "neutral",
    "context_snippet": "母から電話があって、実家の庭に花が咲いたと聞いた。"
  }
]
```

### モニタリング

#### Prometheus メトリクス

```
person_extraction_processing_total{status="success|failure"}
person_extraction_processing_duration_seconds
person_extraction_persons_count_histogram  -- 1日記あたりの抽出人物数
person_extraction_api_calls_total{provider="gemini"}
person_extraction_api_errors_total{provider="gemini",error_type="..."}
```

## 結果

### メリット

1. **人間関係の客観的把握**: 日記を通じて自分が誰と多く関わっているかを可視化できる
2. **感情変遷の追跡**: 特定の人物との関係性が時間とともにどう変化したか把握できる
3. **既存アーキテクチャの活用**: Pub/Sub・Gemini APIの再利用
4. **プライバシー保護**: データはユーザ自身のみが閲覧可能

### デメリット

1. **名前揺れ問題**: 同一人物の異なる表記を完全に自動統合することは困難
2. **LLM精度依存**: 文脈から関係性カテゴリを誤分類する可能性がある
3. **グラフライブラリの追加**: D3.jsまたはCytoscape.jsの導入によるバンドルサイズ増加
4. **センシティブな情報**: 人物名・感情データはプライバシーに配慮した管理が必要

### トレードオフ

- **グラフ構造**: 星型（ユーザ中心）を選択
  - 代替案: 人物間の関係も含めた完全グラフ（LLMで人物間関係を推定する必要があり複雑）
- **名前揺れ統合**: LLM推定 + 手動修正を選択
  - 代替案: 完全自動統合（精度の限界から手動修正手段も必要）
- **グラフライブラリ**: D3.js Force-Directed Graphを選択
  - 代替案: Cytoscape.js（より高機能だが大きい）、Three.js（3Dグラフ）
- **データ保持**: DBに永続化を選択
  - 代替案: Redisキャッシュ（揮発性のため不採用）

## 参考資料

- ADR 0003: LLM統合
- ADR 0004: Redis Pub/Sub実装
- ADR 0008: 日記エントリのLLMハイライト機能
- ADR 0009: 自然言語による日記検索機能

## 実装チェックリスト

### データベース

- [ ] `person_mentions` テーブル作成（マイグレーション）
- [ ] `person_aliases` テーブル作成（マイグレーション）
- [ ] インデックス作成
- [ ] xoコード生成

### バックエンド

- [ ] `GetRelationshipGraph` RPC実装
- [ ] `GetPersonDetail` RPC実装
- [ ] `UpdatePersonAlias` RPC実装
- [ ] Subscriber: `PersonExtractionHandler` 実装
- [ ] LLMプロンプト作成とテスト
- [ ] 名前揺れ統合ロジック実装
- [ ] Prometheus メトリクス追加
- [ ] テスト作成

### フロントエンド

- [ ] D3.jsまたはCytoscape.jsのインストール（pnpm）
- [ ] `/people` ルート作成
- [ ] Force-Directed Graphコンポーネント実装
- [ ] フィルタパネルコンポーネント実装
- [ ] 人物詳細サイドパネル実装
- [ ] 別名管理UIの実装
- [ ] i18n対応（ja.json, en.json）
- [ ] エラーハンドリング実装
- [ ] PWAオフラインキャッシュ対応

### インフラ・モニタリング

- [ ] Grafana ダッシュボード更新
- [ ] Prometheus スクレイプ設定確認

### ドキュメント

- [ ] CLAUDE.md 更新
