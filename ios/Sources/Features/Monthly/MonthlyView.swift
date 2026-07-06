import SwiftUI

/// 月毎ページ - 月ナビゲーションと日ごとの日記リストを表示する
struct MonthlyView: View {
    @State private var viewModel: MonthlyViewModel
    /// ハーフモーダルで表示する日記の日付
    @State private var selectedItem: DiarySheetItem?

    private let authViewModel: AuthViewModel
    private let syncManager: SyncManager

    // swiftlint:disable:next type_contents_order
    init(authViewModel: AuthViewModel, syncManager: SyncManager) {
        self.authViewModel = authViewModel
        self.syncManager = syncManager
        _viewModel = State(initialValue: MonthlyViewModel(authViewModel: authViewModel))
    }

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 12) {
                monthNavigation
                if viewModel.isLoading {
                    loadingView
                } else {
                    if let summary = viewModel.monthlySummary {
                        summaryCard(summary)
                    }
                    dayList
                }
            }
            .padding(16)
        }
        .task {
            await viewModel.fetch()
        }
        .refreshable {
            await viewModel.fetch()
        }
        .overlay(alignment: .bottom) {
            if let error = viewModel.errorMessage {
                ErrorBannerView(message: error) { viewModel.errorMessage = nil }
            }
        }
        // 日記詳細をハーフモーダルで表示する（閉じたら編集内容を一覧へ反映する）
        .sheet(
            item: $selectedItem,
            onDismiss: { Task { await viewModel.fetch() } },
            content: { item in
                // 左右スワイプでその月の各日付を行き来できるよう1ヶ月分を渡す
                DiaryDetailSheet(
                    items: monthSheetItems,
                    initialIndex: monthSheetItems.firstIndex { $0.id == item.id } ?? 0,
                    authViewModel: authViewModel,
                    syncManager: syncManager
                )
            }
        )
    }

    /// 左右スワイプ用にその月の全日付をまとめたリスト
    private var monthSheetItems: [DiarySheetItem] {
        (1 ... viewModel.daysInMonth).map { DiarySheetItem(date: viewModel.ymd(day: $0)) }
    }

    // MARK: - 月ナビゲーション

    private var monthNavigation: some View {
        HStack(spacing: 16) {
            Button {
                Task { await viewModel.previousMonth() }
            } label: {
                Image(systemName: "chevron.left")
                    .frame(width: 36, height: 36)
            }
            .buttonStyle(.glass)

            Text(String(format: "%d年%d月", viewModel.year, viewModel.month))
                .font(.title3)
                .fontWeight(.semibold)
                .foregroundStyle(Color.twHeading)
                .frame(maxWidth: .infinity)

            Button {
                Task { await viewModel.goToToday() }
            } label: {
                Text("今日")
                    .font(.caption)
                    .frame(height: 36)
                    .padding(.horizontal, 4)
            }
            .buttonStyle(.glass)

            Button {
                Task { await viewModel.nextMonth() }
            } label: {
                Image(systemName: "chevron.right")
                    .frame(width: 36, height: 36)
            }
            .buttonStyle(.glass)
        }
    }

    // MARK: - 日リスト

    private var dayList: some View {
        ForEach(1 ... viewModel.daysInMonth, id: \.self) { day in
            Button {
                selectedItem = DiarySheetItem(date: viewModel.ymd(day: day))
            } label: {
                dayRow(day: day)
            }
            .buttonStyle(.plain)
        }
    }

    // MARK: - その他

    private var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .controlSize(.large)
            Text("読み込み中...")
                .font(.subheadline)
                .foregroundStyle(Color.twSecondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 60)
    }

    /// 月間まとめカード
    private func summaryCard(_ summary: Diary_MonthlySummary) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 6) {
                Image(systemName: "sparkles")
                    .font(.caption)
                Text("月間まとめ")
                    .font(.subheadline)
                    .fontWeight(.semibold)
            }
            .foregroundStyle(Color.twGreen)

            Text(summary.summary)
                .font(.subheadline)
                .foregroundStyle(Color.twBody)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular.tint(.green.opacity(0.2)), in: .rect(cornerRadius: 14))
    }

    private func dayRow(day: Int) -> some View {
        let entry = viewModel.entryMap[day]
        return VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("\(day)")
                    .font(.headline)
                    .foregroundStyle(Color.twHeading)
                Text(viewModel.weekdayName(day: day))
                    .font(.caption)
                    .foregroundStyle(Color.twSecondary)
                Spacer()
                if entry == nil {
                    Label("書く", systemImage: "plus")
                        .font(.caption)
                        .foregroundStyle(Color.twBlue)
                }
            }

            if let entry {
                Text(contentPreview(entry.content))
                    .font(.subheadline)
                    .foregroundStyle(Color.twBody)
                    .lineLimit(3)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(10)
                    .background(.blue.opacity(0.08))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
            } else {
                Text("日記がありません")
                    .font(.caption)
                    .italic()
                    .foregroundStyle(Color.twSecondary)
            }
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
    }

    /// 内容のプレビュー文字列（100文字まで）を返す
    private func contentPreview(_ content: String) -> String {
        if content.count > 100 {
            return String(content.prefix(100)) + "..."
        }
        return content
    }
}
