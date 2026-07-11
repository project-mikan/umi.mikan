# ADR 0014: 「n年前の今日」振り返り機能（On This Day）

## ステータス

Proposed

## コンテキスト

ADR 0013（改善アイデア集）の候補1で挙げた「n年前の今日」を、独立ADRとして詳細化する。

日記サービス最大の敵は「書かなくなること」であり、能動的な検索なしに過去が目に入る導線を作ることが継続率向上の鍵になる。過去の同月同日の日記を全年分まとめて見せる「On This Day」は振り返り系サービスの定番機能であり、LLMを必要とせず既存資産（`GetDiaryEntries` の応用、`DiaryDetailSheet`、ローカル通知基盤）だけで実現できるため実装コストが小さい。

### スコープ

今回は iOS アプリを主対象とする。表示先は次の2つ:

1. Home 画面の専用セクション（スクロール可能な「n年前の今日」カード群）
2. ローカル通知（朝の指定時刻に「n年前の今日の日記があります」と知らせる）

ホーム画面ウィジェット（WidgetKit のタイムライン更新型）は対象外とする（`ios/Widgets/` には現状 Live Activity 用の `PendingDiaryLiveActivity` のみが存在し、通常ウィジェットは未実装のため、別途着手する場合は本ADRの延長として扱う）。Web フロントエンドへの展開は本ADRのスコープ外とし、必要になった時点で別ADRとする。

## 決定

### 全体像

```
iOS Home 画面表示時 / プルリフレッシュ時
  → GetOnThisDayEntries RPC（当日の同月同日、全年分）
  → Backend: diaries テーブルを EXTRACT(MONTH/DAY) で横断検索
  → オンライン時のみ取得・表示（LocalDiaryStore にはキャッシュしない）

ローカル通知（設定でON/OFFできる、デフォルトOFF）
  → 毎日決まった時刻に UNUserNotificationCenter でスケジュール
  → 通知タップで Home の「n年前の今日」セクションへ遷移
```

### なぜオンライン専用（ローカルキャッシュなし）か

`LocalDiaryStore` は「自分が今日書いている／編集中の日記」をオフラインファーストで守るための仕組みで、`needsSync` を軸にした同期ロジックはすでに複雑である。「n年前の今日」は読み取り専用の付加的な導線であり、これをローカル永続化の対象に加えると、通常の日記データと「振り返り専用に複製したデータ」が混在し、同期・整合性の考慮点が増える。オフライン時は単にセクションを非表示（またはプレースホルダー表示）にすれば十分なので、素朴にオンライン時のみ取得する方針とする。

### バックエンド

#### 新規 RPC: `GetOnThisDayEntries`

`GetDiaryEntries`（単日取得）の応用として、月・日を指定して全年分を横断取得する。

```protobuf
message GetOnThisDayEntriesRequest {
  int32 month = 1; // 1-12
  int32 day = 2;   // 1-31
}

message OnThisDayEntry {
  string diary_id = 1;
  int32 year = 2;
  string content = 3;
  google.protobuf.Timestamp date = 4;
}

message GetOnThisDayEntriesResponse {
  repeated OnThisDayEntry entries = 1; // year 降順（直近年から）
}

service DiaryService {
  rpc GetOnThisDayEntries(GetOnThisDayEntriesRequest) returns (GetOnThisDayEntriesResponse);
}
```

- `month`/`day` はリクエストの当日（クライアントのローカル日付）をそのまま渡す。サーバ側で「今日」を計算しない（クライアントのタイムゾーンに委ねる。iOS は常に端末ローカル時刻を使う既存方針と一貫）。
- 2/29 は存在しない年をスキップする（`EXTRACT` ベースのクエリなら自然に処理される。うるう年以外は 2/29 の日記が存在しないので結果に含まれない）。
- 当年（今年）は「n年前」ではないため結果から除外する。

#### database 層

`backend/infrastructure/database/diary_export.go` と同じファイル、または新規ファイル `on_this_day.go` に追加する（クエリドメインが異なるため新規ファイル `on_this_day_queries.go` を推奨）。

```sql
SELECT id, user_id, content, date, created_at, updated_at
FROM diaries
WHERE user_id = $1
  AND EXTRACT(MONTH FROM date) = $2
  AND EXTRACT(DAY FROM date) = $3
  AND EXTRACT(YEAR FROM date) < $4  -- 当年を除外
ORDER BY date DESC
```

