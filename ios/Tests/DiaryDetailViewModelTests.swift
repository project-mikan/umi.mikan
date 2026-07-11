import Foundation
import Testing
@testable import umi_mikan

/// DiaryDetailViewModelのテスト
@MainActor
struct DiaryDetailViewModelTests {
    /// テスト用に一時ファイルへ保存するストアを生成する
    private func makeStore() -> LocalDiaryStore {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("detail_viewmodel_test_\(UUID().uuidString).json")
        return LocalDiaryStore(fileURL: url)
    }

    /// テスト用の Diary_YMD を生成する
    private func ymd(_ year: Int, _ month: Int, _ day: Int) -> Diary_YMD {
        var date = Diary_YMD()
        date.year = UInt32(year)
        date.month = UInt32(month)
        date.day = UInt32(day)
        return date
    }

    /// テスト用のViewModelを生成する
    private func makeViewModel(store: LocalDiaryStore) -> DiaryDetailViewModel {
        let authViewModel = AuthViewModel()
        let syncManager = SyncManager(authViewModel: authViewModel, store: store)
        return DiaryDetailViewModel(
            date: ymd(2026, 7, 5),
            authViewModel: authViewModel,
            syncManager: syncManager,
            store: store
        )
    }

    @Test("正常系: 初期状態では未保存の変更がない")
    func noUnsavedChangesInitially() {
        let viewModel = makeViewModel(store: makeStore())

        #expect(viewModel.hasUnsavedChanges == false)
    }

    @Test("正常系: 本文を編集すると未保存の変更ありになる")
    func hasUnsavedChangesAfterEdit() {
        let viewModel = makeViewModel(store: makeStore())

        viewModel.content = "今日の日記"

        #expect(viewModel.hasUnsavedChanges == true)
    }

    @Test("正常系: 編集後に元の内容へ戻すと未保存の変更なしになる")
    func noUnsavedChangesAfterRevert() {
        let viewModel = makeViewModel(store: makeStore())

        viewModel.content = "一時的な編集"
        viewModel.content = ""

        #expect(viewModel.hasUnsavedChanges == false)
    }
}
