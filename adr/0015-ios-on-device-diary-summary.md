# ADR 0015: iOS内蔵ローカルLLMによる月次日記個別要約

## ステータス

Proposed

## コンテキスト

### モチベーション

- 月ごとの日記画面（`MonthlyView`）は各日を `dayRow` で一覧表示するが、本文プレビューは `contentPreview` による単純な先頭100文字カットのみで、内容の要点が伝わらない
- 既存の月間まとめ（`monthlySummary`、ADR 0003）はGemini APIによる「月全体」の要約であり、個々の日の日記までは要約しない。1日ごとの要約を既存のサーバサイドLLM経路（Redis Pub/Sub + Subscriber + Gemini API、ADR 0008と同構成）で作ろうとすると、日数分のAPI呼び出しコストとユーザのAPIトークン消費が発生し、月表示のたびに回すには重い
- iOS 26.2 をデプロイターゲットとしており、Apple の **Foundation Models framework**（オンデバイスLLM、iOS 18.1+ で導入）が利用できる。オンデバイス推論はネットワーク不要・APIトークン不要・追加コストゼロで、日次要約のような軽量・高頻度な用途に向く
- 「日記本文をサーバ外に一切送らずに要約したい」というプライバシー動機とも一致する（オンデバイス処理はGemini APIにテキストを送信しない）

### 要件

- 月ごとの日記画面で、各日の日記本文をオンデバイスLLMで1〜2文程度に要約し、`dayRow` のプレビュー部分に表示する
- 要約は非同期に生成し、生成中は元の本文プレビュー（先頭100文字カット）を表示し続け、生成完了時に要約テキストへアニメーション付きで切り替える（「文字がおしゃれに切り替わる」というUX要望）
- サーバのDB・API・Pub/Subは一切使わない。既存のバックエンドアーキテクチャ（ADR 0003, 0008）とは独立した、iOSアプリ内で完結する機能とする
- 要約結果はローカルにキャッシュし、日記本文が変わらない限り再生成しない（オンデバイスとはいえ推論はCPU/GPU/ANEを使うため、無駄な再計算を避ける）
- Foundation Models が利用不可の端末・OSバージョン（非対応デバイス、Apple Intelligence 無効設定、モデル未ダウンロード等）では機能ごと非表示にし、既存の `contentPreview` 表示にフォールバックする

## 決定

### 全体像

```
MonthlyView 表示 / 月送り
  → MonthlyViewModel.fetch() で entryMap 確定（既存フロー）
  → 各 dayRow の表示時、DiarySummaryStore にオンデバイス要約をリクエスト（画面内で完結、ネットワークなし）
       ├─ キャッシュ命中（本文ハッシュ一致）→ 即座に要約を返す
       └─ キャッシュ未命中 → Foundation Models で非同期生成 → キャッシュ保存 → 完了通知
  → dayRow は「本文プレビュー」→「要約」への切り替えを contentTransition でアニメーション表示
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

- 役割: 日記本文 → オンデバイス要約のキャッシュ付き非同期生成を担当する、`MonthlyViewModel` から独立したシングルトンストア（`LocalDiaryStore` と同様の位置づけ）
- キャッシュキー: `"\(year)-\(month)-\(day)"` + 本文の内容ハッシュ（本文が変わったら再生成させるため、日付キーだけでなくハッシュも突き合わせる）
- キャッシュ永続化: `UserDefaults` または軽量JSONファイル（`LocalDiaryStore` の永続化方式に合わせる）。要約は失っても再生成できる派生データなので、`LocalDiaryStore` 本体ほど厳密な永続化は不要
- 生成処理: `LanguageModelSession` に1〜2文程度の要約を指示するプロンプトを渡し、`respond(to:)` で結果を取得する（ストリーミング表示までは行わず、完成した要約をまとめて受け取ってから切り替えアニメーションを開始する。ストリーミング途中経過の表示は本ADRのスコープ外）
- 同時実行制御: 月表示ではダウンロードキャッシュ次第で最大31日分のリクエストが並行しうるため、`AsyncStream` かセマフォ的な仕組みで同時生成数を制限する（例: 3並列まで）。Foundation Modelsのセッションは軽量とはいえ、31件を無制限に並列実行するとメモリ・レイテンシが悪化するため

```swift
@Observable
final class DiarySummaryStore {
    static let shared = DiarySummaryStore()

