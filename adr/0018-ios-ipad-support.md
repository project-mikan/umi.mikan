# ADR 0018: iOSアプリのiPad対応

## ステータス

Accepted（2026-09-25 Phase 1 実装済み。Phase 2 は未着手）

## コンテキスト

### モチベーション

- iPadでも日記を書きたい・読み返したいという要望がある。特に「月ごと」「検索」で過去の日記を読み返す用途は大画面と相性が良い
- 外付けキーボード（Magic Keyboard等）での長文入力はiPhoneより快適であり、書く体験の向上も見込める

### 現状調査

プロジェクト設定上は既にiPadをターゲットに含んでいるが、UIはiPhone前提のまま全幅に引き伸ばされている。

| 項目 | 現状 | 評価 |
|---|---|---|
| `TARGETED_DEVICE_FAMILY` | `1,2`（app / Widgets / Tests 全ターゲット） | 対応済み。iPadでネイティブ解像度で動作する |
| iPadの画面回転 | `UISupportedInterfaceOrientations_iPad` に4方向すべて指定 | 対応済み |
| AppIcon | `ipad` idiom（20/29/40/76/83.5pt）登録済み | 対応済み |
| ルートナビゲーション | `MainView` の `TabView`（5タブ、スタイル指定なし） | iPadではトップのタブバーになる。サイドバー化されていない |
| 各画面のレイアウト | `ScrollView` + `LazyVStack` + `.padding(16)` | 13インチ横向きで1行が非常に長くなり、読みにくい |
| 日記詳細 | `DiaryDetailSheet`（`.sheet` + `presentationDetents([.medium, .large])`） | iPadのregular幅ではフォームシート表示となりdetentが効かない。編集領域が狭い |
| 前後の日記への移動 | 横スワイプジェスチャーのみ | 外付けキーボード/トラックパッド利用時に操作手段がない |
| 保存・キーボードを閉じる | `ToolbarItemGroup(placement: .keyboard)` | 外付けキーボード接続時はソフトウェアキーボードが出ず、ツールバーも表示されない。フォーカス喪失時の自動保存は機能する |
| Live Activity | `ActivityAuthorizationInfo().areActivitiesEnabled` でガード済み | iPadOSはLive Activity非対応のため起動しないが、既存ガードで安全に無効化される。未同期件数はHomeの `syncStatusBanner` で確認可能 |
| オンデバイス要約（ADR 0015） | `SystemLanguageModel.default.availability` で判定 | M1以降・A17 Pro搭載iPadでは動作する。非対応機種は既存どおりプレビュー表示 |
| おもいで通知 | `UNUserNotificationCenter` | iPadでもそのまま動作する |
| マルチウィンドウ | `UIApplicationSceneManifest_Generation = YES`（複数シーン未宣言） | 単一ウィンドウ |
| CI / Makefile | `IOS_DESTINATION = iPhone 17` のみ | iPadのビルド・表示崩れを検知できない |

### 要件

- iPad（全サイズ・縦横・Split View / Slide Over / Stage Manager）で崩れずに使える
- iPhoneの既存UX（ハーフモーダル、スワイプ移動、キーボードツールバー等）は一切変えない
- 外付けキーボードで日記の保存・前後移動・検索ができる
- オフラインファースト・同期の仕組み（`LocalDiaryStore` / `SyncManager`）はそのまま使う。サーバ側の変更は無し

## 決定

### 1. レイアウト切り替えの判定基準は `horizontalSizeClass` とする

- `UIDevice.current.userInterfaceIdiom == .pad` ではなく、`@Environment(\.horizontalSizeClass)` で切り替える
- 理由: iPadのSplit View / Slide Over / Stage Managerの狭いウィンドウでは `compact` になり、iPhoneと同じUIが適切になる。idiom判定だと狭いウィンドウで崩れる
- 逆にiPhone Pro Max横向きの `regular` 判定にも自然に追従する

### 2. ルートナビゲーション: `TabView` に `.tabViewStyle(.sidebarAdaptable)` を付与する

| 案 | 内容 | 採否 |
|---|---|---|
| A. `.sidebarAdaptable` | 既存の `Tab` APIのまま、regular幅ではサイドバー、compact幅ではタブバーに自動切替 | **採用**。変更1行でiPhoneへの影響なし |
| B. `NavigationSplitView` でルートを作り直す | サイドバー＋コンテンツ＋詳細の3カラム | 不採用。タブ毎の `NavigationStack` 構造を全面改修する必要があり、iPhone側の回帰リスクが大きい |

### 3. 本文幅の上限（readable width）を共通modifierで設ける