- `date` 列にインデックスがない場合、`EXTRACT` を使う関数インデックスの追加を検討する（`CREATE INDEX ON diaries (user_id, EXTRACT(MONTH FROM date), EXTRACT(DAY FROM date))` 相当）。ただし1ユーザーあたりの日記数は数千件規模と想定され、`user_id` で絞り込んだ後の全件スキャンでも実用上問題にならない可能性が高いため、まずはインデックスなしで実装し、実測してから要否を判断する。
- CLAUDE.md の DB アクセス指針に従い、SQL は必ず `backend/infrastructure/database/` に置き、`package database_test` でテストを追加する。

### iOS 実装

#### Home 画面セクション

- `ios/Sources/Features/Home/` に新規セクション（例: `OnThisDaySectionView.swift`）を追加し、`HomeView` に組み込む。
- 表示内容: 年ごとにカード化し、横スクロールまたは縦積みで「n年前」ラベル + 日記本文の冒頭（またはハイライト抜粋）を表示。
- タップで既存の `DiaryDetailSheet` を開く（新しい詳細画面は作らない。CLAUDE.md の「Half-modal detail」方針を踏襲）。`DiarySheetItem` の配列は「n年前の今日」セクション内の並び順（年降順）をそのまま `items` として渡し、スワイプで年をまたいで移動できるようにする。
- 該当日記が1件もない場合はセクション自体を非表示にする（空状態のUIは作らない）。
- オフライン時、または RPC が失敗した場合はセクションを非表示にする（エラー表示は行わない。付加的な導線のため通常の日記表示を邪魔しない）。

#### ローカル通知（設定でON/OFF）

- `ios/Sources/Features/Settings/SettingsView.swift` / `SettingsViewModel.swift` に「n年前の今日 通知」トグルを追加する。デフォルトは **OFF**（通知の許可ダイアログをいきなり出さない、ユーザーの明示的なオプトインを必須にするため）。
- トグルON時に `UNUserNotificationCenter` の通知許可をリクエストし、許可された場合のみ毎日決まった時刻（例: 8:00、時刻は固定でよく将来的にピッカー化してもよい）に繰り返しローカル通知をスケジュールする（`UNCalendarNotificationTrigger` の `repeats: true` で毎日8:00に発火、通知本文は固定文言とし当日時点で該当日記が存在するかはクライアント側で毎回判定できないため「n年前の今日の日記を振り返りましょう」という誘導文言にとどめ、実際に何件あるかはアプリを開いてから表示する）。
- トグルOFF時は `removePendingNotificationRequests` でスケジュール済み通知を取り消す。
- 通知の許可状態（システム側で拒否された等）とアプリ内トグルの状態がズレるケースがあるため、`SettingsView` 表示時に `UNUserNotificationCenter.getNotificationSettings` で実際の許可状態を確認し、トグルON・システム拒否の場合はトグルをOFF表示に補正するか、設定アプリへの誘導を表示する。
- 通知タップ時の遷移は `UNUserNotificationCenterDelegate` で受け取り、Home画面の「n年前の今日」セクションへスクロールする（最小実装としてはHome画面を開くだけでもよい。スクロール位置合わせは実装コスト次第で後回し可）。
- 設定値の永続化は既存の `UserDefaults` ベースの設定と同じ方式に揃える（Keychain化は不要。通知ON/OFFは機密情報ではない）。

### 通知トリガーの選択（Scheduler vs ローカル通知）

サーバサイドの Scheduler + プッシュ通知（APNs）ではなく、端末ローカルの `UNUserNotificationCenter` を選ぶ。

- **理由**: 「n年前の今日」はユーザー個人の端末ローカル日付に紐づく通知であり、サーバがユーザーのタイムゾーンや起床時刻を把握する必要がない。APNs 経由のプッシュ通知はサーバ側に通知トークン管理・送信基盤（新規インフラ）を必要とし、実装コストが跳ね上がる。
- ローカル通知は「n年前の今日に日記があるかどうか」を事前に知らずにスケジュールする（前述の通り誘導文言のみ）。「該当日記がある日だけ通知したい」という要求が今後出てきた場合は、Scheduler が日次で対象ユーザーを判定し APNs で送る方式へ拡張する必要があるが、これは本ADRのスコープ外とする。

