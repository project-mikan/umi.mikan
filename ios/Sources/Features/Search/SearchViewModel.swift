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

    init(authViewModel: AuthViewModel) {
        self.authViewModel = authViewModel
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
        }
    }
}
