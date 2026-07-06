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

    /// スクロール位置の制御用（キーボードを閉じた時に元の位置へ戻すために使う）
    @State private var scrollPosition = ScrollPosition()
    /// 現在のスクロールオフセット（キーボードを閉じる直前の位置の記録に使う）
    @State private var currentScrollOffset: CGFloat = 0

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
    }

    /// キーボードの上に表示するツールバー（保存・キーボードを閉じる）
    @ToolbarContentBuilder private var keyboardToolbar: some ToolbarContent {
        ToolbarItemGroup(placement: .keyboard) {
            Spacer()
            // 閉じるボタンと見分けやすいよう、チェックマーク＋青の塗りつぶしボタンにする
            Button {
                saveFocusedCard()
            } label: {
                Label(
                    isFocusedCardSaved ? "保存済み" : "保存",
                    systemImage: isFocusedCardSaved ? "checkmark" : "checkmark.circle.fill"
                )
                .labelStyle(.titleAndIcon)
                .fontWeight(.semibold)
            }
            .buttonStyle(.borderedProminent)
            .tint(Color.twBlue)
            .controlSize(.small)
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

    /// フォーカス中のカードが保存済み表示中かどうか（キーボードツールバーの表示用）
    private var isFocusedCardSaved: Bool {
        switch focusedCard {
        case .today: viewModel.todaySaved
        case .yesterday: viewModel.yesterdaySaved
        case .dayBeforeYesterday: viewModel.dayBeforeYesterdaySaved
        case nil: false
        }
    }

    /// キーボードが閉じてレイアウトが確定した後に、指定オフセットへスクロール位置を戻す。
    /// キーボードのインセット変化による自動スクロールを上書きして、編集時の表示位置を保つ。
    private func restoreScrollOffset(_ offset: CGFloat) {
        Task {
            // キーボードの閉じるアニメーション（約0.25秒）の完了を待ってから戻す
            try? await Task.sleep(for: .milliseconds(300))
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

#Preview {
    let authViewModel = AuthViewModel()
    HomeView(authViewModel: authViewModel, syncManager: SyncManager(authViewModel: authViewModel))
}
