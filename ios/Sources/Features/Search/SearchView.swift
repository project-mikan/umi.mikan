import SwiftUI

/// 検索ページ - キーワード検索と意味的検索に対応する
struct SearchView: View {
    @State private var viewModel: SearchViewModel

    private let authViewModel: AuthViewModel
    private let syncManager: SyncManager

    // swiftlint:disable:next type_contents_order
    init(authViewModel: AuthViewModel, syncManager: SyncManager) {
        self.authViewModel = authViewModel
        self.syncManager = syncManager
        _viewModel = State(initialValue: SearchViewModel(authViewModel: authViewModel))
    }

    var body: some View {
        ScrollView {
            LazyVStack(alignment: .leading, spacing: 16) {
                modeToggle
                searchField
                if viewModel.isSearching {
                    loadingView
                } else {
                    resultsSection
                }
            }
            .padding(16)
        }
        .overlay(alignment: .bottom) {
            if let error = viewModel.errorMessage {
                errorBanner(message: error)
            }
        }
    }

    // MARK: - 検索フォーム

    private var modeToggle: some View {
        HStack(spacing: 8) {
            modeButton(title: "キーワード", mode: .keyword, tint: .blue)
            modeButton(title: "意味的検索", mode: .semantic, tint: .purple)
            Spacer()
        }
    }

    private var searchField: some View {
        HStack(spacing: 12) {
            TextField(
                viewModel.mode == .semantic ? "例: 楽しかった旅行の思い出" : "キーワードを入力",
                text: $viewModel.keyword
            )
            .textFieldStyle(.plain)
            .submitLabel(.search)
            .onSubmit {
                Task { await viewModel.search() }
            }
            .padding(12)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))

            Button {
                Task { await viewModel.search() }
            } label: {
                Image(systemName: "magnifyingglass")
                    .frame(width: 44, height: 44)
            }
            .buttonStyle(.glassProminent)
            .disabled(viewModel.isSearching || viewModel.keyword.trimmingCharacters(in: .whitespaces).isEmpty)
        }
    }

    // MARK: - 検索結果

    @ViewBuilder private var resultsSection: some View {
        if let results = viewModel.keywordResults {
            keywordResultsView(results)
        } else if let results = viewModel.semanticResults {
            semanticResultsView(results)
        } else if !viewModel.hasSearched {
            emptyStateView(message: "キーワードを入力して検索してください")
        }
    }

    private var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .controlSize(.large)
            Text("検索中...")
                .font(.subheadline)
                .foregroundStyle(Color.twSecondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 60)
    }

    // MARK: - コンポーネント生成

    private func modeButton(title: String, mode: SearchViewModel.Mode, tint: Color) -> some View {
        Button {
            viewModel.mode = mode
        } label: {
            Text(title)
                .font(.caption)
                .fontWeight(.medium)
                .padding(.horizontal, 14)
                .padding(.vertical, 8)
                .background(viewModel.mode == mode ? tint : Color.gray.opacity(0.15))
                .foregroundStyle(viewModel.mode == mode ? .white : Color.twBody)
                .clipShape(Capsule())
        }
        .buttonStyle(.plain)
    }

    @ViewBuilder
    private func keywordResultsView(_ results: Diary_SearchDiaryEntriesResponse) -> some View {
        Text("「\(results.searchedKeyword)」の検索結果: \(results.entries.count)件")
            .font(.subheadline)
            .foregroundStyle(Color.twSecondary)

        if !results.expandedKeywords.isEmpty {
            Text("関連キーワード: \(results.expandedKeywords.joined(separator: "、"))")
                .font(.caption)
                .foregroundStyle(Color.twSecondary)
        }

        if results.entries.isEmpty {
            emptyStateView(message: "検索結果が見つかりませんでした")
        } else {
            ForEach(results.entries, id: \.id) { entry in
                NavigationLink {
                    DiaryDetailView(date: entry.date, authViewModel: authViewModel, syncManager: syncManager)
                } label: {
                    keywordResultRow(entry: entry, results: results)
                }
                .buttonStyle(.plain)
            }
        }
    }

    private func keywordResultRow(entry: Diary_DiaryEntry, results: Diary_SearchDiaryEntriesResponse) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(formatDate(entry.date))
                .font(.headline)
                .foregroundStyle(Color.twBlue)

            Text(viewModel.highlightedSnippet(
                content: entry.content,
                keywords: [results.searchedKeyword] + results.expandedKeywords
            ))
            .font(.subheadline)
            .foregroundStyle(Color.twBody)
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
    }

    @ViewBuilder
    private func semanticResultsView(_ results: Diary_SearchDiaryEntriesSemanticResponse) -> some View {
        Text("意味的検索の結果: \(results.results.count)件")
            .font(.subheadline)
            .foregroundStyle(Color.twSecondary)

        if results.results.isEmpty {
            emptyStateView(message: "検索結果が見つかりませんでした")
        } else {
            ForEach(results.results, id: \.diaryID) { result in
                NavigationLink {
                    DiaryDetailView(date: result.date, authViewModel: authViewModel, syncManager: syncManager)
                } label: {
                    semanticResultRow(result: result)
                }
                .buttonStyle(.plain)
            }
        }
    }

    private func semanticResultRow(result: Diary_SemanticSearchResult) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(formatDate(result.date))
                    .font(.headline)
                    .foregroundStyle(.purple)
                Spacer()
                Text("類似度: \(Int((result.similarity * 100).rounded()))%")
                    .font(.caption2)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(.purple.opacity(0.15))
                    .foregroundStyle(.purple)
                    .clipShape(Capsule())
            }

            if !result.chunkSummary.isEmpty {
                Text(result.chunkSummary)
                    .font(.caption)
                    .fontWeight(.medium)
                    .foregroundStyle(.purple)
            }

            Text(result.snippet)
                .font(.subheadline)
                .foregroundStyle(Color.twBody)
                .lineLimit(5)
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
    }

    private func emptyStateView(message: String) -> some View {
        Text(message)
            .font(.subheadline)
            .foregroundStyle(Color.twSecondary)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 40)
    }

    /// Diary_YMD を "YYYY年M月D日" 形式の文字列に変換する
    private func formatDate(_ date: Diary_YMD) -> String {
        String(format: "%d年%d月%d日", date.year, date.month, date.day)
    }

    private func errorBanner(message: String) -> some View {
        HStack(spacing: 8) {
            Image(systemName: "exclamationmark.circle.fill")
            Text(message)
                .font(.subheadline)
            Spacer()
            Button {
                viewModel.errorMessage = nil
            } label: {
                Image(systemName: "xmark")
                    .font(.caption)
            }
        }
        .padding(16)
        .background(.red.opacity(0.15))
        .foregroundStyle(Color.twRed)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .glassEffect(.regular.tint(.red), in: .rect(cornerRadius: 12))
        .padding(16)
    }
}
