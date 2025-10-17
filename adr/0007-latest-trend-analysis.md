# ADR 0007: 直近トレンド分析機能

## コンテキスト

### モチベーション

- ユーザは数日前に何をしたか、どう感じていたかを忘れがちである
- 自分の最近の傾向(体調、気分、活動パターンなど)を客観的に把握したい
- 日記の要約とは異なり、より長期的な視点での分析が必要

### 要件

- 既存の日記要約機能と同様に、ユーザ自身のLLMトークンを使用
- ユーザが機能のON/OFFを選択できる
- 直近1週間(今日を除く)の日記を元に傾向を分析
- 自動生成とマニュアル実行(デバッグ用)の両方をサポート

## 決定

### アーキテクチャ概要

```
Scheduler (毎日4時) → Redis Pub/Sub → Subscriber → Gemini API → Redis (TTL: 2日)
                                                                       ↓
                                                            Frontend (ホーム画面)
```

### データフロー

1. **Scheduler**: 毎日4時にトレンド分析ジョブを実行
   - `auto_latest_trend_enabled = true` のユーザを取得
   - 今日を除く直近7日間の日記エントリを確認
   - Redis Pub/Subに `latest_trend` メッセージを送信

2. **Subscriber**: メッセージを受信して分析を実行
   - 該当期間の日記内容を取得
   - Gemini APIに分析リクエストを送信(回答は約300字)
   - 結果をRedis(TTL: 2日)に保存
   - メトリクスを記録(Prometheus経由でGrafanaで可視化)

3. **Frontend**: ホーム画面で表示
   - 「日記」セクションの下にトレンド分析結果を表示
   - Redisから最新の分析結果を取得

### データストレージ

#### PostgreSQL

新規テーブル: `user_llms` テーブルに以下のカラムを追加

```sql
ALTER TABLE user_llms ADD COLUMN auto_latest_trend_enabled BOOLEAN NOT NULL DEFAULT FALSE;
```

#### Redis

キー形式: `latest_trend:{user_id}`

```json
{
  "user_id": "uuid",
  "analysis": "トレンド分析の結果テキスト(最大300字程度)",
  "period_start": "2025-10-10T00:00:00Z",
  "period_end": "2025-10-16T23:59:59Z",
  "generated_at": "2025-10-17T04:00:00Z"
}
```

TTL: 2日間(172800秒)

### API設計

#### gRPC Service

```protobuf
// diary.proto

message GetLatestTrendRequest {
  // user_id は認証情報から取得
}

message GetLatestTrendResponse {
  string analysis = 1;
  google.protobuf.Timestamp period_start = 2;
  google.protobuf.Timestamp period_end = 3;
  google.protobuf.Timestamp generated_at = 4;
}

message TriggerLatestTrendRequest {
  // デバッグ用のマニュアル実行
}

message TriggerLatestTrendResponse {
  bool success = 1;
  string message = 2;
}

service DiaryService {
  // 既存のRPCは省略

  rpc GetLatestTrend(GetLatestTrendRequest) returns (GetLatestTrendResponse);
  rpc TriggerLatestTrend(TriggerLatestTrendRequest) returns (TriggerLatestTrendResponse);
}
```

#### Redis Pub/Sub メッセージフォーマット

チャネル: `diary_events`

```json
{
  "type": "latest_trend",
  "user_id": "uuid",
  "period_start": "2025-10-10T00:00:00Z",
  "period_end": "2025-10-16T23:59:59Z"
}
```

### フロントエンド実装

#### ホーム画面への表示

- 位置: 「日記」セクションの下
- 表示内容:
  - 分析期間(YYYY/MM/DD - YYYY/MM/DD)
  - 分析結果テキスト(最大300字)
  - 生成日時
- データ取得: ページロード時に `GetLatestTrend` RPCを呼び出し
- エラーハンドリング: データがない場合は非表示

#### デバッグページ(非production環境のみ)

- パス: `/debug/latest-trend` (または既存のデバッグページに統合)
- 機能:
  - 「トレンド分析を実行」ボタン
  - ボタン押下で `TriggerLatestTrend` RPCを呼び出し
  - 実行結果の表示(成功/失敗メッセージ)
- アクセス制限: `import.meta.env.PROD === false` でビルド時に制御

### バックエンド実装

#### Scheduler

新規ジョブ: `LatestTrendJob`

- 実行間隔: 毎日4時(cron式: `0 4 * * *`)
- 処理内容:
  1. `auto_latest_trend_enabled = true` のユーザを取得
  2. 各ユーザについて、今日を除く直近7日間の日記エントリ数を確認
  3. エントリが存在する場合、Redis Pub/Subにメッセージを送信
- 分散ロック: `latest_trend:{user_id}:{date}` キーを使用

環境変数(オプション):

```bash
SCHEDULER_LATEST_TREND_INTERVAL=24h  # デフォルト: 24h(毎日4時実行の場合はcronで制御)
```

#### Subscriber

新規ハンドラ: `LatestTrendHandler`