## 結果

### メリット

1. **実装コストが小さい**: 新規テーブル不要、LLM不要、既存の `DiaryDetailSheet` をそのまま再利用できる
2. **毎日開く理由を作れる**: 検索なしで過去が目に入る導線を追加でき、継続率向上に直接効く
3. **既存アーキテクチャとの整合性**: オンライン専用・ローカルキャッシュなしとすることで `LocalDiaryStore` / `SyncManager` の複雑化を避けられる
4. **通知はオプトイン**: デフォルトOFF・設定でON/OFF切り替え可能にすることで、望まないプッシュ通知による離脱リスクを避けられる

### デメリット・リスク

1. **オフライン時は使えない**: 機内モードや圏外では「n年前の今日」セクションが非表示になる。オフラインファーストのアプリとしては一貫性に欠けるが、付加機能として許容する
2. **通知内容が事前に確定できない**: ローカル通知は「日記があるかどうか」を判定せずにスケジュールするため、該当日記がない日にも通知が飛ぶ。誘導文言に留めることで違和感を軽減するが、根本解決にはサーバ側の日次判定 + プッシュ通知が必要（将来拡張）
3. **`EXTRACT` クエリのパフォーマンス**: 日記数が非常に多いユーザーでは全件スキャンのコストが増す可能性がある。実測のうえでインデックス追加を検討する

### トレードオフ

- **表示範囲**: 当日の同月同日のみを選択（前後数日を含める代替案は不採用）
  - 理由: ADR 0013 の元案通りシンプルに保つ。前後数日を含めると「何年前の何日の日記か」が曖昧になり、UIの説明コストが増える
- **ローカルキャッシュ**: しない（する代替案は不採用）
  - 理由: `LocalDiaryStore` の同期ロジックを複雑化させないため
- **通知配信方式**: ローカル通知を選択（APNsプッシュは不採用）
  - 理由: サーバ側に新規インフラ（デバイストークン管理・APNs送信基盤）が必要になり実装コストが跳ね上がる。ユーザー個人の端末ローカル日付で完結する通知のため、ローカル通知で要件を満たせる
- **通知デフォルト値**: OFF（ONの代替案は不採用）
  - 理由: 通知許可ダイアログを起動直後にいきなり出すとユーザー体験を損なう。明示的なオプトインとする

## 参考資料

- ADR 0013: 改善アイデア集（本機能の初出）
- CLAUDE.md: iOS Offline Support / iOS UX Features（`DiaryDetailSheet`, `LocalDiaryStore`, `SyncManager` の既存方針）

## 実装チェックリスト

### バックエンド

- [ ] `proto/diary/diary.proto` に `GetOnThisDayEntries` RPC 追加
- [ ] `make grpc` で Go/TypeScript/Swift 生成物を更新（proto変更時は `make grpc-swift` の実行漏れに注意。CLAUDE.md 参照）
- [ ] `backend/infrastructure/database/on_this_day_queries.go` にクエリ関数実装
- [ ] `package database_test` でテスト追加
- [ ] Service層に `GetOnThisDayEntries` ハンドラ実装 + テスト
- [ ] 当年除外・2/29境界のテストケース追加（正常系/異常系を日本語で記述）

### iOS

- [ ] `ios/Sources/Features/Home/OnThisDaySectionView.swift` 新規作成
- [ ] `HomeView` にセクション組み込み、`DiaryDetailSheet` 連携（スワイプ対応）
- [ ] `SettingsView`/`SettingsViewModel` に通知ON/OFFトグル追加（デフォルトOFF）
- [ ] `UNUserNotificationCenter` の許可リクエスト・スケジュール・取り消しロジック実装
- [ ] システム側の通知許可状態とアプリ内トグルの不整合補正
- [ ] 通知タップ時のHome遷移（`UNUserNotificationCenterDelegate`）
- [ ] オフライン時・RPC失敗時にセクションを非表示にするフォールバック
- [ ] `make ios-lint`（`--strict`）/ `make ios-build` / `make ios-test` の確認

### ドキュメント

- [ ] CLAUDE.md の iOS UX Features に本機能の概要を追記
- [ ] ADR 0013 から本ADRへの参照リンクを追加
