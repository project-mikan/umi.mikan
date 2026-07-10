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

    /// parsePoints のテーブル駆動テスト用ケース
    struct ParsePointsCase {
        let name: String
        let raw: String
        let expected: [String]
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

    // MARK: - parsePoints

    @Test(
        "parsePoints: モデル出力を箇条書き配列にパースする",
        arguments: [
            ParsePointsCase(
                name: "正常系: 改行区切りの3行がそのまま3件の箇条書きになる",
                raw: "海に行った\n友人と再会\n天気が良い",
                expected: ["海に行った", "友人と再会", "天気が良い"]
            ),
            ParsePointsCase(
                name: "正常系: 先頭の箇条書き記号が取り除かれる",
                raw: "・海に行った\n- 友人と再会\n* 天気が良い",
                expected: ["海に行った", "友人と再会", "天気が良い"]
            ),
            ParsePointsCase(
                name: "正常系: 4件以上出力されても先頭3件だけ採用される",
                raw: "1つ目\n2つ目\n3つ目\n4つ目",
                expected: ["1つ目", "2つ目", "3つ目"]
            ),
            ParsePointsCase(
                name: "正常系: 空行は無視される",
                raw: "海に行った\n\n友人と再会",
                expected: ["海に行った", "友人と再会"]
            ),
            ParsePointsCase(
                name: "異常系: 16文字を超える行があると16文字に切り詰められる",
                raw: "とても長い箇条書きの一文はここで切り詰められる",
                expected: ["とても長い箇条書きの一文はここで"]
            ),
            ParsePointsCase(
                name: "異常系: 空文字を渡すと空配列になる",
                raw: "",
                expected: []
            )
        ]
    )
    func parsePoints(testCase: ParsePointsCase) {
        let result = DiarySummaryStore.parsePoints(testCase.raw)
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
