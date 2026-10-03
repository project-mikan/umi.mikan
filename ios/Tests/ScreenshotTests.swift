import SwiftUI
import Testing
import UIKit
@testable import umi_mikan

/// PRで画面の見た目を人間が確認するためのスクリーンショット撮影テスト。
///
/// 画像の比較（合否判定）は行わず、撮影した PNG を出力ディレクトリへ保存するだけ。
/// 環境変数 SCREENSHOT_OUTPUT_DIR が設定されている時のみ実行し、通常の `make ios-test` ではスキップされる。
/// xcodebuild からは TEST_RUNNER_ 接頭辞付きで渡す（例: TEST_RUNNER_SCREENSHOT_OUTPUT_DIR=/path）。
/// 撮影は ios/scripts/capture-screenshots.sh（`make ios-screenshot` / CI）から行う。
@MainActor
@Suite(.serialized, .enabled(if: ScreenshotEnvironment.outputDirectory != nil))
struct ScreenshotTests {
    /// 本番サーバーへ通信させないための到達不能な接続先（接続拒否で即座にネットワークエラーになる）
    private static let unreachableHost = "http://127.0.0.1:9"

    @Test("正常系: 各画面のスクリーンショットを保存できる", arguments: ScreenshotScreen.allCases)
    func capture(screen: ScreenshotScreen) async throws {
        let outputDirectory = try #require(ScreenshotEnvironment.outputDirectory)

        // 実データの取得・トークンリフレッシュ（失敗時のログアウト）を起こさないよう、通信先を到達不能にする。
        // 各ViewModelはネットワークエラーをオフライン扱いにするため、ダミーデータのみが表示される
        ConnectClient.shared.replaceHost(Self.unreachableHost)
        defer { ConnectClient.shared.replaceHost(ConnectClient.defaultHost) }

        let fixture = ScreenshotFixture()
        let image = try await ScreenshotRenderer.capture(screen.makeView(fixture: fixture), wait: screen.renderWait)
        let data = try #require(image.pngData())

        try FileManager.default.createDirectory(at: outputDirectory, withIntermediateDirectories: true)
        let fileURL = outputDirectory.appendingPathComponent("\(ScreenshotEnvironment.deviceLabel)_\(screen.rawValue).png")
        try data.write(to: fileURL)
    }
}

// MARK: - 環境変数

/// スクリーンショット撮影の設定（環境変数から読む）
enum ScreenshotEnvironment {
    /// 画像の保存先。未設定なら撮影テスト自体を実行しない
    static var outputDirectory: URL? {
        guard let path = ProcessInfo.processInfo.environment["SCREENSHOT_OUTPUT_DIR"], !path.isEmpty else { return nil }
        return URL(fileURLWithPath: path, isDirectory: true)
    }

    /// ファイル名の接頭辞にする端末名（iphone / iphone-duo / ipad）。未設定なら端末種別から決める
    @MainActor static var deviceLabel: String {
        if let label = ProcessInfo.processInfo.environment["SCREENSHOT_DEVICE"], !label.isEmpty {
            return label
        }
        return UIDevice.current.userInterfaceIdiom == .pad ? "ipad" : "iphone"
    }
}

// MARK: - 撮影対象の画面

/// 撮影対象の画面（rawValue はファイル名に使う）
@MainActor
enum ScreenshotScreen: String, CaseIterable {
    case home
    case monthly
    case search
    case detail
    case settings

    /// 撮影までの待ち時間。.task の読み込みとシートの表示アニメーションが終わるのを待つ
    var renderWait: Duration {
        self == .detail ? .seconds(3) : .seconds(2)
    }