- メッセージタイプ: `latest_trend`
- 処理内容:
  1. 指定期間の日記エントリを取得
  2. ユーザのLLM設定(APIキー、プロバイダー)を取得
  3. Gemini APIに分析リクエストを送信
     - プロンプト例: "以下は過去1週間の日記です。最近の傾向(体調、気分、活動、考え方など)を300字程度で分析してください。"
  4. 結果をRedisに保存(TTL: 2日)
  5. メトリクスを記録
- エラーハンドリング:
  - API呼び出し失敗時はリトライ(最大3回)
  - 日記エントリがない場合はスキップ

#### DiaryService実装

新規RPC実装:

1. `GetLatestTrend`: Redisからトレンド分析結果を取得
2. `TriggerLatestTrend`: マニュアルでトレンド分析生成を実行(デバッグ用、非production環境のみ)

### モニタリング

#### Prometheus メトリクス

Schedulerメトリクス:

```
latest_trend_jobs_total{status="success|failure"}
latest_trend_jobs_duration_seconds
latest_trend_users_processed_total
```

Subscriberメトリクス:

```
latest_trend_processing_total{status="success|failure"}
latest_trend_processing_duration_seconds
latest_trend_api_calls_total{provider="gemini"}
latest_trend_api_errors_total{provider="gemini",error_type="..."}
```

#### Grafana ダッシュボード

既存の「Pub/Sub Monitoring」ダッシュボードに以下のパネルを追加:

- トレンド分析生成ジョブ実行数(成功/失敗)
- トレンド分析生成処理時間
- トレンド分析生成対象ユーザ数
- Gemini API呼び出し数・エラー数

### 設定方法

#### ユーザ設定

LLM設定画面に以下を追加:

- 「トレンド分析自動生成」チェックボックス
- 説明テキスト: "毎日4時に、直近1週間の日記を元に最近の傾向を自動分析します。"
- 前提条件: LLMプロバイダーとAPIキーが設定済み

### セキュリティ考慮事項

1. **認証**: すべてのRPCは認証必須(既存のミドルウェアを使用)
2. **APIキー保護**: ユーザのAPIキーは暗号化して保存
3. **レート制限**: Redis Pub/Subのメッセージ送信にレート制限を適用
4. **デバッグエンドポイント**: 非production環境でのみ有効化

## 結果

### メリット

1. **ユーザ体験の向上**: 自分の傾向を客観的に把握できる
2. **既存アーキテクチャの活用**: Scheduler/Subscriber/Redis Pub/Subパターンを再利用
3. **スケーラビリティ**: 非同期処理により多数のユーザにも対応可能
4. **コスト管理**: ユーザ自身のAPIキーを使用するため、運営コスト増加なし
5. **データ効率**: Redisの短期保存により、古いデータによるストレージ圧迫を回避

### デメリット

1. **Redis依存**: Redisダウン時はトレンド表示不可(PostgreSQLにフォールバックする選択肢もあり)
2. **リアルタイム性の制限**: 毎日4時実行のため、最新の日記は翌日反映
3. **LLM API依存**: ユーザのAPIキーが無効な場合、分析失敗

### トレードオフ

- **データ保存場所**: Redisを選択(短期データに適している)
  - 代替案: PostgreSQLに保存(永続性は高いが、古いデータの管理が必要)
- **実行タイミング**: 毎日4時に固定
  - 代替案: ユーザが任意のタイミングで実行(UI複雑化、コスト増加)
- **分析期間**: 直近7日間に固定
  - 代替案: ユーザが期間を選択可能(UI複雑化、実装コスト増加)

## 参考資料

- ADR 0004: Redis Pub/Sub実装
- ADR 0005: Scheduler システムアーキテクチャ
- ADR 0003: LLM統合
- Issue #243: https://github.com/project-mikan/umi.mikan/issues/243

## 実装チェックリスト

### バックエンド

- [ ] `user_llms` テーブルに `auto_latest_trend_enabled` カラムを追加(マイグレーション)
- [ ] Redis用のデータ構造体とヘルパー関数を実装
- [ ] `GetLatestTrend` RPC実装
- [ ] `TriggerLatestTrend` RPC実装(デバッグ用)
- [ ] Scheduler: `LatestTrendJob` 実装
- [ ] Subscriber: `LatestTrendHandler` 実装
- [ ] Prometheus メトリクス追加
- [ ] テスト作成(単体テスト、統合テスト)

### フロントエンド

- [ ] トレンド分析表示コンポーネント作成(Atomic Design)
- [ ] ホーム画面への統合
- [ ] LLM設定画面に「トレンド分析自動生成」チェックボックス追加
- [ ] デバッグページ作成(非production環境のみ)
- [ ] i18n対応(ja.json, en.json)
- [ ] エラーハンドリング実装

### インフラ・モニタリング

- [ ] Grafana ダッシュボード更新
- [ ] Prometheus スクレイプ設定確認
- [ ] Redis TTL設定確認

### ドキュメント

- [ ] CLAUDE.md 更新(新機能の説明、設定方法)
- [ ] README更新(必要に応じて)