- `Features/Common/ReadableWidth.swift` に `readableContentWidth()` を追加し、regular幅ではコンテンツを最大幅（初期値 720pt）で中央寄せする。compact幅では何もしない
- Home / Monthly / Search / Settings / Entities / EntityDetail / 日記詳細 / Login / Register に適用する
- サイズクラスで修飾子自体を付け外しすると、Split Viewのリサイズでビューが作り直されて入力中のフォーカスが外れるため、常に同じframeを付けて幅の値だけを変える
- 最大幅の決定ロジックは `horizontalSizeClass` を引数に取る純粋関数に切り出してユニットテストする

### 4. 日記詳細: Phase 1はシートのまま大きく表示、Phase 2で月ごと・検索を2カラム化

**Phase 1（最小対応）**

- `DiaryDetailSheet` に `.presentationSizing(.page)` を付与し、regular幅では大きなページシートで表示する（compact幅ではdetentsが従来どおり効く）
- スワイプ移動はそのまま残す

**Phase 2（大画面を活かす）**

- regular幅の「月ごと」「検索」は `NavigationSplitView`（2カラム）にし、左に一覧、右に日記詳細をインライン表示する
- 右カラムは `DiaryDetailSheet` から「前後移動＋ViewModel保持＋未保存保存」の中身を `DiaryPager`（仮称）として切り出し、シートとインラインの両方から使う（既存の自動保存・スワイプ処理を二重実装しない）
- 一覧で別の日を選んだ時は、切り替え前に未保存の変更を保存する（`DiaryDetailSheet.transition(to:)` と同じ扱い）
- Homeは「今日を大きく＋昨日・一昨日・おもいでを横に並べる」2カラムを検討するが、編集中のスクロール位置復元ロジックとの干渉を見て判断する（Phase 2のスコープ外にしても良い）

| 案 | 内容 | 採否 |
|---|---|---|
| A. シート継続（Phase 1） | `presentationSizing(.page)` のみ | 採用（短期） |
| B. `NavigationSplitView` 2カラム（Phase 2） | 一覧と詳細を同時表示 | 採用（中期）。iPadで最も価値が大きい |
| C. iPad専用Viewを別途作る | `HomeView_iPad` 等 | 不採用。ロジックが二重化し、ADR 0015等の機能追加のたびに両方を直すことになる |

### 5. 外付けキーボード対応: `keyboardShortcut` を追加する

| ショートカット | 動作 | 配置 |
|---|---|---|
| ⌘S | 編集中の日記を保存 | Home（フォーカス中カード）、日記詳細 |
| ⌘[ / ⌘] | 前 / 次の日記へ移動 | 日記詳細（スワイプと同じ `showPrevious` / `showNext` を呼ぶ） |
| ⌘← / ⌘→ | 前月 / 次月 | 月ごと |
| ⌘T | 今日（月ごと） | 月ごと |
| ⌘F | 検索フィールドにフォーカス | 検索 |

- ⌘キー長押しで表示されるショートカット一覧に載るよう、ボタンに `keyboardShortcut` を付ける形で実装する
- ⌘S / ⌘[ / ⌘] / ⌘F は **regular幅のときだけナビゲーションバーにボタンを出し**、そのボタンにショートカットを付ける。compact幅（iPhone）は保存ボタンを置かない既存UX（フォーカス喪失時の自動保存）を変えないため、ボタンもショートカットも出さない。⌘S のボタンは編集中（フォーカス中）のみ表示する
  - 見えないボタンにショートカットだけ付ける方法や、`FocusedValues` + `commands` でメニューに出す方法も考えたが、前者はiOSでの動作が不確実で、後者はシートを表示しているときにどの画面の値が使われるかが不明確なため採らなかった
- 月ごとの ⌘← / ⌘→ / ⌘T は既存の月送りボタンに付けるだけなので、サイズクラスに関係なく有効にする
- `.keyboard` 配置のツールバーはソフトウェアキーボード用として残す（iPhone・iPadのソフトウェアキーボード時は従来どおり）

### 6. マルチウィンドウ（複数シーン）は当面サポートしない

- `UIApplicationSupportsMultipleScenes` は `NO` のままとする
- 理由: `LocalDiaryStore` / `SyncManager` / `LiveActivityManager` / `DiarySummaryStore` がアプリ単位の単一インスタンス前提であり、2つのウィンドウで同じ日付を同時編集すると「後勝ち」で入力が消える。ロック・変更通知の設計が必要になり、iPad対応の本筋に比べてコストが大きい
- 将来必要になったら別ADRで扱う

### 7. iPadで使えない機能は既存ガードのまま静かに無効化する

- Live Activity: iPadOSは非対応。既存の `areActivitiesEnabled` ガードで何もしない。未同期表示はHomeのバナーで代替済みなので追加対応はしない
- オンデバイス要約: 既存の `isAvailable` 判定のまま（非対応機種は本文プレビュー）
- いずれもiPad専用の分岐コードは追加しない

