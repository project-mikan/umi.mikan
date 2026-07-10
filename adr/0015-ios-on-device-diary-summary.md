# ADR 0015: iOS内蔵ローカルLLMによる月次日記個別要約

## ステータス

Accepted

## コンテキスト

### モチベーション

- 月ごとの日記画面（`MonthlyView`）は各日を `dayRow` で一覧表示するが、本文プレビューは `contentPreview` による単純な先頭100文字カットのみで、内容の要点が伝わらない
- 既存の月間まとめ（`monthlySummary`、ADR 0003）はGemini APIによる「月全体」の要約であり、個々の日の日記までは要約しない。1日ごとの要約を既存のサーバサイドLLM経路（Redis Pub/Sub + Subscriber + Gemini API、ADR 0008と同構成）で作ろうとすると、日数分のAPI呼び出しコストとユーザのAPIトークン消費が発生し、月表示のたびに回すには重い
- iOS 26.2 をデプロイターゲットとしており、Apple の **Foundation Models framework**（オンデバイスLLM、iOS 18.1+ で導入）が利用できる。オンデバイス推論はネットワーク不要・APIトークン不要・追加コストゼロで、日次要約のような軽量・高頻度な用途に向く
- 「日記本文をサーバ外に一切送らずに要約したい」というプライバシー動機とも一致する（オンデバイス処理はGemini APIにテキストを送信しない）

### 要件

- 月ごとの日記画面で、各日の日記本文をオンデバイスLLMで「重要なこと最大3点、各16文字程度の箇条書き」に要約し、`dayRow` のプレビュー部分に表示する
- 要約は月の1日目から順に（日付昇順で）生成を開始する。SwiftUIの行描画順に依存させず、`MonthlyViewModel` 側で明示的に日付昇順のリクエスト列を組み立てる
- 要約は非同期に生成し、生成中は元の本文プレビュー（先頭100文字カット）を表示し続け、生成完了の瞬間だけ虹色グラデーション＋ブラーを一瞬光らせてから要約表示へ切り替える（「AIらしく虹色のブラーをかけて、変わったことが見えるようにしてほしい」というUX要望）
- サーバのDB・API・Pub/Subは一切使わない。既存のバックエンドアーキテクチャ（ADR 0003, 0008）とは独立した、iOSアプリ内で完結する機能とする
- 要約結果はローカルにキャッシュし、日記本文が変わらない限り再生成しない（オンデバイスとはいえ推論はCPU/GPU/ANEを使うため、無駄な再計算を避ける）
- Foundation Models が利用不可の端末・OSバージョン（非対応デバイス、Apple Intelligence 無効設定、モデル未ダウンロード等）では要約が生成されず、既存の `contentPreview` 表示が継続する

## 決定

### 全体像

```
MonthlyView 表示 / 月送り
  → MonthlyViewModel.fetch() で entryMap 確定（既存フロー）
  → MonthlyViewModel.requestOnDeviceSummaries() が日付昇順（1日→末日）のリクエスト列を組み立てて
    DiarySummaryStore.requestSummaries(_:) へ一括発行（画面内で完結、ネットワークなし）
       ├─ キャッシュ命中（本文ハッシュ一致）→ 即座に要約（箇条書き配列）を返す
       └─ キャッシュ未命中 → 日付昇順でキューイングし、同時実行数3件の枠内で Foundation Models が順次生成
            → キャッシュ保存 → justCompletedKeys に完了マーク
  → dayRow は完了マークを検知した瞬間だけ虹色グラデーション＋ブラーを光らせ、
    「本文プレビュー」→「箇条書き要約」への切り替えを演出する
```

サーバサイドの関与はゼロ。既存の `GetDiaryEntriesByMonth` で取得済みの本文（`entryMap`）を入力にするだけなので、新規RPC・新規テーブル・Pub/Subメッセージタイプは不要。

### なぜオンデバイスか（Gemini API方式との比較）

| 観点 | オンデバイス（Foundation Models） | サーバ（Gemini API, ADR 0008方式） |
|---|---|---|
| コスト | 無料（ユーザのAPIトークン消費なし） | ユーザのAPIトークンを消費 |
| 通信 | 不要（オフラインでも動作） | 必須（Redis Pub/Sub経由） |
| 月表示のたびに全日再生成する運用 | 現実的（無料・低レイテンシ） | 非現実的（コスト・レイテンシが月表示のたびに発生） |
| プライバシー | 本文が端末外に出ない | 本文がGemini APIに送信される |
| 要約品質 | オンデバイスモデルのため簡易的 | Gemini本体を使うため高品質 |
| 対応端末 | Apple Intelligence対応機種・OSのみ | 全端末（Gemini API利用ユーザなら） |

