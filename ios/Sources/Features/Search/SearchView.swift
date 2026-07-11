import SwiftUI

/// 検索ページ - キーワード検索と意味的検索に対応する
struct SearchView: View {
    @State private var viewModel: SearchViewModel
    /// ハーフモーダルで表示する日記（ハイライトするキーワード付き）
    @State private var selectedItem: DiarySheetItem?
    /// 検索フィールドのフォーカス状態（検索実行時にキーボードを閉じるために使う）
    @FocusState private var isSearchFieldFocused: Bool

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
                ErrorBannerView(message: error) { viewModel.errorMessage = nil }
            }
        }
        // 検索完了時に成功の触覚フィードバックを鳴らす
        .sensoryFeedback(.success, trigger: viewModel.completedSearchCount) { old, new in new > old }
        // 日記詳細を検索キーワードのハイライト付きハーフモーダルで表示する
        .sheet(item: $selectedItem) { item in
            // 左右スワイプで前後の検索結果を行き来できるよう結果一覧を渡す
            let items = searchSheetItems ?? [item]
            DiaryDetailSheet(
                items: items,
                initialIndex: items.firstIndex { $0.id == item.id } ?? 0,
                authViewModel: authViewModel,
                syncManager: syncManager
            )
        }
    }

    /// 左右スワイプ用に現在の検索結果をシート項目へ変換したリスト（結果表示順）
    private var searchSheetItems: [DiarySheetItem]? {
        if let results = viewModel.keywordResults {
            let keywords = [results.searchedKeyword] + results.expandedKeywords
            return results.entries.map { DiarySheetItem(date: $0.date, highlightKeywords: keywords) }
        }
        if let results = viewModel.semanticResults {
            // 同じ日記の複数チャンクがヒットした場合に備えて日付を重複排除する（順序は維持）
            let query = viewModel.keyword.trimmingCharacters(in: .whitespacesAndNewlines)
            var seenKeys = Set<String>()
            return results.results
                .filter { seenKeys.insert(LocalDiaryEntry.dateKey($0.date)).inserted }
                .map { DiarySheetItem(date: $0.date, highlightKeywords: [query]) }
        }
        return nil
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
            .focused($isSearchFieldFocused)
            .submitLabel(.search)
            .onSubmit {
                performSearch()
            }
            .padding(12)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))

            Button {
                performSearch()
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

    /// キーボードを閉じてから検索を実行する（検索後に入力フォーカスが残らないようにする）
    private func performSearch() {
        isSearchFieldFocused = false
        Task { await viewModel.search() }
    }

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
                Button {
                    selectedItem = DiarySheetItem(
                        date: entry.date,
                        highlightKeywords: [results.searchedKeyword] + results.expandedKeywords
                    )
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

            Text(TextHighlighter.snippet(
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
                Button {
                    // 意味的検索は自然文クエリのため、本文に一致した場合のみハイライトされる
                    selectedItem = DiarySheetItem(
                        date: result.date,
                        highlightKeywords: [viewModel.keyword.trimmingCharacters(in: .whitespacesAndNewlines)]
                    )
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
}
