# ADR 0010: 自己分析レポート機能

## ステータス
Proposed

## コンテキスト

### モチベーション

- ユーザは日記を書き続ける中で、自分の感情・行動・思考のパターンを客観的に把握したい
- 「最近どんなことに喜びを感じているか」「繰り返し悩んでいるテーマは何か」を可視化したい
- 週次・月次・年次など複数の期間で自分の変化を振り返りたい
- 既存の日記要約（1日・1ヶ月単位）より長期・俯瞰的な視点での分析が求められる

### 要件

- 既存のLLM機能と同様に、ユーザ自身のGemini APIトークンを使用
- 分析対象期間: 直近7日・直近30日・直近90日・カスタム期間
- 分析内容: 感情トレンド、頻出テーマ、行動パターン、成長・変化の観察
- オンデマンド生成（ユーザが「レポート生成」ボタンを押した時）と定期自動生成の両方をサポート
- 生成結果はキャッシュして再表示コストを削減
- ユーザが機能のON/OFFを選択できる（LLM設定と連動）

## 決定

### アーキテクチャ概要

```
オンデマンド実行
Frontend (ボタンクリック) → Backend gRPC → Redis Pub/Sub → Subscriber → Gemini API
                                ↓                                           ↓
                           200 Response                          self_analysis_reports テーブル
                                                                            ↓
                                                                   Frontend (ポーリング)

定期自動生成
Scheduler (毎週日曜4時) → Redis Pub/Sub → Subscriber → Gemini API
                                                             ↓
                                                 self_analysis_reports テーブル
```

### データフロー

#### オンデマンド生成

1. **Frontend**: 「自己分析レポートを生成」ボタンをクリック
   - `TriggerSelfAnalysisReport` RPCを呼び出し（期間を指定）
   - 即座に202レスポンスを受け取り、処理中表示を開始

2. **Backend**: gRPCリクエストを受信
   - 既存のキャッシュ（生成済みレポート）を確認し、有効であればそのまま返却
   - キャッシュ無効の場合、Redis Pub/Subに `self_analysis` メッセージを送信

3. **Subscriber**: メッセージを受信して処理を実行
   - 指定期間の日記エントリを全件取得
   - Gemini APIに分析リクエストを送信
   - 結果を `self_analysis_reports` テーブルに保存

4. **Frontend**: ポーリングで `GetSelfAnalysisReport` RPCを呼び出し
   - レポート取得後、セクション別に表示

#### 定期自動生成（週次）

- Schedulerが毎週日曜日の深夜4時に自動実行
- auto-summary有効なユーザ全員に対して直近7日分のレポートを生成
- 既存レポートがある場合はスキップ

### データベース設計

#### 新規テーブル: `self_analysis_reports`

```sql
CREATE TABLE self_analysis_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_type TEXT NOT NULL,   -- 'weekly' | 'monthly' | 'quarterly' | 'custom'
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    diary_count INTEGER NOT NULL DEFAULT 0, -- 分析対象の日記件数
    report JSONB NOT NULL,                  -- レポート本文（構造化JSON）
    model_version TEXT NOT NULL DEFAULT 'gemini-1.5-flash',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, period_type, period_start, period_end)
);

CREATE INDEX idx_self_analysis_reports_user_id ON self_analysis_reports(user_id);
CREATE INDEX idx_self_analysis_reports_period ON self_analysis_reports(user_id, period_start, period_end);
```

#### レポートデータ構造 (JSONB)

```json
{
  "summary": "この期間のあなたは仕事の達成感と家族との時間に多くの喜びを感じていました。",
  "emotional_trends": {
    "dominant_emotions": ["充実感", "疲労", "期待"],
    "emotion_timeline": [
      {"week": "2026-03-01", "tone": "positive", "score": 0.7},
      {"week": "2026-03-08", "tone": "neutral", "score": 0.5}
    ],
    "emotional_range": "high"
  },
  "recurring_themes": [
    {"theme": "仕事・プロジェクト", "frequency": 12, "sentiment": "mixed"},
    {"theme": "家族", "frequency": 8, "sentiment": "positive"},
    {"theme": "健康・体調", "frequency": 5, "sentiment": "neutral"}
  ],
  "behavioral_patterns": [
    "週末は屋外活動を多く記録している",
    "仕事量が増えると睡眠に関する記述が増える",
    "読書や学習に関する記述が増加傾向にある"
  ],
  "growth_observations": [
    "新しい技術への挑戦を前向きに記述することが増えた",
    "人間関係における感謝の表現が増えた"
  ],
  "recommendations": [
    "体調管理について定期的に記録を続けることで、より詳細な分析が可能になります"
  ]
}
```