月次個別要約は「1日ごとの軽い要約を毎回大量に生成する」という高頻度・低コスト志向の用途であり、Gemini API方式（ADR 0008の日記ハイライトのような重い非同期処理）よりオンデバイスの特性に合う。既存の「月間まとめ」（Gemini API、月1回生成でコストが小さい）とは住み分けが成立する。

### iOS 実装

#### Foundation Models 利用可否の判定

```swift
import FoundationModels

/// 端末がオンデバイスLLM要約に対応しているかを判定する
enum DiarySummaryAvailability {
    static var isSupported: Bool {
        switch SystemLanguageModel.default.availability {
        case .available:
            return true
        case .unavailable:
            return false
        }
    }
}
```

- `SystemLanguageModel.default.availability` が `.unavailable(.deviceNotEligible)` / `.unavailable(.appleIntelligenceNotEnabled)` / `.unavailable(.modelNotReady)` などの場合は非対応として扱い、機能全体を隠す（エラー表示はしない。ADR 0014の「付加機能はオフライン時に静かに非表示」という既存方針を踏襲）
- 判定は起動時・設定変更を検知するタイミング（`scenePhase == .active` 復帰時など）で再評価する

#### 新規コンポーネント: `DiarySummaryStore`

`ios/Sources/Infrastructure/DiarySummaryStore.swift`（新規）

- 役割: 日記本文 → オンデバイス要約（箇条書き配列）のキャッシュ付き非同期生成を担当する、`MonthlyViewModel` から独立したシングルトンストア（`LocalDiaryStore` と同様の位置づけ）
- キャッシュキー: `LocalDiaryEntry.dateKey`（`"YYYY-MM-DD"`）+ 本文の内容ハッシュ（本文が変わったら再生成させるため、日付キーだけでなくハッシュも突き合わせる）
- キャッシュ永続化: 軽量JSONファイル（`LocalDiaryStore` の永続化方式に合わせる）。要約は失っても再生成できる派生データなので、`LocalDiaryStore` 本体ほど厳密な永続化は不要
- 生成処理: `LanguageModelSession` に「重要なことを最大3つ、各12〜16文字程度の体言止めフレーズ、改行区切りで出力」を指示するプロンプトを渡し、`respond(to:)` で結果を取得後 `parsePoints` で箇条書き配列にパースする（先頭の記号除去、最大3件、各 `maxPointLength`（16）文字で切り詰め）。ストリーミング表示は行わず、完成した箇条書きをまとめて受け取ってから演出を開始する
- 同時実行制御: 月表示では最大31日分のリクエストが発生しうるため、内部の FIFO キューで同時生成数を3件に制限する。Foundation Modelsのセッションは軽量とはいえ、31件を無制限に並列実行するとメモリ・レイテンシが悪化するため
- 生成順序の保証: `requestSummaries(_:)` は渡された配列の順序どおりにキューへ積む。呼び出し側（`MonthlyViewModel`）が日付昇順の配列を渡すことで、「月の上（1日）から順に」生成が開始される（同時実行数3件の範囲内では並列実行される）
- 完了通知: `justCompletedKeys`（`Set<String>`）に完了した dateKey を追加し、View 側が虹色フラッシュ演出を再生した後 `consumeJustCompleted(key:)` で消費する

```swift
@Observable
final class DiarySummaryStore {
    static let shared = DiarySummaryStore()

    /// 日付キー → 要約箇条書き配列のキャッシュ（生成中はマップに含まれない）
    private(set) var summaries: [String: [String]] = [:]
    /// 直前に生成が完了した日付キーの集合（虹色フラッシュ演出のトリガーに使う）
    private(set) var justCompletedKeys: Set<String> = []

    /// 呼び出し順（= 渡した配列の順序）どおりにキューへ積み、日付昇順で処理を開始させる
    func requestSummaries(_ requests: [DiarySummaryRequest]) { ... }
    func requestSummary(key: String, content: String) { ... }
    func consumeJustCompleted(key: String) { ... }
}
```

#### `MonthlyViewModel` への組み込み

