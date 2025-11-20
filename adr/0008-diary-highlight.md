# ADR 0008: 日記エントリのLLMハイライト機能

## ステータス
Proposed

## コンテキスト

### モチベーション

- 日記を読み返す際、重要な箇所や印象的な出来事を素早く把握したい
- 長文の日記エントリでも、ハイライトされた部分を中心に効率的に振り返りたい
- ユーザ自身が手動でハイライトするのではなく、LLMの判断で重要箇所を自動抽出したい

### 要件

- 既存のLLM機能と同様に、ユーザ自身のGemini APIトークンを使用
- 個別の日記詳細ページに「LLMハイライト」ボタンを配置
- ハイライトは日記本文を変更せず、視覚的に黄色でマーキング
- 重い処理のため、Redis Pub/Sub経由で非同期実行
- サーバーは即座に200レスポンスを返し、処理完了後にフロントエンドで反映
- 日記本文が更新された場合、古いハイライトは自動的に無効化

## 決定

### アーキテクチャ概要

```
Frontend (ボタンクリック) → Backend gRPC → Redis Pub/Sub → Subscriber → Gemini API
                                ↓                                           ↓
                           200 Response                                Database
                                                                           ↓
                                                                    Frontend (再取得)
```

### データフロー

1. **Frontend**: 日記詳細ページで「LLMハイライト」ボタンをクリック
   - `TriggerDiaryHighlight` RPCを呼び出し
   - 即座に200レスポンスを受け取り、処理中表示を開始

2. **Backend**: gRPCリクエストを受信
   - Redis Pub/Subに `diary_highlight` メッセージを送信
   - 即座にクライアントに200レスポンスを返す

3. **Subscriber**: メッセージを受信して処理を実行
   - 対象日記エントリと既存ハイライトを確認
   - Gemini APIにハイライト生成リクエストを送信
   - 結果を `diary_highlights` テーブルに保存

4. **Frontend**: ポーリングまたはWebSocket経由で完了を検知
   - `GetDiaryHighlight` RPCで最新のハイライト情報を取得
   - 日記本文にハイライトを適用して表示

### データベース設計

#### 新規テーブル: `diary_highlights`

```sql
CREATE TABLE diary_highlights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    diary_id UUID NOT NULL REFERENCES diaries(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    highlights JSONB NOT NULL, -- ハイライト情報の配列
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(diary_id)
);

CREATE INDEX idx_diary_highlights_diary_id ON diary_highlights(diary_id);
CREATE INDEX idx_diary_highlights_user_id ON diary_highlights(user_id);
```

#### ハイライトデータ構造 (JSONB)

```json
[
  {
    "start": 0,      // ハイライト開始位置(文字数)
    "end": 25,       // ハイライト終了位置(文字数)
    "text": "今日はとても良い天気だった。"  // ハイライト対象のテキスト
  },
  {
    "start": 120,
    "end": 180,
    "text": "友人との会話で新しい気づきがあり、とても有意義な時間だった。"
  }
]
```

#### 自動無効化ロジック

```sql
-- 日記更新時に、古いハイライトを自動削除
DELETE FROM diary_highlights
WHERE diary_id = $1
  AND updated_at < (SELECT updated_at FROM diaries WHERE id = $1);
```

または、アプリケーション層で以下のロジックを実装:

```go
// ハイライト取得時にチェック
if highlight.UpdatedAt.Before(diary.UpdatedAt) {
    // ハイライトを無効として扱う(削除またはフラグ設定)
    return nil, ErrHighlightOutdated
}
```

### API設計

#### gRPC Service

```protobuf
// diary.proto

message TriggerDiaryHighlightRequest {
  string diary_id = 1;
  // user_id は認証情報から取得
}

message TriggerDiaryHighlightResponse {
  bool queued = 1;  // キューイング成功/失敗
  string message = 2;
}

message GetDiaryHighlightRequest {
  string diary_id = 1;
  // user_id は認証情報から取得
}

message HighlightRange {
  int32 start = 1;
  int32 end = 2;
  string text = 3;
}

message GetDiaryHighlightResponse {
  repeated HighlightRange highlights = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Timestamp updated_at = 3;
}

message DeleteDiaryHighlightRequest {
  string diary_id = 1;
}

message DeleteDiaryHighlightResponse {
  bool success = 1;
}

service DiaryService {
  // 既存のRPCは省略

  rpc TriggerDiaryHighlight(TriggerDiaryHighlightRequest) returns (TriggerDiaryHighlightResponse);
  rpc GetDiaryHighlight(GetDiaryHighlightRequest) returns (GetDiaryHighlightResponse);
  rpc DeleteDiaryHighlight(DeleteDiaryHighlightRequest) returns (DeleteDiaryHighlightResponse);
}
```

#### Redis Pub/Sub メッセージフォーマット

チャネル: `diary_events`

```json
{
  "type": "diary_highlight",
  "user_id": "uuid",
  "diary_id": "uuid"
}
```

### フロントエンド実装

#### 日記詳細ページ

- 「LLMハイライト」ボタンを配置
  - ボタンクリック → `TriggerDiaryHighlight` RPC呼び出し
  - 処理中スピナー表示
  - ポーリング(1秒間隔)で `GetDiaryHighlight` を呼び出し
  - ハイライト取得成功 → 本文に黄色マーキングを適用

- ハイライト表示
  - `<span class="highlight">` でラップ
  - CSSで黄色背景を適用
  - ハイライトがある場合、「ハイライトを削除」ボタンも表示

- エラーハンドリング
  - ハイライト生成失敗時はエラーメッセージ表示
  - 日記が更新された場合、ハイライトの再適用を促す

### バックエンド実装

#### DiaryService RPC実装