### API設計

#### gRPC Service

```protobuf
enum AnalysisPeriodType {
  PERIOD_TYPE_UNSPECIFIED = 0;
  PERIOD_TYPE_WEEKLY = 1;
  PERIOD_TYPE_MONTHLY = 2;
  PERIOD_TYPE_QUARTERLY = 3;
  PERIOD_TYPE_CUSTOM = 4;
}

message TriggerSelfAnalysisReportRequest {
  AnalysisPeriodType period_type = 1;
  google.protobuf.Timestamp period_start = 2; // PERIOD_TYPE_CUSTOM 時のみ使用
  google.protobuf.Timestamp period_end = 3;   // PERIOD_TYPE_CUSTOM 時のみ使用
  bool force_regenerate = 4;                  // キャッシュを無視して再生成
}

message TriggerSelfAnalysisReportResponse {
  bool queued = 1;
  bool cache_hit = 2;   // キャッシュから即返却した場合 true
  string report_id = 3; // キャッシュヒット時はレポートIDを返却
}

message GetSelfAnalysisReportRequest {
  AnalysisPeriodType period_type = 1;
  google.protobuf.Timestamp period_start = 2;
  google.protobuf.Timestamp period_end = 3;
}

message RecurringTheme {
  string theme = 1;
  int32 frequency = 2;
  string sentiment = 3;
}

message SelfAnalysisReportResponse {
  string report_id = 1;
  string summary = 2;
  repeated string dominant_emotions = 3;
  repeated RecurringTheme recurring_themes = 4;
  repeated string behavioral_patterns = 5;
  repeated string growth_observations = 6;
  repeated string recommendations = 7;
  int32 diary_count = 8;
  google.protobuf.Timestamp period_start = 9;
  google.protobuf.Timestamp period_end = 10;
  google.protobuf.Timestamp generated_at = 11;
}

message ListSelfAnalysisReportsRequest {
  int32 limit = 1;
  int32 offset = 2;
}

message ListSelfAnalysisReportsResponse {
  repeated SelfAnalysisReportSummary reports = 1;
  int32 total = 2;
}

service DiaryService {
  // 既存のRPCは省略
  rpc TriggerSelfAnalysisReport(TriggerSelfAnalysisReportRequest) returns (TriggerSelfAnalysisReportResponse);
  rpc GetSelfAnalysisReport(GetSelfAnalysisReportRequest) returns (SelfAnalysisReportResponse);
  rpc ListSelfAnalysisReports(ListSelfAnalysisReportsRequest) returns (ListSelfAnalysisReportsResponse);
}
```

#### Redis Pub/Sub メッセージフォーマット

チャネル: `diary_events`

```json
{
  "type": "self_analysis",
  "user_id": "uuid",
  "period_type": "weekly",
  "period_start": "2026-03-01",
  "period_end": "2026-03-07"
}
```

### フロントエンド実装

#### レポートページ（新規ルート: `/report`）

- 期間選択タブ（直近7日 / 直近30日 / 直近90日 / カスタム）
- 「レポートを生成」ボタン
- 処理中スピナー（ポーリング）

#### レポート表示コンポーネント

- **サマリーセクション**: 期間全体の要約テキスト
- **感情トレンドセクション**: 支配的な感情のタグ、感情の振れ幅
- **頻出テーマセクション**: テーマ一覧とポジティブ/ネガティブの色分けバッジ
- **行動パターンセクション**: 箇条書きリスト
- **成長・変化セクション**: 前向きな変化の観察結果
- **過去レポート一覧**: 生成済みレポートの履歴

#### LLM設定との連動

- ユーザ設定画面に「自己分析レポートの自動生成」トグルを追加
- `user_llms` テーブルの既存フラグと連動

### バックエンド実装

#### Subscriber ハンドラ

新規ハンドラ: `SelfAnalysisHandler`

1. 指定期間の全日記エントリを取得（日付昇順）
2. 日記が3件未満の場合は分析をスキップ（データ不足）
3. ユーザのLLM設定（APIキー）を確認
4. 日記テキストを結合してGemini APIにプロンプト送信
5. レスポンスをパースしてJSONB形式で `self_analysis_reports` テーブルにUPSERT
6. メトリクスを記録

#### Scheduler ジョブ追加

