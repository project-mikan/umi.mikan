import SwiftUI

/// ホーム画面 - 今日・昨日・一昨日の日記を表示・編集する
struct HomeView: View {
    @State private var viewModel: DiaryViewModel
    @State private var onThisDayViewModel: OnThisDayViewModel
    @State private var todayContent: String = ""
    @State private var yesterdayContent: String = ""
    @State private var dayBeforeYesterdayContent: String = ""

    /// 最後にストアから反映した内容（ユーザーの入力途中の上書きを防ぐ）
    @State private var lastAppliedToday = ""
    @State private var lastAppliedYesterday = ""
    @State private var lastAppliedDayBefore = ""

    /// ハーフモーダルで表示する日記の日付
    @State private var selectedItem: DiarySheetItem?
    /// 「n年前の今日」セクションでタップされた項目（このセクション内でのスワイプ切り替えに使う）
    @State private var selectedOnThisDayItem: DiarySheetItem?

    /// フォーカス中のカード（キーボードツールバーの保存対象）
    @FocusState private var focusedCard: DiaryCardFocus?

    /// スクロール位置の制御用（キーボードを閉じた時に元の位置へ戻すために使う）
    @State private var scrollPosition = ScrollPosition()
    /// 現在のスクロールオフセット（キーボードを閉じる直前の位置の記録に使う）
    @State private var currentScrollOffset: CGFloat = 0
    /// スクロール復元 Task（前の Task をキャンセルするために保持する）
    @State private var scrollRestoreTask: Task<Void, Never>?

    /// アプリのフォアグラウンド状態（書きかけのLive Activity制御に使う）
    @Environment(\.scenePhase)
    private var scenePhase

    private let authViewModel: AuthViewModel
    private let syncManager: SyncManager
    private let launchState: AppLaunchState?

    // swiftlint:disable:next type_contents_order
    init(authViewModel: AuthViewModel, syncManager: SyncManager, launchState: AppLaunchState? = nil) {
        self.authViewModel = authViewModel
        self.syncManager = syncManager
        self.launchState = launchState
        _viewModel = State(initialValue: DiaryViewModel(authViewModel: authViewModel, syncManager: syncManager))
        _onThisDayViewModel = State(initialValue: OnThisDayViewModel(authViewModel: authViewModel))
    }

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                syncStatusBanner
                if viewModel.isLoading {
                    loadingView
                } else {
                    diaryCards
                    OnThisDaySectionView(items: onThisDayViewModel.items) { item in
                        selectedOnThisDayItem = DiarySheetItem(date: item.entry.date)
                    }
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
            // 「n年前の今日」はオンライン専用・付加的な導線のため、失敗しても通常表示は妨げない
            await onThisDayViewModel.load()
        }
        .onAppear {
            // タブ切替などで戻った時にローカルの最新を反映する（入力途中の内容は保持）
            viewModel.loadLocal()
            applyLoadedContents(preservingEdits: true)
        }
        .refreshable {
            await viewModel.refreshFromServer()
            applyLoadedContents(preservingEdits: true)
            await onThisDayViewModel.load()
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
        // 「n年前の今日」をタップした場合は、そのセクション内でスワイプできるようにする
        .sheet(item: $selectedOnThisDayItem) { item in
            let items = onThisDaySheetItems
            DiaryDetailSheet(
                items: items,
                initialIndex: items.firstIndex { $0.id == item.id } ?? 0,
                authViewModel: authViewModel,
                syncManager: syncManager
            )
        }
        .overlay(alignment: .bottom) {
            if let error = viewModel.errorMessage {
                ErrorBannerView(message: error) { viewModel.errorMessage = nil }
            }
        }
        .toolbar { keyboardToolbar }
        // スクロール位置を常時追跡し、キーボードを閉じた時の位置復元に備える
        .scrollPosition($scrollPosition)
        .onScrollGeometryChange(for: CGFloat.self) { geometry in
            geometry.contentOffset.y
        } action: { _, newValue in
            currentScrollOffset = newValue
        }
        // カードからカーソル（フォーカス）が外れたら、未保存の変更を自動保存する
        .onChange(of: focusedCard) { oldValue, newValue in
            if let oldValue, oldValue != newValue {
                autoSaveIfChanged(card: oldValue)
            }
            // キーボードが閉じられた場合は、閉じる直前のスクロール位置を復元する
            if oldValue != nil, newValue == nil {
                restoreScrollOffset(currentScrollOffset)
            }
        }
        // バックグラウンド移行時の書きかけ保存とLive Activity制御
        .onChange(of: scenePhase) { _, newPhase in
            handleScenePhase(newPhase)
        }
    }

