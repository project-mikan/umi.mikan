import Foundation
import Testing
@testable import umi_mikan

/// DiarySummaryStoreのテスト
@MainActor
struct DiarySummaryStoreTests {
    /// contentHash のテーブル駆動テスト用ケース
    struct ContentHashCase {
        let name: String
        let lhs: String
        let rhs: String
        let expectSame: Bool
    }

    /// parseSummary のテーブル駆動テスト用ケース
    struct ParseSummaryCase {
        let name: String
        let raw: String
        let expected: String
    }

    /// テスト用に一時ファイルへ保存するストアを生成する
    private func makeStore() -> DiarySummaryStore {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("diary_summary_store_test_\(UUID().uuidString).json")
        return DiarySummaryStore(fileURL: url)
    }

    // MARK: - contentHash

    @Test(
        "contentHash: 同一本文は同じハッシュ、異なる本文は異なるハッシュになる",
        arguments: [
            ContentHashCase(
                name: "正常系: 同一の本文からは同じハッシュが得られる",
                lhs: "今日は海に行った",
                rhs: "今日は海に行った",
                expectSame: true
            ),
            ContentHashCase(
                name: "正常系: 内容が異なる本文からは異なるハッシュが得られる",
                lhs: "今日は海に行った",
                rhs: "今日は山に行った",
                expectSame: false
            ),
            ContentHashCase(
                name: "異常系: 空文字と非空文字を比較するとハッシュが一致しないので別内容として扱われる",
                lhs: "",
                rhs: "今日は海に行った",
                expectSame: false
            )
        ]
    )
    func contentHash(testCase: ContentHashCase) {
        let lhsHash = DiarySummaryStore.contentHash(testCase.lhs)
        let rhsHash = DiarySummaryStore.contentHash(testCase.rhs)
        #expect((lhsHash == rhsHash) == testCase.expectSame, Comment(rawValue: testCase.name))
    }

    // MARK: - parseSummary

    @Test(
        "parseSummary: モデル出力を1行の要約文にパースする",
        arguments: [
            ParseSummaryCase(
                name: "正常系: 1行の出力がそのまま要約文になる",
                raw: "海に行って友人と再会し天気も良かった",
                expected: "海に行って友人と再会し天気も良かった"
            ),
            ParseSummaryCase(
                name: "正常系: 先頭の箇条書き記号が取り除かれる",
                raw: "・海に行って友人と再会した",
                expected: "海に行って友人と再会した"
            ),
            ParseSummaryCase(
                name: "正常系: 複数行返ってきても先頭の非空行だけが採用される",
                raw: "海に行って友人と再会した\n天気が良い一日だった",
                expected: "海に行って友人と再会した"
            ),
            ParseSummaryCase(
                name: "正常系: 先頭が空行の場合は次の非空行が採用される",
                raw: "\n海に行って友人と再会した",
                expected: "海に行って友人と再会した"
            ),
            ParseSummaryCase(
                name: "異常系: 60文字を超える行があると60文字に切り詰められる",
                raw: String(repeating: "あ", count: 80),
                expected: String(repeating: "あ", count: 60)
            ),
            ParseSummaryCase(
                name: "異常系: 空文字を渡すと空文字になる",
                raw: "",
                expected: ""
            )
        ]
    )
    func parseSummary(testCase: ParseSummaryCase) {
        let result = DiarySummaryStore.parseSummary(testCase.raw)
        #expect(result == testCase.expected, Comment(rawValue: testCase.name))
    }

    // MARK: - requestSummary

    @Test("異常系: オンデバイスLLMが利用不可な環境（CI等）ではrequestSummaryを呼んでも要約が生成されない")
    func requestSummaryDoesNothingWhenUnavailable() {
        let store = makeStore()

        store.requestSummary(key: "2026-07-05", content: "今日は海に行った")

        // シミュレータ/CI環境では Foundation Models が利用不可のため isAvailable が false になり、
        // 要約リクエストは即座に無視される（summaries に反映されない）
        if !store.isAvailable {
            #expect(store.summaries["2026-07-05"] == nil)
            #expect(store.pendingKeys.isEmpty)
        }
    }

    @Test("正常系: 空文字の本文はisAvailableに関わらず要約リクエストされない")
    func requestSummaryIgnoresEmptyContent() {
        let store = makeStore()

        store.requestSummary(key: "2026-07-05", content: "")

        #expect(store.summaries["2026-07-05"] == nil)
        #expect(store.pendingKeys.isEmpty)
    }

    @Test("異常系: オンデバイスLLMが利用不可な環境ではrequestSummariesを複数件渡しても何も生成されない")
    func requestSummariesDoesNothingWhenUnavailable() {
        let store = makeStore()

        store.requestSummaries([
            DiarySummaryRequest(key: "2026-07-01", content: "1日目の日記"),
            DiarySummaryRequest(key: "2026-07-02", content: "2日目の日記")
        ])

        if !store.isAvailable {
            #expect(store.summaries.isEmpty)
            #expect(store.pendingKeys.isEmpty)
        }
    }

    // MARK: - consumeJustCompleted

    @Test("正常系: 生成完了していないキーに対してconsumeJustCompletedを呼んでも何も起きない")
    func consumeJustCompletedIsNoOpForUnknownKey() {
        let store = makeStore()

        store.consumeJustCompleted(key: "2026-07-05")

        #expect(store.justCompletedKeys.isEmpty)
    }
}