新規ジョブ: `SelfAnalysisWeeklyJob`（`DailyScheduledJob`を継承）

- 実行タイミング: 毎週日曜日 4:00 JST
- 処理: auto-summary有効なユーザを取得し、直近7日分のメッセージをPub/Subに送信
- 既存レポートがある場合はスキップ

### LLMプロンプト設計

```
以下は{period_start}から{period_end}までの間に書かれた{diary_count}件の日記です。

【日記一覧】
{日付}: {タイトル}
{本文}
---
（繰り返し）

これらの日記を分析し、以下の項目についてJSON形式で回答してください。

1. summary: この期間全体を100〜200文字で要約
2. emotional_trends.dominant_emotions: 特に強く現れた感情トップ3（日本語）
3. emotional_trends.emotional_range: 感情の振れ幅 ("high" | "medium" | "low")
4. recurring_themes: 繰り返し登場するテーマ（テーマ名・登場頻度・全体的な感情傾向）
5. behavioral_patterns: 観察できる行動パターン（3〜5項目）
6. growth_observations: ポジティブな変化や成長の観察（2〜3項目）
7. recommendations: 日記をより充実させるための提案（1〜2項目）

出力はJSON形式のみとし、説明文は不要です。
```

### モニタリング

#### Prometheus メトリクス

```
self_analysis_processing_total{status="success|failure"}
self_analysis_processing_duration_seconds
self_analysis_diary_count_histogram
self_analysis_api_calls_total{provider="gemini"}
self_analysis_api_errors_total{provider="gemini",error_type="..."}
self_analysis_cache_hit_total
```

## 結果

### メリット

1. **自己理解の促進**: 日記を書くだけでなく、自分の傾向を客観的に把握できる
2. **長期的な視点**: 既存の日次・月次要約よりも俯瞰的な分析が可能
3. **既存アーキテクチャの活用**: Scheduler・Pub/Sub・Gemini APIの再利用
4. **キャッシュによるコスト削減**: 同一期間の再生成コストを最小化

### デメリット

1. **大量テキスト処理**: 長期間の日記を一度に送信すると、トークン上限に達する可能性がある
2. **分析の主観性**: LLMの解釈によって分析結果が変わり得る
3. **最低データ要件**: 日記件数が少ない場合は有用な分析ができない

### トレードオフ

- **処理方式**: 非同期（Pub/Sub）を選択
  - 代替案: 同期処理（数十件の日記処理で数秒〜数十秒かかるため非採用）
- **キャッシュ戦略**: DBキャッシュ（期間単位のUNIQUE制約）を選択
  - 代替案: Redisキャッシュ（揮発性のため不採用）
- **定期生成頻度**: 週次を選択
  - 代替案: 毎日（コスト増加）、月次（フィードバックの鮮度が落ちる）
- **トークン上限対策**: 長期間は日記の先頭200文字のみ使用する方針
  - 代替案: 複数リクエストに分割して集約（実装複雑化）

## 参考資料

- ADR 0003: LLM統合
- ADR 0004: Redis Pub/Sub実装
- ADR 0005: Schedulerシステム
- ADR 0007: 直近トレンド分析機能

## 実装チェックリスト

### データベース

- [ ] `self_analysis_reports` テーブル作成（マイグレーション）
- [ ] インデックス作成
- [ ] xoコード生成

### バックエンド

- [ ] `TriggerSelfAnalysisReport` RPC実装
- [ ] `GetSelfAnalysisReport` RPC実装
- [ ] `ListSelfAnalysisReports` RPC実装
- [ ] Subscriber: `SelfAnalysisHandler` 実装
- [ ] Scheduler: `SelfAnalysisWeeklyJob` 実装
- [ ] LLMプロンプト作成とテスト
- [ ] トークン上限対策（日記テキストの切り詰め）実装
- [ ] Prometheus メトリクス追加
- [ ] テスト作成

### フロントエンド

- [ ] `/report` ルート作成
- [ ] 期間選択コンポーネント実装
- [ ] レポート表示コンポーネント実装（各セクション）
- [ ] ポーリング処理実装
- [ ] 過去レポート一覧実装
- [ ] LLM設定画面に自動生成トグル追加
- [ ] i18n対応（ja.json, en.json）
- [ ] エラーハンドリング実装

### インフラ・モニタリング

- [ ] Grafana ダッシュボード更新
- [ ] Prometheus スクレイプ設定確認

### ドキュメント

- [ ] CLAUDE.md 更新