    /// キーボードの上に表示するツールバー（キーボードを閉じる）
    /// 保存はキーボードが閉じた際（フォーカス喪失）に自動保存されるため、保存ボタンは表示しない
    @ToolbarContentBuilder private var keyboardToolbar: some ToolbarContent {
        ToolbarItemGroup(placement: .keyboard) {
            Spacer()
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

    /// 「n年前の今日」セクション内でのスワイプ用リスト（表示順=直近年から）
    private var onThisDaySheetItems: [DiarySheetItem] {
        onThisDayViewModel.items.map { DiarySheetItem(date: $0.entry.date) }
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

    @ViewBuilder private var diaryCards: some View {
        todayCard
        yesterdayCard
        dayBeforeYesterdayCard
    }

    private var todayCard: some View {
        DiaryCardView(
            title: "今日",
            date: viewModel.today.date,
            content: $todayContent,
            isSaved: viewModel.todaySaved,
            focusValue: .today,
            focusedCard: $focusedCard
        ) {
            selectedItem = DiarySheetItem(date: viewModel.today.date)
        }
    }

    private var yesterdayCard: some View {
        DiaryCardView(
            title: "昨日",
            date: viewModel.yesterday.date,
            content: $yesterdayContent,
            isSaved: viewModel.yesterdaySaved,
            focusValue: .yesterday,
            focusedCard: $focusedCard
        ) {
            selectedItem = DiarySheetItem(date: viewModel.yesterday.date)
        }
    }

    private var dayBeforeYesterdayCard: some View {
        DiaryCardView(
            title: "一昨日",
            date: viewModel.dayBeforeYesterday.date,
            content: $dayBeforeYesterdayContent,
            isSaved: viewModel.dayBeforeYesterdaySaved,
            focusValue: .dayBeforeYesterday,
            focusedCard: $focusedCard
        ) {
            selectedItem = DiarySheetItem(date: viewModel.dayBeforeYesterday.date)
        }
    }

    /// キーボードが閉じた後に指定オフセットへスクロール位置を戻す。
    /// 前回の Task をキャンセルして上書きし、連続開閉でも最後の位置のみ反映する。
    private func restoreScrollOffset(_ offset: CGFloat) {
        scrollRestoreTask?.cancel()
        scrollRestoreTask = Task {
            try? await Task.sleep(for: .milliseconds(300))
            guard !Task.isCancelled else { return }
            scrollPosition.scrollTo(point: CGPoint(x: 0, y: offset))
        }
    }

    /// フォーカス中のカードの日記を保存する（キーボードツールバーの保存ボタン用）
    private func saveFocusedCard() {
        guard let focusedCard else { return }
        save(card: focusedCard)
    }

    /// 指定カードに未保存の変更がある場合のみ保存する（フォーカスが外れた時の自動保存用）
    private func autoSaveIfChanged(card: DiaryCardFocus) {
        let hasChanges = switch card {
        case .today: todayContent != lastAppliedToday
        case .yesterday: yesterdayContent != lastAppliedYesterday
        case .dayBeforeYesterday: dayBeforeYesterdayContent != lastAppliedDayBefore
        }
        guard hasChanges else { return }
        save(card: card)
    }

    /// 指定カードの日記を保存する
    private func save(card: DiaryCardFocus) {
        switch card {
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

// MARK: - バックグラウンド移行時の書きかけ保存・Live Activity制御

extension HomeView {
    /// 書きかけ（保存していない編集途中の内容）があるかどうか
    private var hasDraftInProgress: Bool {
        todayContent != lastAppliedToday
            || yesterdayContent != lastAppliedYesterday
            || dayBeforeYesterdayContent != lastAppliedDayBefore
    }

    /// アプリのフォアグラウンド状態の変化に応じて、書きかけの保存とLive Activityを制御する
    func handleScenePhase(_ phase: ScenePhase) {
        switch phase {
        case .inactive:
            // Live Activityの開始はフォアグラウンド中しかできないため、
            // バックグラウンド移行直前の inactive の時点で開始する
            if hasDraftInProgress {
                LiveActivityManager.shared.setDraft(true)
            }

        case .background:
            // 書きかけを失わないようにローカルへ自動保存し、全カード保存後に書きかけフラグを解除する
            Task { await backgroundSaveAndClearDraft() }

        case .active:
            // フォアグラウンド復帰したら書きかけのLive Activityを終了する
            LiveActivityManager.shared.setDraft(false)

        @unknown default:
            break
        }
    }

    /// 全カードを await 可能な形で順番に保存し、完了後に書きかけフラグを解除する（バックグラウンド移行時用）
    private func backgroundSaveAndClearDraft() async {
        if todayContent != lastAppliedToday {
            await viewModel.saveToday(content: todayContent)
            lastAppliedToday = todayContent
        }
        if yesterdayContent != lastAppliedYesterday {
            await viewModel.saveYesterday(content: yesterdayContent)
            lastAppliedYesterday = yesterdayContent
        }
        if dayBeforeYesterdayContent != lastAppliedDayBefore {
            await viewModel.saveDayBeforeYesterday(content: dayBeforeYesterdayContent)
            lastAppliedDayBefore = dayBeforeYesterdayContent
        }
        LiveActivityManager.shared.setDraft(false)
    }
}

#Preview {
    let authViewModel = AuthViewModel()
    HomeView(authViewModel: authViewModel, syncManager: SyncManager(authViewModel: authViewModel))
}