    /// ダミーデータを流し込んだ画面を生成する。
    /// MainView と同じく TabView + NavigationStack に載せ、iPad のサイドバーも含めて撮影する
    @ViewBuilder
    func makeView(fixture: ScreenshotFixture) -> some View {
        switch self {
        case .home:
            ScreenshotTabContainer(selected: .home) {
                homeScreen(fixture: fixture)
            }

        case .monthly:
            ScreenshotTabContainer(selected: .monthly) {
                NavigationStack {
                    MonthlyView(
                        authViewModel: fixture.authViewModel,
                        syncManager: fixture.syncManager,
                        store: fixture.store,
                        summaryStore: fixture.summaryStore
                    )
                    .navigationTitle("月ごと")
                    .navigationBarTitleDisplayMode(.inline)
                }
            }

        case .search:
            ScreenshotTabContainer(selected: .search) {
                NavigationStack {
                    SearchView(
                        viewModel: fixture.makeSearchViewModel(),
                        authViewModel: fixture.authViewModel,
                        syncManager: fixture.syncManager,
                        store: fixture.store
                    )
                    .navigationTitle("検索")
                    .navigationBarTitleDisplayMode(.inline)
                }
            }

        case .detail:
            // ホーム画面から今日の日記を開いた状態を再現する
            ScreenshotTabContainer(selected: .home) {
                homeScreen(fixture: fixture)
                    .sheet(isPresented: .constant(true)) {
                        DiaryDetailSheet(
                            items: fixture.homeDates.map { DiarySheetItem(date: $0) },
                            initialIndex: 0,
                            authViewModel: fixture.authViewModel,
                            syncManager: fixture.syncManager,
                            store: fixture.store
                        )
                    }
            }

        case .settings:
            ScreenshotTabContainer(selected: .settings) {
                NavigationStack {
                    SettingsView(viewModel: fixture.makeSettingsViewModel())
                        .navigationTitle("設定")
                        .navigationBarTitleDisplayMode(.inline)
                }
            }
        }
    }

    /// ホーム画面（MainView の homeTab と同じ構成）
    private func homeScreen(fixture: ScreenshotFixture) -> some View {
        NavigationStack {
            HomeView(authViewModel: fixture.authViewModel, syncManager: fixture.syncManager, store: fixture.store)
                .navigationTitle("日記")
                .navigationBarTitleDisplayMode(.large)
        }
    }
}

/// MainView のタブ構成を再現するコンテナ。
/// MainView は各画面を本番のストアで生成するため、ダミーデータを渡せるようテスト側で同じタブを組み立てる。
/// MainView のタブ（名前・アイコン・順序）を変えた場合はここも合わせること。
private struct ScreenshotTabContainer<Content: View>: View {
    /// MainView のタブ定義
    enum TabItem: CaseIterable {
        case home
        case monthly
        case search
        case entities
        case settings

        var title: String {
            switch self {
            case .home: "ホーム"
            case .monthly: "月ごと"
            case .search: "検索"
            case .entities: "よびな"
            case .settings: "設定"
            }
        }

        var systemImage: String {
            switch self {
            case .home: "house.fill"
            case .monthly: "calendar"
            case .search: "magnifyingglass"
            case .entities: "person.text.rectangle"
            case .settings: "gearshape"
            }
        }
    }

    let selected: TabItem
    @ViewBuilder let content: Content

    var body: some View {
        TabView(selection: .constant(selected)) {
            ForEach(TabItem.allCases, id: \.self) { tab in
                Tab(tab.title, systemImage: tab.systemImage, value: tab) {
                    // 選択中のタブだけ中身を描画する（他のタブは表示されないため空でよい）
                    if tab == selected {
                        content
                    } else {
                        Color.clear
                    }
                }
            }
        }
        .tabViewStyle(.sidebarAdaptable)
    }
}

// MARK: - ダミーデータ

/// 撮影用のダミーデータ一式。
/// 実データ（端末のローカルストア・要約キャッシュ）に触れないよう、すべて一時ファイルのストアを使う。
@MainActor
struct ScreenshotFixture {
    /// 日記本文のサンプル（日付ごとに順番に割り当てる）
    private static let sampleContents = [
        "朝から雨。駅前のカフェでモーニングを食べながら本を読んだ。午後は図書館で調べもの、帰りに八百屋でみかんを買った。夜は久しぶりに鍋にした。",
        "仕事が立て込んでいたけど、昼休みに少し散歩できたのが良かった。夕方から友人と電話して、来月の旅行の計画を立てた。",
        "休日。洗濯と掃除を済ませてから自転車で海まで行った。風が強かったが、夕日がとてもきれいだった。",
        "新しいカフェを開拓した。深煎りのコーヒーがおいしく、店主と少し話せた。また行きたい。",
        "朝ランニング5km。昼はパスタを作り、午後は溜まっていた映画を2本観た。",
        "会議続きで疲れた一日。帰りに銭湯に寄ってリフレッシュ。"
    ]

    let authViewModel = AuthViewModel()
    let store: LocalDiaryStore
    let summaryStore: DiarySummaryStore
    let syncManager: SyncManager
    /// ホーム画面に表示する今日・昨日・一昨日の日付（JST）
    let homeDates: [Diary_YMD]

