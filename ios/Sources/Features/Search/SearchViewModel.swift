import Connect
import Foundation

/// 検索ページのViewModel - キーワード検索と意味的検索を管理する
@MainActor
@Observable
final class SearchViewModel {
    /// 検索モード
    enum Mode {
        case keyword
        case semantic
    }

    var mode: Mode = .keyword
    var keyword: String = ""
    var isSearching: Bool = false
    var errorMessage: String?

    /// キーワード検索の結果
    var keywordResults: Diary_SearchDiaryEntriesResponse?
    /// 意味的検索の結果
    var semanticResults: Diary_SearchDiaryEntriesSemanticResponse?
    /// 検索実行済みかどうか（結果0件と未検索を区別する）
    var hasSearched: Bool = false
    /// 検索が完了した回数（完了時の触覚フィードバックのトリガーに使う）
    var completedSearchCount: Int = 0

    private let authViewModel: AuthViewModel
    private let store: LocalDiaryStore
    /// 意味的検索結果の本文プリフェッチ処理（再検索時にキャンセルする）
    private var prefetchTask: Task<Void, Never>?

    init(authViewModel: AuthViewModel, store: LocalDiaryStore = .shared) {
        self.authViewModel = authViewModel
        self.store = store
    }

    /// 現在のモードで検索を実行する
    func search() async {
        let query = keyword.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !query.isEmpty else { return }

        isSearching = true
        errorMessage = nil
        keywordResults = nil
        semanticResults = nil
        defer { isSearching = false }

        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)

        switch mode {
        case .keyword:
            var request = Diary_SearchDiaryEntriesRequest()
            request.keyword = query

            let response = await APIHelper.withTokenRefresh(authViewModel) {
                await client.searchDiaryEntries(request: request, headers: ConnectClient.shared.headers())
            }
            if let error = response.error {
                errorMessage = APIHelper.errorMessage(error)
                return
            }
            keywordResults = response.message
            hasSearched = true
            completedSearchCount += 1
            // 結果に本文が含まれるためローカルストアへ反映し、詳細表示を即座に開けるようにする
            for entry in response.message?.entries ?? [] where !entry.id.isEmpty {
                store.applyServerEntry(entry)
            }

        case .semantic:
            var request = Diary_SearchDiaryEntriesSemanticRequest()
            request.query = query
            request.limit = 10

            let response = await APIHelper.withTokenRefresh(authViewModel) {
                await client.searchDiaryEntriesSemantic(request: request, headers: ConnectClient.shared.headers())
            }
            if let error = response.error {
                errorMessage = APIHelper.errorMessage(error)
                return
            }
            semanticResults = response.message
            hasSearched = true
            completedSearchCount += 1
            // 結果はスニペットのみのため、裏で本文を一括プリフェッチする
            prefetchEntries(dates: response.message?.results.map(\.date) ?? [])
        }
    }

    /// 指定した日付の日記本文を裏で一括取得してローカルストアへ反映する。
    /// 詳細表示（ハーフモーダル）をサーバー待ちなしで開けるようにするためのプリフェッチ。
    private func prefetchEntries(dates: [Diary_YMD]) {
        prefetchTask?.cancel()

        // 同じ日記の複数チャンクがヒットした場合に備えて日付を重複排除する（順序は維持）
        var seenKeys = Set<String>()
        let uniqueDates = dates.filter { seenKeys.insert(LocalDiaryEntry.dateKey($0)).inserted }
        guard !uniqueDates.isEmpty else { return }

        prefetchTask = Task { [authViewModel, store] in
            let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
            var request = Diary_GetDiaryEntriesRequest()
            request.dates = uniqueDates

            let response = await APIHelper.withTokenRefresh(authViewModel) {
                await client.getDiaryEntries(request: request, headers: ConnectClient.shared.headers())
            }
            // プリフェッチは失敗しても検索機能に影響させない（詳細表示時に通常取得される）
            guard !Task.isCancelled, let message = response.message else { return }
            for entry in message.entries where !entry.id.isEmpty {
                store.applyServerEntry(entry)
            }
        }
    }
}