    /// 日付キー → 要約テキストのキャッシュ（生成中は nil のまま、完了後に値が入る）
    private(set) var summaries: [String: String] = [:]

    /// 指定日の要約をリクエストする（キャッシュ済みなら即座に反映、未済なら非同期生成をキューイングする）
    func requestSummary(key: String, content: String) { ... }
}
```

#### `MonthlyViewModel` への組み込み

- `dayRow` 表示時（`.onAppear` 相当、または `MonthlyView` が `entryMap` 確定後にまとめて要求）に `DiarySummaryStore.shared.requestSummary(key:content:)` を呼ぶ
- `MonthlyViewModel` 自体は `DiarySummaryStore` の判定・生成ロジックを持たず、Viewから直接 `DiarySummaryStore.shared` を参照する薄い依存とする（既存の `LocalDiaryStore` 参照パターンと同様）

#### `dayRow` のUI変更

- 対応端末では、本文プレビューの代わりに要約（取得できていれば）を表示する
- 要約未生成〜生成中は既存の `contentPreview`（先頭100文字カット）を表示し続ける。生成が完了した瞬間に要約テキストへ切り替える
- 切り替えアニメーション: SwiftUIの `contentTransition(.opacity)` + `withAnimation` を用いて、テキストがフェード切り替わるようにする（「おしゃれに切り替わる」要望への対応）。文字ごとのタイプライター演出のような凝った実装は本ADRのスコープ外とし、まずはフェードトランジションのみとする

```swift
Text(displayText(for: day))
    .contentTransition(.opacity)
    .animation(.easeInOut(duration: 0.4), value: displayText(for: day))
```

- `sparkles` アイコン（既存の「月間まとめ」カードで使用中）を要約テキストの先頭に小さく添え、AI生成であることを視覚的に示す（生成中はアイコンを表示しない、または控えめなプレースホルダーとする）

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
- **切り替え演出**: シンプルなフェードトランジション（`contentTransition(.opacity)`）を選択（タイプライター演出等の凝ったアニメーションは不採用）
  - 理由: まず最小実装で「おしゃれな切り替わり」の要件を満たし、演出の作り込みは反応を見てから拡張する
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

- [ ] `ios/Sources/Infrastructure/DiarySummaryStore.swift` 新規作成（キャッシュ付き非同期要約生成）
- [ ] `SystemLanguageModel.default.availability` による対応端末判定ロジック実装
- [ ] Foundation Models 向け要約プロンプト作成（1〜2文程度、日本語）
- [ ] 同時実行数の制限（3並列など）を実装
- [ ] `MonthlyView` の `dayRow` に要約表示 + `contentTransition(.opacity)` によるフェード切り替え実装
- [ ] 生成中/未生成時は既存の `contentPreview` にフォールバック
- [ ] 非対応端末では要約UIを完全に非表示にするフォールバック確認
- [ ] キャッシュの永続化（本文ハッシュによる無効化判定含む）
- [ ] 単体テスト: キャッシュキー生成、本文ハッシュ変化時の再生成判定（正常系/異常系を日本語で記述）
- [ ] `make ios-lint`（`--strict`）/ `make ios-build` / `make ios-test` の確認

### ドキュメント

- [ ] CLAUDE.md の iOS UX Features に本機能の概要を追記
- [ ] ADR 0013 の機能候補一覧に本ADRへの参照を追加（該当する場合）
