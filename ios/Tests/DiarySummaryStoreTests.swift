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
}