### 8. ビルド・テスト・CI

- `Makefile` に `IOS_DESTINATION_IPAD = platform=iOS Simulator,name=iPad Pro 11-inch (M5)` と `ios-build-ipad` ターゲットを追加する（iOS 26.2 以降のどのシミュレータランタイムにも存在する機種を選ぶ）
- CIへのiPadビルド追加は**見送る**。シミュレータ向けバイナリはiPhone/iPadで同一で、iPad向けに変わるのはアセットカタログのコンパイル程度のため、macOSランナーの時間を使う割に検出できるものがほぼ無い。レイアウト崩れはビルドでは検出できないため手動QAで確認する
- UIの確認はiPad Pro 13-inch / iPad mini の2機種 × 縦横 × Split View 1/3幅 で手動QAする

## 実装計画

| Phase | 作業 | 主な変更ファイル | 完了条件 |
|---|---|---|---|
| 1-1 | サイドバー化 | `ContentView.swift` | iPadでサイドバー表示、iPhoneは従来のタブバー |
| 1-2 | 本文幅の上限 | `Features/Common/ReadableWidth.swift`（新規）、各View、`ios/Tests/ReadableWidthTests.swift`（新規） | 13インチ横向きで1行が最大720ptに収まる。テーブル駆動テストが通る |
| 1-3 | 詳細シートの大型化 | `DiaryDetailSheet.swift` | regular幅でページシート、compact幅は従来のハーフモーダル |
| 1-4 | キーボードショートカット | `HomeView.swift`、`DiaryDetailSheet.swift`、`DiaryDetailView.swift`、`MonthlyView.swift`、`SearchView.swift` | 5章の表のショートカットが動き、⌘長押しで一覧表示される |
| 1-5 | iPadビルドのMakefile追加 | `Makefile` | `make ios-build-ipad` が成功 |
| 1-6 | ドキュメント更新 | `CLAUDE.md`（iOS UX Features / iOS Development） | iPad対応の方針・コマンドが記載されている |
| 2-1 | 詳細ページャーの切り出し | `Features/Detail/DiaryPager.swift`（新規）、`DiaryDetailSheet.swift` | シートの挙動が変わらない（自動保存・スワイプ・未保存保存） |
| 2-2 | 月ごとの2カラム化 | `MonthlyView.swift` | regular幅で一覧＋詳細を同時表示、別日選択時に未保存分が保存される |
| 2-3 | 検索の2カラム化 | `SearchView.swift` | regular幅で結果＋詳細（キーワードハイライト付き）を同時表示 |
| 2-4 | Homeの2カラム化（任意） | `HomeView.swift` | スクロール位置復元・自動保存が回帰しないことを確認できた場合のみ実施 |

Phase 1は1PR、Phase 2は2-1〜2-3を1PR、2-4は別PRとする。

### QAチェックリスト

| 観点 | 確認内容 |
|---|---|
| サイズ | iPad Pro 13-inch / iPad mini、縦・横 |
| マルチタスク | Split View 1/3・1/2・2/3、Slide Over、Stage Managerでウィンドウをリサイズした際にcompact/regularが正しく切り替わる |
| 入力 | ソフトウェアキーボード（ツールバー表示）、外付けキーボード（ショートカット・フォーカス喪失時の自動保存） |
| オフライン | 機内モードで編集 → 復帰時に同期される（`SyncManager` の既存挙動） |
| 付加機能 | Live Activityが起動しないこと、非対応iPadでオンデバイス要約がプレビュー表示のままになること |
| iPhone回帰 | iPhone 17でハーフモーダル・スワイプ・キーボードツールバー・スクロール位置復元が従来どおり |

## 影響

| 項目 | 内容 |
|---|---|
| サーバ | 変更なし |
| iPhone UX | `horizontalSizeClass == .compact` では従来どおり。Pro Max横向き（regular）ではサイドバー・幅上限・ナビゲーションバーの保存/前後移動ボタンが出る |
| App Store | iPad用スクリーンショット（13インチ）の提出が必要になる |
| 保守 | 画面追加時は `readableContentWidth()` の適用と、regular幅での表示確認が必要になる（CLAUDE.mdに明記する） |

## 却下した代替案

| 案 | 却下理由 |
|---|---|
| `TARGETED_DEVICE_FAMILY = 1` に戻してiPhoneアプリ互換モードで動かす | iPad上で拡大表示になり、外付けキーボードや大画面の利点を活かせない |
| Mac Catalyst / Designed for iPad on Mac を同時対応 | 要望はiPadのみ。Phase 2完了後に別途検討する |
| iPad専用画面を別Viewとして作成 | ロジックの二重化。機能追加のたびに両方の修正が必要になる |
