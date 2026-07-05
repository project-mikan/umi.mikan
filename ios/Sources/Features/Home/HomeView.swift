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
            // ローカルストアから即座に表示し、読み込み済みならスプラッシュを終了する
            viewModel.loadLocal()
            applyLoadedContents()
            if viewModel.hasLocalData {
                launchState?.isInitialLoading = false
            }
            // サーバーから最新を取得して反映する
            await viewModel.refreshFromServer()
            applyLoadedContents(preservingEdits: true)
            launchState?.isInitialLoading = false
        }
        .onAppear {
            // 詳細画面から戻った時などにローカルの最新を反映する（入力途中の内容は保持）
            viewModel.loadLocal()
            applyLoadedContents(preservingEdits: true)
        }
        .refreshable {
            await viewModel.refreshFromServer()
            applyLoadedContents(preservingEdits: true)
        }
        .navigationDestination(for: Diary_YMD.self) { date in
            DiaryDetailView(date: date, authViewModel: authViewModel, syncManager: syncManager)
        }
        .overlay {
            if let error = viewModel.errorMessage {
                errorBanner(message: error)
            }
        }
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
            isSaved: viewModel.todaySaved
        ) {
            Task {
                await viewModel.saveToday(content: todayContent)
                lastAppliedToday = todayContent
            }
        }
    }

    private var yesterdayCard: some View {
        DiaryCardView(
            title: "昨日",
            date: viewModel.yesterday.date,
            content: $yesterdayContent,
            isSaving: viewModel.yesterdaySaving,
            isSaved: viewModel.yesterdaySaved
        ) {
            Task {
                await viewModel.saveYesterday(content: yesterdayContent)
                lastAppliedYesterday = yesterdayContent
            }
        }
    }

    private var dayBeforeYesterdayCard: some View {
        DiaryCardView(
            title: "一昨日",
            date: viewModel.dayBeforeYesterday.date,
            content: $dayBeforeYesterdayContent,
            isSaving: viewModel.dayBeforeYesterdaySaving,
            isSaved: viewModel.dayBeforeYesterdaySaved
        ) {
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

    private func errorBanner(message: String) -> some View {
        VStack {
            Spacer()
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
            .foregroundStyle(.white)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .glassEffect(.regular.tint(.red), in: .rect(cornerRadius: 12))
            .padding(.horizontal, 16)
            .padding(.bottom, 16)
        }
    }
}

#Preview {
    let authViewModel = AuthViewModel()
    HomeView(authViewModel: authViewModel, syncManager: SyncManager(authViewModel: authViewModel))
}
