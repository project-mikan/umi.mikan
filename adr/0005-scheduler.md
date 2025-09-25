# ADR-0005: Scheduler Implementation for Automated Summary Generation

## Status
Accepted

## Context
日記要約の自動生成機能において、定期的なタスク実行が必要となりました。具体的には以下の要件があります：

- 日次要約の自動生成（今日を除く）
- 月次要約の自動生成（今月を除く）
- スケーラブルなジョブ管理
- 非同期処理によるパフォーマンス確保

## Decision
汎用的なSchedulerシステムを実装し、Redis Pub/Subを利用した非同期ジョブキューシステムを採用します。

### Architecture
```
Scheduler → Redis Pub/Sub → Subscriber → Database
    ↓           ↓              ↓           ↓
  定期実行    キューイング    処理実行    結果保存
```

### Components

#### 1. Scheduler (`backend/cmd/scheduler`)
- **役割**: 定期タスクの実行とジョブのキューイング
- **実行間隔**: 5分毎
- **設計**: `ScheduledJob` インターフェースによる拡張可能な設計

```go
type ScheduledJob interface {
    Name() string
    Interval() time.Duration
    Execute(ctx context.Context, s *Scheduler) error
}
```

#### 2. Jobs

##### DailySummaryJob
- **対象**: `user_llms.auto_summary_daily = true` のユーザー
- **条件**: `diaries` テーブルに日記があり、`diary_summary_days` に要約がない日
- **除外**: 今日 (`d.date < CURRENT_DATE`)
- **メッセージ形式**:
```json
{
  "type": "daily_summary",
  "user_id": "uuid",
  "date": "2024-01-15"
}
```

##### MonthlySummaryJob
- **対象**: `user_llms.auto_summary_monthly = true` のユーザー
- **条件**: `diary_summary_days` に日次要約があり、`diary_summary_months` に月次要約がない月
- **除外**: 今月 (`年 < 今年 OR (年 = 今年 AND 月 < 今月)`)
- **メッセージ形式**:
```json
{
  "type": "monthly_summary",
  "user_id": "uuid",
  "year": 2024,
  "month": 1
}
```

#### 3. Subscriber (`backend/cmd/subscriber`)
- **役割**: キューメッセージの処理と要約生成
- **処理内容**:
  - 日次要約: 日記内容 → LLM → `diary_summary_days`
  - 月次要約: 日次要約群 → LLM → `diary_summary_months`
- **重複処理**: `ON CONFLICT` による回避

#### 4. Deployment
- **Development**: Air によるホットリロード対応
- **Production**: マルチステージビルドによる最適化
- **Isolation**: 各サービス独立した tmp ディレクトリ使用

### Database Schema Dependencies
- `user_llms`: 自動要約設定
- `diaries`: 日記データ
- `diary_summary_days`: 日次要約
- `diary_summary_months`: 月次要約

### Redis Channel
- **Channel**: `diary_events`
- **Message Format**: JSON
- **Persistence**: Redis AOF による永続化

## Consequences

### Positive
- **スケーラビリティ**: ジョブ単位での並行処理
- **拡張性**: 新しいジョブタイプを容易に追加可能
- **信頼性**: Redis による永続化とエラーハンドリング
- **パフォーマンス**: 非同期処理による応答性向上

### Negative
- **複雑性**: 3つのサービス間の依存関係
- **監視**: 分散システムの監視が必要
- **デバッグ**: 非同期処理のデバッグが困難

### Risks
- Redis障害時のジョブ喪失
- 大量ジョブ発生時のメモリ使用量
- LLM API制限によるジョブ失敗

## Implementation Notes
- 各サービスは独立したDockerコンテナとして実行
- 環境変数による設定管理
- ログによる実行状況の追跡
- TODO: LLM API実装（現在はモック）
- TODO: 本番環境での監視・アラート設定