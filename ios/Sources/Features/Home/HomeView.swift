import SwiftUI

/// ホーム画面 - 今日・昨日・一昨日の日記を表示・編集する
struct HomeView: View {
    @State private var viewModel: DiaryViewModel
    @State private var todayContent: String = ""
    @State private var yesterdayContent: String = ""
    @State private var dayBeforeYesterdayContent: String = ""

    /// 最後にストアから反映した内容（ユーザーの入力途中の上書きを防ぐ）
    @State private var lastAppliedToday = ""
    @State private var lastAppliedYesterday = ""
    @State private var lastAppliedDayBefore = ""

    /// ハーフモーダルで表示する日記の日付
    @State private var selectedItem: DiarySheetItem?

    /// フォーカス中のカード（キーボードツールバーの保存対象）
    @FocusState private var focusedCard: DiaryCardFocus?

    private let authViewModel: AuthViewModel
    private let syncManager: SyncManager
    private let launchState: AppLaunchState?

    // swiftlint:disable:next type_contents_order
    init(authViewModel: AuthViewModel, syncManager: SyncManager, launchState: AppLaunchState? = nil) {
        self.authViewModel = authViewModel
        self.syncManager = syncManager
        self.launchState = launchState
        _viewModel = State(initialValue: DiaryViewModel(authViewModel: authViewModel, syncManager: syncManager))
    }

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                syncStatusBanner
                if viewModel.isLoading {
                    loadingView
                } else {
                    diaryCards
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 16)
        }
        .task {
            // ローカルストアから即座に表示する。データの有無によらず即座にスプラッシュを終了する
            viewModel.loadLocal()
            applyLoadedContents()
            launchState?.isInitialLoading = false
            // サーバーから最新を取得して反映する
            await viewModel.refreshFromServer()
            applyLoadedContents(preservingEdits: true)
        }
        .onAppear {
            // タブ切替などで戻った時にローカルの最新を反映する（入力途中の内容は保持）
            viewModel.loadLocal()
            applyLoadedContents(preservingEdits: true)
        }
        .refreshable {
            await viewModel.refreshFromServer()
            applyLoadedContents(preservingEdits: true)
        }
        // 日記詳細をハーフモーダルで表示する（閉じたら編集内容をカードへ反映する）
        .sheet(
            item: $selectedItem,
            onDismiss: {
                viewModel.loadLocal()
                applyLoadedContents(preservingEdits: true)
            },
            content: { item in
                // 左右スワイプで今日・昨日・一昨日を行き来できるよう3日分を渡す
                let items = homeSheetItems
                DiaryDetailSheet(
                    items: items,
                    initialIndex: items.firstIndex { $0.id == item.id } ?? 0,
                    authViewModel: authViewModel,
                    syncManager: syncManager
                )
            }
        )
        .overlay(alignment: .bottom) {
            if let error = viewModel.errorMessage {
                ErrorBannerView(message: error) { viewModel.errorMessage = nil }
            }
        }
        .toolbar { keyboardToolbar }
    }

    /// キーボードの上に表示するツールバー（保存・キーボードを閉じる）
    @ToolbarContentBuilder private var keyboardToolbar: some ToolbarContent {
        ToolbarItemGroup(placement: .keyboard) {
            Spacer()
            Button {
                saveFocusedCard()
            } label: {
                Label("保存", systemImage: "square.and.arrow.down")
            }
            Button {
                focusedCard = nil
            } label: {
                Image(systemName: "keyboard.chevron.compact.down")
            }
        }
    }

    /// 左右スワイプ用に今日・昨日・一昨日の3日分をまとめたリスト
    private var homeSheetItems: [DiarySheetItem] {
        [
            DiarySheetItem(date: viewModel.today.date),
            DiarySheetItem(date: viewModel.yesterday.date),
            DiarySheetItem(date: viewModel.dayBeforeYesterday.date)
        ]
    }

    /// オフライン・同期待ちの状態表示バナー
    @ViewBuilder private var syncStatusBanner: some View {
        if !syncManager.isOnline {
            HStack(spacing: 6) {
                Image(systemName: "wifi.slash")
                Text(
                    syncManager.pendingCount > 0
                        ? "オフライン：端末に保存されます（未同期 \(syncManager.pendingCount)件）"
                        : "オフライン：端末に保存されます"
                )
                Spacer()
            }
            .font(.caption)
            .foregroundStyle(Color.twSecondary)
            .padding(10)
            .clipShape(RoundedRectangle(cornerRadius: 10))
            .glassEffect(.regular, in: .rect(cornerRadius: 10))
        } else if syncManager.pendingCount > 0 {
            HStack(spacing: 6) {
                Image(systemName: "arrow.triangle.2.circlepath")
                Text("同期待ち \(syncManager.pendingCount)件")
                Spacer()
            }
            .font(.caption)
            .foregroundStyle(Color.twBlue)
            .padding(10)
            .clipShape(RoundedRectangle(cornerRadius: 10))
            .glassEffect(.regular, in: .rect(cornerRadius: 10))
        }
    }

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

    private var diaryCards: some View {
        Group {
            todayCard
            yesterdayCard
            dayBeforeYesterdayCard
        }
    }

    private var todayCard: some View {
        DiaryCardView(
            title: "今日",
            date: viewModel.today.date,
            content: $todayContent,
            isSaving: viewModel.todaySaving,
            isSaved: viewModel.todaySaved,
            focusValue: .today,
            focusedCard: $focusedCard,
            onOpenDetail: { selectedItem = DiarySheetItem(date: viewModel.today.date) },
            onSave: {
                Task {
                    await viewModel.saveToday(content: todayContent)
                    lastAppliedToday = todayContent
                }
            }
        )
    }

    private var yesterdayCard: some View {
        DiaryCardView(
            title: "昨日",
            date: viewModel.yesterday.date,
            content: $yesterdayContent,
            isSaving: viewModel.yesterdaySaving,
            isSaved: viewModel.yesterdaySaved,
            focusValue: .yesterday,
            focusedCard: $focusedCard,
            onOpenDetail: { selectedItem = DiarySheetItem(date: viewModel.yesterday.date) },
            onSave: {
                Task {
                    await viewModel.saveYesterday(content: yesterdayContent)
                    lastAppliedYesterday = yesterdayContent
                }
            }
        )
    }

    private var dayBeforeYesterdayCard: some View {
        DiaryCardView(
            title: "一昨日",
            date: viewModel.dayBeforeYesterday.date,
            content: $dayBeforeYesterdayContent,
            isSaving: viewModel.dayBeforeYesterdaySaving,
            isSaved: viewModel.dayBeforeYesterdaySaved,
            focusValue: .dayBeforeYesterday,
            focusedCard: $focusedCard,
            onOpenDetail: { selectedItem = DiarySheetItem(date: viewModel.dayBeforeYesterday.date) },
            onSave: {
                Task {
                    await viewModel.saveDayBeforeYesterday(content: dayBeforeYesterdayContent)
                    lastAppliedDayBefore = dayBeforeYesterdayContent
                }
            }
        )
    }

    /// フォーカス中のカードの日記を保存する（キーボードツールバーの保存ボタン用）
    private func saveFocusedCard() {
        switch focusedCard {
        case .today:
            Task {
                await viewModel.saveToday(content: todayContent)
                lastAppliedToday = todayContent
            }

        case .yesterday:
            Task {
                await viewModel.saveYesterday(content: yesterdayContent)
                lastAppliedYesterday = yesterdayContent
            }

        case .dayBeforeYesterday:
            Task {
                await viewModel.saveDayBeforeYesterday(content: dayBeforeYesterdayContent)
                lastAppliedDayBefore = dayBeforeYesterdayContent
            }

        case nil:
            break
        }
    }

    /// ViewModelのエントリ内容をテキスト欄へ反映する。
    /// preservingEdits=true の場合、ユーザーが入力途中のテキストは上書きしない。
    private func applyLoadedContents(preservingEdits: Bool = false) {
        let newToday = viewModel.today.entry?.content ?? ""
        if !preservingEdits || todayContent == lastAppliedToday {
            todayContent = newToday
            lastAppliedToday = newToday
        }

        let newYesterday = viewModel.yesterday.entry?.content ?? ""
        if !preservingEdits || yesterdayContent == lastAppliedYesterday {
            yesterdayContent = newYesterday
            lastAppliedYesterday = newYesterday
        }

        let newDayBefore = viewModel.dayBeforeYesterday.entry?.content ?? ""
        if !preservingEdits || dayBeforeYesterdayContent == lastAppliedDayBefore {
            dayBeforeYesterdayContent = newDayBefore
            lastAppliedDayBefore = newDayBefore
        }
    }
}

#Preview {
    let authViewModel = AuthViewModel()
    HomeView(authViewModel: authViewModel, syncManager: SyncManager(authViewModel: authViewModel))
}