- `fetch()` が `entryMap` を確定させた直後（`loadLocalMonth()` 後）に `requestOnDeviceSummaries()` を1回呼ぶ。このメソッドが `entryMap.keys.sorted()`（日付昇順）でリクエスト配列を組み立て、`DiarySummaryStore.shared.requestSummaries(_:)` へまとめて渡す
- 個々の `dayRow`/`dayPreview` は要約リクエストを一切トリガーしない（以前は `.task(id:)` で行単位にリクエストしていたが、SwiftUIの描画順に依存し「上から順」を保証できないためやめた）
- `MonthlyViewModel` 自体は `DiarySummaryStore` の判定・生成ロジックを持たず、日付昇順のリクエスト列を組み立てる薄い依存とする

#### `dayRow` のUI変更

- 対応端末では、本文プレビューの代わりに要約の箇条書き（取得できていれば、最大3行）を表示する。各行は `sparkles` アイコン + 短いフレーズ（`lineLimit(1)` + `truncationMode(.tail)`）
- 要約未生成〜生成中は既存の `contentPreview`（先頭100文字カット）を表示し続ける
- 生成完了の瞬間: `justCompletedKeys` にキーが含まれる間、`dayRow` の上に虹色 `LinearGradient`（red〜purple）+ `.blur(radius: 12)` を `.opacity(0.55)` でオーバーレイし、「AIが生成した」ことを一瞬で示す。約0.5秒後に自身が `consumeJustCompleted` を呼びフェードアウトする
- 切り替えアニメーション: `contentTransition(.opacity)` + `withAnimation` でプレビュー→要約のフェードも維持する

```swift
Group {
    if let points, !points.isEmpty {
        VStack(alignment: .leading, spacing: 4) {
            ForEach(points, id: \.self) { point in
                HStack(spacing: 6) {
                    Image(systemName: "sparkles").foregroundStyle(Color.twGreen)
                    Text(point).lineLimit(1).truncationMode(.tail)
                }
            }
        }
    } else {
        Text(contentPreview(entry.content)).lineLimit(3)
    }
}
.contentTransition(.opacity)
.animation(.easeInOut(duration: 0.4), value: points)
.overlay {
    if justCompleted {
        LinearGradient(colors: [.red, .orange, .yellow, .green, .blue, .purple], startPoint: .leading, endPoint: .trailing)
            .opacity(0.55)
            .blur(radius: 12)
    }
}
```

### スコープ外（将来拡張）

- ストリーミング表示（生成途中のテキストを逐次表示する演出）
- 要約のタップでのオン/オフ切り替えや、要約か原文かをユーザが選べるトグルUI
- Web フロントエンドへの展開（Foundation Models はApple専用のため、Web版は既存のGemini API経路のままで別ADRとする）
- 月間まとめ（ADR 0003のGemini API版）の置き換え。あくまで個別日の補助的なプレビュー要約であり、月全体のまとめは既存のサーバサイド機能を維持する

## 結果

### メリット

1. **コストゼロ**: ユーザのAPIトークンを消費せず、月表示のたびに全日要約を再生成しても問題にならない
2. **プライバシー**: 日記本文が端末外（サーバ・Gemini API）に一切送信されない
3. **オフライン動作**: ネットワーク不要のため、オフラインファーストという既存のiOSアプリ方針とも一貫する
4. **既存アーキテクチャに影響しない**: 新規RPC・DBテーブル・Pub/Subメッセージタイプが不要で、バックエンドの変更ゼロで実装できる
5. **UX向上**: 一覧性の低い先頭100文字カットに代わり、要点が伝わる要約が表示され、月ごとの振り返り体験が向上する

### デメリット・リスク

1. **対応端末が限定される**: Apple Intelligence対応機種・OS・設定でのみ動作し、非対応環境では従来通りの100文字カット表示に留まる（機能の恩恵を受けられないユーザ層が一定数存在する）
2. **要約品質のばらつき**: オンデバイスモデルはサーバサイドのGemini本体と比べて要約精度が低い可能性がある。誤解を招く要約が生成されるリスクは既存のADR 0008（LLMハイライト）と同様に残る
3. **端末負荷**: 月送りのたびに最大31件のオンデバイス推論が走るため、古い（対応最小要件ギリギリの）端末では発熱・バッテリー消費・レイテンシが懸念される。同時実行数の制限（3並列など）で緩和するが、実測での調整が必要
4. **キャッシュ肥大化**: 日記が多いユーザは要約キャッシュも比例して増える。`LocalDiaryStore` 同様、上限や古いキャッシュの掃除方針は実装時に検討が必要

