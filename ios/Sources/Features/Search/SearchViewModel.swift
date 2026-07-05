import Connect
import Foundation
import SwiftUI

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
        }
    }

    /// キーワード検索結果のスニペットをハイライト付きで生成する。
    ///
    /// フロントエンドと同様に、最初のマッチ位置がスニペットの先頭30文字付近に
    /// 来るよう150文字の窓で切り出し、全キーワードの出現箇所をハイライトする。
    func highlightedSnippet(content: String, keywords: [String]) -> AttributedString {
        // 改行を空白に置換して1行化し、連続スペースを正規化する
        let normalized = content
            .replacingOccurrences(of: "\r\n", with: " ")
            .replacingOccurrences(of: "\r", with: " ")
            .replacingOccurrences(of: "\n", with: " ")
            .replacingOccurrences(of: "  +", with: " ", options: .regularExpression)
            .trimmingCharacters(in: .whitespaces)

        let window = 150
        let prefixLength = 30
        let activeKeywords = keywords
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }

        // 全キーワードから最初のマッチ位置を検索する（大文字小文字無視）
        var firstMatchIndex: String.Index?
        for kw in activeKeywords {
            if let range = normalized.range(of: kw, options: .caseInsensitive) {
                if firstMatchIndex == nil || range.lowerBound < firstMatchIndex! {
                    firstMatchIndex = range.lowerBound
                }
            }
        }

        // マッチ位置に応じてスニペットを切り出す
        var excerpt: String
        var prefix = ""
        var suffix = ""
        let matchOffset = firstMatchIndex.map { normalized.distance(from: normalized.startIndex, to: $0) } ?? -1

        if matchOffset == -1 || matchOffset < prefixLength {
            excerpt = String(normalized.prefix(window))
            if normalized.count > window { suffix = "..." }
        } else {
            let start = max(0, matchOffset - prefixLength)
            let startIndex = normalized.index(normalized.startIndex, offsetBy: start)
            let endOffset = min(normalized.count, start + window)
            let endIndex = normalized.index(normalized.startIndex, offsetBy: endOffset)
            excerpt = String(normalized[startIndex ..< endIndex])
            if start > 0 { prefix = "..." }
            if endOffset < normalized.count { suffix = "..." }
        }

        // スニペット内のキーワード出現箇所をハイライトする
        var attributed = AttributedString(prefix + excerpt + suffix)
        for kw in activeKeywords {
            var searchStart = attributed.startIndex
            while let range = attributed[searchStart...].range(of: kw, options: .caseInsensitive) {
                attributed[range].backgroundColor = .yellow.opacity(0.5)
                searchStart = range.upperBound
            }
        }
        return attributed
    }
}
