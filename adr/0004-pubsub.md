# ADR-0004: Redis Pub/Sub for Asynchronous Processing

## Status
Accepted

## Context
日記の要約生成やLLM処理において、重い処理を非同期で実行する必要がある。リアルタイム性よりも確実性を重視し、バックグラウンドでの処理を可能にしたい。

## Decision
Redis Pub/Subを使用して非同期処理システムを構築する。

### Architecture
```
Publisher (Scheduler) → Redis Pub/Sub → Subscriber
                           ↓
                     Channel: diary_events
```

### Components

#### 1. Redis Configuration
- **Version**: Redis 8 (Alpine)
- **Persistence**: AOF (Append Only File) 有効
- **Library**: rueidis (高性能Go Redis クライアント)
- **Channel**: `diary_events`

#### 2. Message Format
全メッセージは JSON 形式で以下の基本構造：

```json
{
  "type": "message_type",
  "user_id": "uuid",
  ...additional_fields
}
```

#### Message Types

##### Daily Summary
```json
{
  "type": "daily_summary",
  "user_id": "uuid",
  "date": "2024-01-15"
}
```

##### Monthly Summary
```json
{
  "type": "monthly_summary",
  "user_id": "uuid",
  "year": 2024,
  "month": 1
}
```

#### 3. Publisher (Scheduler)
- **役割**: 定期タスクの実行とメッセージ送信
- **実装**: `backend/cmd/scheduler`
- **送信方法**: `PUBLISH diary_events {json}`
- **エラーハンドリング**: 送信失敗時のログ出力

#### 4. Subscriber
- **役割**: メッセージ受信と処理実行
- **実装**: `backend/cmd/subscriber`
- **受信方法**: `SUBSCRIBE diary_events`
- **処理**: メッセージタイプに応じた分岐処理
- **エラーハンドリング**: 処理失敗時のログ出力（メッセージは破棄）

#### 5. Processing Flow
1. Scheduler が定期実行（5分間隔）
2. DB から処理対象を特定
3. 各対象について JSON メッセージを Redis に PUBLISH
4. Subscriber が SUBSCRIBE でメッセージ受信
5. メッセージタイプに応じて処理実行
6. 結果を DB に保存

#### 6. Infrastructure
- **Development**: Docker Compose での Redis サービス
- **Production**: 永続化ボリューム + ネットワーク分離
- **Port**: 6379 (標準)
- **Environment Variables**:
  - `REDIS_HOST`
  - `REDIS_PORT`

### Error Handling
- **Connection Failure**: アプリケーション起動時に接続確認
- **Publish Failure**: ログ出力して次の処理を継続
- **Processing Failure**: ログ出力してメッセージ破棄
- **Message Parse Failure**: 不正メッセージは無視

### Limitations
- **At-Most-Once Delivery**: Redis Pub/Sub は配信保証なし
- **No Persistence**: Subscriber ダウン時のメッセージ喪失
- **No Dead Letter Queue**: 失敗メッセージの再処理なし

## Consequences

### Positive
- 非同期処理によりgRPCサーバーの応答速度向上
- スケーラブルな処理（Subscriberを増やせる）
- Redisの信頼性
- 高性能な rueidis ライブラリ使用
- JSON による柔軟なメッセージ形式

### Negative
- システムの複雑化
- Redis依存
- メッセージ配信保証の欠如
- 分散システムのデバッグ困難

### Future Considerations
- Redis Streams への移行検討（配信保証が必要な場合）
- メッセージ監視・メトリクス追加
- バックプレッシャー制御

## Implementation Notes

### 機能概要
- 非同期で実行開始・処理できる仕組みを作る
- 外部依存でなくdocker composeで実行できるものとする
- 用途はLLMの要約生成など時間を要するものを、日記の保存などをトリガーとして実行するため

### 技術選択
- メッセージキューイングはRedis Pub/Subを用いる
- PublisherはScheduler (backend/cmd/scheduler) で実装
- SubscriberはBackend/cmd/subscriberに実装し、Goで処理を行う

### 動作フロー
1. Scheduler が定期的に要約生成対象を特定
2. Redis Pub/Sub でメッセージをキューイング
3. Subscriber がメッセージを受信して要約生成処理を実行
4. 生成結果をデータベースに保存