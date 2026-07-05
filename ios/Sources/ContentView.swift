import SwiftUI

/// 認証状態に応じてログイン画面とメイン画面を切り替えるルートビュー
struct ContentView: View {
    @State private var authViewModel = AuthViewModel()

    /// 初期読み込み状態（スプラッシュ表示の制御、Preview等ではnil可）
    var launchState: AppLaunchState?

    var body: some View {
        if authViewModel.isLoggedIn {
            MainView(authViewModel: authViewModel, launchState: launchState)
        } else {
            LoginView(viewModel: authViewModel)
                .onAppear {
                    // 未ログイン時は読み込みがないためスプラッシュを即終了する
                    launchState?.isInitialLoading = false
                }
        }
    }
}

/// メイン画面 - ホーム・月ごと・検索タブを持つTabView
struct MainView: View {
    let authViewModel: AuthViewModel
    let launchState: AppLaunchState?

    @State private var syncManager: SyncManager

    @Environment(\.scenePhase)
    private var scenePhase

    // swiftlint:disable:next type_contents_order
    init(authViewModel: AuthViewModel, launchState: AppLaunchState? = nil) {
        self.authViewModel = authViewModel
        self.launchState = launchState
        _syncManager = State(initialValue: SyncManager(authViewModel: authViewModel))
    }

    var body: some View {
        TabView {
            Tab("ホーム", systemImage: "house.fill") { homeTab }
            Tab("月ごと", systemImage: "calendar") { monthlyTab }
            Tab("検索", systemImage: "magnifyingglass") { searchTab }
            Tab("よびな", systemImage: "person.text.rectangle") { entitiesTab }
            Tab("設定", systemImage: "gearshape") { settingsTab }
        }
        .onChange(of: scenePhase) { _, newPhase in
            // フォアグラウンド復帰時に未同期の編集をサーバーへ送信する
            if newPhase == .active {
                Task { await syncManager.syncPending() }
            }
        }
    }

    private var homeTab: some View {
        NavigationStack {
            HomeView(authViewModel: authViewModel, syncManager: syncManager, launchState: launchState)
                .navigationTitle("日記")
                .navigationBarTitleDisplayMode(.large)
        }
    }

    private var monthlyTab: some View {
        NavigationStack {
            MonthlyView(authViewModel: authViewModel, syncManager: syncManager)
                .navigationTitle("月ごと")
                .navigationBarTitleDisplayMode(.inline)
        }
    }

    private var searchTab: some View {
        NavigationStack {
            SearchView(authViewModel: authViewModel, syncManager: syncManager)
                .navigationTitle("検索")
                .navigationBarTitleDisplayMode(.inline)
        }
    }

    private var entitiesTab: some View {
        NavigationStack {
            EntitiesView(authViewModel: authViewModel)
                .navigationTitle("よびな")
                .navigationBarTitleDisplayMode(.inline)
        }
    }

    private var settingsTab: some View {
        NavigationStack {
            SettingsView(authViewModel: authViewModel)
                .navigationTitle("設定")
                .navigationBarTitleDisplayMode(.inline)
        }
    }
}
