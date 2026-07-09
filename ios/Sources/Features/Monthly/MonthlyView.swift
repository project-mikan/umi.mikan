import SwiftUI

/// 月毎ページ - 月ナビゲーションと日ごとの日記リストを表示する
struct MonthlyView: View {
    @State private var viewModel: MonthlyViewModel
    /// ハーフモーダルで表示する日記の日付
    @State private var selectedItem: DiarySheetItem?
    /// 年月選択シートの表示状態
    @State private var isMonthPickerPresented = false
    /// 年月選択シートで選択中の年
    @State private var pickerYear = 2000
    /// 年月選択シートで選択中の月
    @State private var pickerMonth = 1

    private let authViewModel: AuthViewModel
    private let syncManager: SyncManager
    /// オンデバイスLLMによる日記要約ストア（非対応端末では isAvailable が false になり機能全体が非表示になる）
    private let summaryStore = DiarySummaryStore.shared

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
        // 年月選択シート（任意の年・月へジャンプする）
        .sheet(isPresented: $isMonthPickerPresented) {
            monthPickerSheet
        }
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

            yearMonthButton

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

    /// 年月ラベルのボタン。タップすると任意の年・月を選べるピッカーを開く
    private var yearMonthButton: some View {
        Button {
            pickerYear = viewModel.year
            pickerMonth = viewModel.month
            isMonthPickerPresented = true
        } label: {
            HStack(spacing: 4) {
                Text(String(format: "%d年%d月", viewModel.year, viewModel.month))
                    .font(.title3)
                    .fontWeight(.semibold)
                    .foregroundStyle(Color.twHeading)
                Image(systemName: "chevron.up.chevron.down")
                    .font(.caption2)
                    .foregroundStyle(Color.twSecondary)
            }
            .frame(maxWidth: .infinity)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    /// 年月選択シート。ホイールで年・月を選んで任意の月へ移動できる（数年前へ遡る時のショートカット）
    private var monthPickerSheet: some View {
        VStack(spacing: 16) {
            Text("表示する年月を選択")
                .font(.headline)
                .foregroundStyle(Color.twHeading)
                .padding(.top, 24)

            monthPickerWheels

            Button {
                isMonthPickerPresented = false
                Task { await viewModel.goTo(year: pickerYear, month: pickerMonth) }
            } label: {
                Text("この年月へ移動")
                    .fontWeight(.semibold)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 6)
            }
            .buttonStyle(.glassProminent)
            .padding(.horizontal, 16)

            Spacer(minLength: 0)
        }
        .presentationDetents([.height(380)])
        .presentationDragIndicator(.visible)
    }

    /// 年・月のホイールピッカー
    private var monthPickerWheels: some View {
        HStack(spacing: 0) {
            Picker("年", selection: $pickerYear) {
                ForEach(selectableYears, id: \.self) { year in
                    Text(String(format: "%d年", year)).tag(year)
                }
            }
            .pickerStyle(.wheel)

            Picker("月", selection: $pickerMonth) {
                ForEach(1 ... 12, id: \.self) { month in
                    Text("\(month)月").tag(month)
                }
            }
            .pickerStyle(.wheel)
        }
        .padding(.horizontal, 16)
    }

    /// 年ピッカーで選択できる年の範囲（1980年〜現在の年）。
    /// デバイスの暦設定が和暦などに変わっても正しいグレゴリオ年を使うため Calendar.gregorian を明示する。
    /// また computed var にすると Picker スクロールのたびに再計算されるため let で一度だけ生成する。
    private let selectableYears: [Int] = {
        let currentYear = Calendar(identifier: .gregorian).component(.year, from: Date())
        return Array(1980 ... max(currentYear, 1980))
    }()

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
                dayPreview(day: day, entry: entry)
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

    /// 日記本文のプレビュー表示。
    /// オンデバイス要約が利用可能な端末では生成をリクエストし、完了したらフェードで要約へ切り替える。
    /// 生成中・非対応端末では従来通りの100文字カットプレビューを表示する。
    private func dayPreview(day: Int, entry: Diary_DiaryEntry) -> some View {
        let key = LocalDiaryEntry.dateKey(entry.date)
        let summary = summaryStore.summaries[key]
        return Group {
            if let summary {
                HStack(alignment: .top, spacing: 6) {
                    Image(systemName: "sparkles")
                        .font(.caption2)
                        .foregroundStyle(Color.twGreen)
                    Text(summary)
                        .font(.subheadline)
                        .foregroundStyle(Color.twBody)
                        .lineLimit(3)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
            } else {
                Text(contentPreview(entry.content))
                    .font(.subheadline)
                    .foregroundStyle(Color.twBody)
                    .lineLimit(3)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
        .contentTransition(.opacity)
        .animation(.easeInOut(duration: 0.4), value: summary)
        .padding(10)
        .background(.blue.opacity(0.08))
        .clipShape(RoundedRectangle(cornerRadius: 10))
        .task(id: key) {
            summaryStore.requestSummary(key: key, content: entry.content)
        }
    }

    /// 内容のプレビュー文字列（100文字まで）を返す
    private func contentPreview(_ content: String) -> String {
        if content.count > 100 {
            return String(content.prefix(100)) + "..."
        }
        return content
    }
}