1. **TriggerDiaryHighlight**
   - 対象日記の存在確認
   - 権限チェック(ユーザ自身の日記か)
   - Redis Pub/Subにメッセージ送信
   - 即座に200レスポンス

2. **GetDiaryHighlight**
   - `diary_highlights` テーブルから取得
   - 日記の `updated_at` とハイライトの `updated_at` を比較
   - 日記が新しい場合は無効として扱う(または自動削除)

3. **DeleteDiaryHighlight**
   - ユーザが手動でハイライトを削除可能

#### Subscriber ハンドラ

新規ハンドラ: `DiaryHighlightHandler`

- メッセージタイプ: `diary_highlight`
- 処理内容:
  1. 対象日記エントリと内容を取得
  2. 既存ハイライトの確認(重複処理を防ぐ)
  3. ユーザのLLM設定(APIキー)を取得
  4. Gemini APIにハイライト生成リクエストを送信
     - プロンプト例: "以下の日記から、特に重要な部分や印象的な出来事、感情が表現されている箇所を3〜5箇所抽出し、開始位置と終了位置をJSON形式で返してください。"
  5. 結果を `diary_highlights` テーブルに保存(UPSERT)
  6. メトリクスを記録

- エラーハンドリング:
  - API呼び出し失敗時はリトライ(最大3回)
  - パース失敗時はログ出力してスキップ

### LLMプロンプト設計

#### Gemini APIへのプロンプト例

```
以下の日記から、特に重要な部分や印象的な出来事、感情が強く表現されている箇所を3〜5箇所抽出してください。

【日記本文】
{diary_content}

【出力形式】
JSON配列形式で、各ハイライトの開始位置(start)、終了位置(end)、テキスト(text)を返してください。
位置は文字数(0から始まるインデックス)で指定してください。

例:
[
  {"start": 0, "end": 25, "text": "今日はとても良い天気だった。"},
  {"start": 120, "end": 180, "text": "友人との会話で新しい気づきがあった。"}
]
```

#### 期待される出力

```json
[
  {"start": 15, "end": 45, "text": "プロジェクトが無事に完了して本当に嬉しい"},
  {"start": 120, "end": 165, "text": "チームメンバーに感謝の気持ちでいっぱいだ"},
  {"start": 230, "end": 280, "text": "次はもっと大きな挑戦をしてみたいと思った"}
]
```

### モニタリング

#### Prometheus メトリクス

Subscriberメトリクス:

```
diary_highlight_processing_total{status="success|failure"}
diary_highlight_processing_duration_seconds
diary_highlight_api_calls_total{provider="gemini"}
diary_highlight_api_errors_total{provider="gemini",error_type="..."}
```

#### Grafana ダッシュボード

既存の「Pub/Sub Monitoring」ダッシュボードに以下のパネルを追加:

- ハイライト生成処理数(成功/失敗)
- ハイライト生成処理時間
- Gemini API呼び出し数・エラー数

## 結果

### メリット

1. **ユーザ体験の向上**: 重要箇所を素早く把握でき、振り返りが効率的になる
2. **日記の不変性**: ハイライトは別テーブルで管理し、日記本文は変更しない
3. **既存アーキテクチャの活用**: Redis Pub/Subパターンを再利用
4. **非同期処理**: サーバーの即座のレスポンスにより、UXが向上
5. **自動無効化**: 日記更新時に古いハイライトを自動で無効化
6. **コスト管理**: ユーザ自身のAPIキーを使用

### デメリット

1. **複雑性の増加**: 非同期処理とポーリングによる実装の複雑化
2. **LLM精度依存**: ハイライト品質はLLMの判断に依存
3. **レスポンス遅延**: 処理完了までユーザは待つ必要がある
4. **コスト**: ユーザのAPIトークンを消費

### トレードオフ

- **処理方式**: 非同期処理を選択
  - 代替案: 同期処理(シンプルだが、レスポンス遅延が大きい)
- **無効化方法**: 自動無効化(日記更新時)
  - 代替案: 手動削除のみ(古いハイライトが残る可能性)
- **ハイライト数**: LLMに3〜5箇所を推奨
  - 代替案: ユーザが数を指定(UI複雑化)
- **完了通知**: ポーリング方式
  - 代替案: WebSocket(インフラ複雑化)

## 参考資料

- ADR 0003: LLM統合
- ADR 0004: Redis Pub/Sub実装
- Issue #268: https://github.com/project-mikan/umi.mikan/issues/268

## 実装チェックリスト

### データベース

- [ ] `diary_highlights` テーブル作成(マイグレーション)
- [ ] インデックス作成
- [ ] xoコード生成

### バックエンド

- [ ] `TriggerDiaryHighlight` RPC実装
- [ ] `GetDiaryHighlight` RPC実装
- [ ] `DeleteDiaryHighlight` RPC実装
- [ ] Subscriber: `DiaryHighlightHandler` 実装
- [ ] LLMプロンプト作成とテスト
- [ ] 自動無効化ロジック実装
- [ ] Prometheus メトリクス追加
- [ ] テスト作成(単体テスト、統合テスト)

### フロントエンド

- [ ] 「LLMハイライト」ボタンコンポーネント作成
- [ ] ハイライト表示ロジック実装
- [ ] ポーリング処理実装
- [ ] 「ハイライトを削除」ボタン実装
- [ ] i18n対応(ja.json, en.json)
- [ ] エラーハンドリング実装
- [ ] ローディング状態の表示

### インフラ・モニタリング

- [ ] Grafana ダッシュボード更新
- [ ] Prometheus スクレイプ設定確認

### ドキュメント

- [ ] CLAUDE.md 更新
- [ ] README更新(必要に応じて)