### トレードオフ

- **生成主体**: オンデバイス（Foundation Models）を選択（Gemini API方式は不採用）
  - 理由: 高頻度・低コスト志向の用途にオンデバイスの特性が合う。ADR 0008のような重い非同期処理はコスト・レイテンシの面で月表示には不向き
- **要約の形式**: 「重要なこと最大3点、各16文字程度の箇条書き」を選択（1〜2文の文章要約は不採用）
  - 理由: `dayRow` は3行分の表示枠があり、文章よりも箇条書きの方が短時間で要点を把握しやすいというフィードバックに基づく
- **生成順序**: `MonthlyViewModel` 側で日付昇順のリクエスト列を明示的に組み立てて一括発行（View の描画順に任せる方式は不採用）
  - 理由: SwiftUIの行描画順は保証されず、「月の上（1日）から順に」処理してほしいという要望を満たせないため
- **完了演出**: 虹色グラデーション＋ブラーの一瞬のフラッシュを選択（シンプルなフェードのみ、タイプライター演出は不採用）
  - 理由: 「AIらしく」「変わったことが見える」という要望に対し、フェードだけでは変化に気づきにくかったため、視覚的に強いシグナルを一瞬だけ入れる方式にした
- **非対応端末の扱い**: 機能を静かに非表示にしてフォールバック表示（エラー表示や代替オンライン生成は不採用）
  - 理由: ADR 0014と同じ方針。付加的な機能が非対応環境で目立ったエラーを出すと体験を損なう
- **要約対象**: 月ごとの日記画面の各日プレビューに限定（日記詳細画面や検索結果への展開は不採用、将来拡張として保留）
  - 理由: スコープを小さく保ち、まずは最も要約の恩恵が大きい一覧画面（`MonthlyView`）に絞る

## 参考資料

- ADR 0003: LLM統合（既存のサーバサイドLLM機能全体）
- ADR 0008: 日記エントリのLLMハイライト機能（Gemini API経由の非同期処理パターン）
- ADR 0013: 改善アイデア集
- ADR 0014: 「n年前の今日」振り返り機能（オンデバイス完結・オプトイン的UXの前例）
- CLAUDE.md: iOS UX Features（`MonthlyView`, `glassEffect` 等の既存UIパターン）
- Apple Developer Documentation: Foundation Models framework

## 実装チェックリスト

### iOS

- [x] `ios/Sources/Infrastructure/DiarySummaryStore.swift` 新規作成（キャッシュ付き非同期要約生成、箇条書きパース `parsePoints`、完了通知 `justCompletedKeys`）
- [x] `SystemLanguageModel.default.availability` による対応端末判定ロジック実装
- [x] Foundation Models 向け要約プロンプト作成（重要な点を最大3つ、各12〜16文字程度、日本語）
- [x] 同時実行数の制限（3並列）を維持しつつ、`requestSummaries(_:)` で呼び出し順（日付昇順）どおりにキューへ積むよう実装
- [x] `MonthlyViewModel.requestOnDeviceSummaries()` で `entryMap` を日付昇順に並べ替えて一括リクエスト発行
- [x] `MonthlyView` の `dayRow` に箇条書き要約表示（最大3行、各16文字目安、`lineLimit(1)`）を実装
- [x] 生成完了の瞬間に虹色グラデーション＋ブラーのフラッシュ演出（`summaryCompletionFlash`）を実装
- [x] 生成中/未生成時は既存の `contentPreview` にフォールバック
- [x] 非対応端末では要約が生成されず、既存の `contentPreview` 表示が継続されることを確認（要約UI自体は同じ `dayPreview` 内で条件分岐しており、非対応端末でも `sparkles` アイコン等は表示されない）
- [x] キャッシュの永続化（本文ハッシュによる無効化判定含む、キャッシュ値は箇条書き配列）
- [x] 単体テスト: `contentHash` の一致/不一致判定、`parsePoints` のパース・切り詰めロジック、`isAvailable == false` 環境での `requestSummary`/`requestSummaries` no-op 確認、`consumeJustCompleted` の安全性（正常系/異常系を日本語で記述、`ios/Tests/DiarySummaryStoreTests.swift`）
- [x] `make ios-lint`（`--strict`）/ `make ios-build` / `make ios-test` の確認

### ドキュメント

- [x] CLAUDE.md の iOS UX Features に本機能の概要を追記
- [ ] ADR 0013 の機能候補一覧に本ADRへの参照を追加（該当する場合）