    init() {
        let directory = FileManager.default.temporaryDirectory
            .appendingPathComponent("screenshot_\(UUID().uuidString)", isDirectory: true)
        store = LocalDiaryStore(fileURL: directory.appendingPathComponent("diaries.json"))
        summaryStore = DiarySummaryStore(fileURL: directory.appendingPathComponent("summaries.json"))
        syncManager = SyncManager(authViewModel: authViewModel, store: store)

        let calendar = Calendar.jst
        let now = Date()
        homeDates = (0 ..< 3).compactMap { calendar.date(byAdding: .day, value: -$0, to: now) }
            .map { Self.ymd(from: $0, calendar: calendar) }

        // 月ごと画面が埋まるよう今月の全日分と、ホーム画面用の3日分（月をまたぐ場合がある）を入れる
        let monthDates = Self.allDatesInMonth(of: now, calendar: calendar)
        for (index, date) in (monthDates + homeDates).enumerated() {
            store.applyServerEntry(Self.entry(
                date: date,
                content: Self.sampleContents[index % Self.sampleContents.count]
            ))
        }
    }

    /// キーワード「カフェ」で検索した直後の状態の検索ViewModelを生成する
    func makeSearchViewModel() -> SearchViewModel {
        let viewModel = SearchViewModel(authViewModel: authViewModel, store: store)
        var results = Diary_SearchDiaryEntriesResponse()
        results.searchedKeyword = "カフェ"
        results.expandedKeywords = ["喫茶店"]
        results.entries = homeDates.enumerated().map { index, date in
            Self.entry(date: date, content: Self.sampleContents[[0, 3, 0][index]])
        }
        viewModel.keyword = "カフェ"
        viewModel.keywordResults = results
        viewModel.hasSearched = true
        return viewModel
    }

    /// ユーザー情報を取得済みの状態の設定ViewModelを生成する
    func makeSettingsViewModel() -> SettingsViewModel {
        let viewModel = SettingsViewModel(authViewModel: authViewModel, notificationManager: MemoryNotificationManager.shared)
        viewModel.userName = "うみみかん"
        viewModel.email = "demo@example.com"
        return viewModel
    }

    /// サーバーから取得した体のダミー日記を生成する（needsSync にならないよう serverID を付ける）
    private static func entry(date: Diary_YMD, content: String) -> Diary_DiaryEntry {
        var entry = Diary_DiaryEntry()
        entry.id = "screenshot-\(LocalDiaryEntry.dateKey(date))"
        entry.date = date
        entry.content = content
        entry.updatedAt = Int64(Date().timeIntervalSince1970)
        return entry
    }

    /// 指定日を含む月の全日付を返す
    private static func allDatesInMonth(of date: Date, calendar: Calendar) -> [Diary_YMD] {
        guard
            let range = calendar.range(of: .day, in: .month, for: date),
            let firstDay = calendar.date(from: calendar.dateComponents([.year, .month], from: date))
        else {
            return []
        }
        return range.compactMap { calendar.date(byAdding: .day, value: $0 - 1, to: firstDay) }
            .map { ymd(from: $0, calendar: calendar) }
    }

    /// Date を Diary_YMD に変換する
    private static func ymd(from date: Date, calendar: Calendar) -> Diary_YMD {
        var ymd = Diary_YMD()
        ymd.year = UInt32(calendar.component(.year, from: date))
        ymd.month = UInt32(calendar.component(.month, from: date))
        ymd.day = UInt32(calendar.component(.day, from: date))
        return ymd
    }
}

// MARK: - 描画

/// SwiftUI の View を実際のウィンドウに表示してスクリーンショットを撮る。
/// Liquid Glass やマテリアルはオフスクリーン描画（layer.render）では再現されないため、
/// テストホストのウィンドウシーン上に最前面のウィンドウを出して drawHierarchy で撮影する。
@MainActor
enum ScreenshotRenderer {
    static func capture(_ view: some View, wait: Duration) async throws -> UIImage {
        let scene = try #require(
            UIApplication.shared.connectedScenes.compactMap { $0 as? UIWindowScene }.first,
            "テストホストのウィンドウシーンが見つかりません"
        )
        let window = UIWindow(windowScene: scene)
        window.frame = scene.effectiveGeometry.coordinateSpace.bounds
        // テストホスト本体の画面（ログイン画面など）より前面に出す
        window.windowLevel = .alert + 1
        // 実行環境の外観設定で結果が変わらないようライトモードに固定する
        window.overrideUserInterfaceStyle = .light
        window.rootViewController = UIHostingController(rootView: view)
        window.makeKeyAndVisible()
        defer { window.isHidden = true }

        try await Task.sleep(for: wait)

        let renderer = UIGraphicsImageRenderer(bounds: window.bounds)
        return renderer.image { _ in
            window.drawHierarchy(in: window.bounds, afterScreenUpdates: true)
        }
    }
}
