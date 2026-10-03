import SwiftUI
import Testing
import UIKit
@testable import umi_mikan

/// PRで人間が見た目を確認するための撮影で、比較はしない（ios/scripts/capture-screenshots.sh から実行する）
@MainActor
@Suite(.serialized, .enabled(if: ScreenshotEnvironment.outputDirectory != nil))
struct ScreenshotTests {
    /// 接続拒否で即座にネットワークエラーになり、各ViewModelがオフライン扱いにする
    private static let unreachableHost = "http://127.0.0.1:9"

    @Test("正常系: 各画面のスクリーンショットを保存できる", arguments: ScreenshotScreen.allCases)
    func capture(screen: ScreenshotScreen) async throws {
        let outputDirectory = try #require(ScreenshotEnvironment.outputDirectory)

        // 本番に繋ぐと実データの表示やトークンリフレッシュ失敗によるログアウトが起きる
        ConnectClient.shared.replaceHost(Self.unreachableHost)
        defer { ConnectClient.shared.replaceHost(ConnectClient.defaultHost) }

        try FileManager.default.createDirectory(at: outputDirectory, withIntermediateDirectories: true)
        let label = try #require(ScreenshotDevice.label(
            for: UIDevice.current.userInterfaceIdiom,
            deviceName: ScreenshotEnvironment.simulatorDeviceName
        ))
        let image = try await ScreenshotRenderer.capture(screen.makeView(fixture: ScreenshotFixture()), wait: screen.renderWait)
        let data = try #require(image.pngData())
        try data.write(to: outputDirectory.appendingPathComponent("\(label)_\(screen.rawValue).png"))
    }
}

struct ScreenshotDeviceTests {
    struct LabelCase: Sendable {
        let name: String
        let idiom: UIUserInterfaceIdiom
        let deviceName: String
        let expected: String?
    }

    @Test(
        "label: 端末種別とシミュレータ名からファイル名の接頭辞が決まる",
        arguments: [
            // 通常の iPhone シミュレータ
            LabelCase(name: "正常系: iPhoneシミュレータはiphoneになる", idiom: .phone, deviceName: "iPhone 17", expected: "iphone"),
            // iPhone Duo は iPhone 扱いの端末種別だが、safe area 等が違うため別の列として撮る
            LabelCase(name: "正常系: iPhone Duoシミュレータはiphone-duoになる", idiom: .phone, deviceName: "iPhone Duo", expected: "iphone-duo"),
            // iPad シミュレータ
            LabelCase(name: "正常系: iPadシミュレータはipadになる", idiom: .pad, deviceName: "iPad Pro 11-inch (M5)", expected: "ipad"),
            // 想定外の端末では撮らない
            LabelCase(name: "異常系: 想定外の端末を渡すとnilになるので撮影しない", idiom: .tv, deviceName: "Apple TV", expected: nil)
        ]
    )
    func label(testCase: LabelCase) {
        #expect(ScreenshotDevice.label(for: testCase.idiom, deviceName: testCase.deviceName) == testCase.expected, "\(testCase.name)")
    }
}

// MARK: - 環境変数

enum ScreenshotEnvironment {
    /// xcodebuild からは TEST_RUNNER_SCREENSHOT_OUTPUT_DIR として渡す
    static var outputDirectory: URL? {
        guard let path = ProcessInfo.processInfo.environment["SCREENSHOT_OUTPUT_DIR"], !path.isEmpty else { return nil }
        return URL(fileURLWithPath: path, isDirectory: true)
    }

    /// シミュレータが自動で設定する環境変数
    static var simulatorDeviceName: String {
        ProcessInfo.processInfo.environment["SIMULATOR_DEVICE_NAME"] ?? ""
    }
}

// MARK: - 端末ラベル

enum ScreenshotDevice {
    /// iPhone Duo は safe area 等が違い近似できないため、iPhone Duo シミュレータでのみ撮る
    static func label(for idiom: UIUserInterfaceIdiom, deviceName: String) -> String? {
        switch idiom {
        case .phone:
            deviceName.localizedCaseInsensitiveContains("Duo") ? "iphone-duo" : "iphone"

        case .pad:
            "ipad"

        default:
            nil
        }
    }
}

// MARK: - 撮影対象の画面

/// rawValue はファイル名に使う
@MainActor
enum ScreenshotScreen: String, CaseIterable {
    case home
    case monthly
    case search
    case detail
    case settings

    /// .task の読み込みとシートの表示アニメーションを待つ
    var renderWait: Duration {
        self == .detail ? .milliseconds(1500) : .milliseconds(700)
    }

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

    private func homeScreen(fixture: ScreenshotFixture) -> some View {
        NavigationStack {
            HomeView(authViewModel: fixture.authViewModel, syncManager: fixture.syncManager, store: fixture.store)
                .navigationTitle("日記")
                .navigationBarTitleDisplayMode(.large)
        }
    }
}

/// MainView は本番のストアで各画面を作るため同じタブを再現している（MainView のタブを変えたら合わせること）
private struct ScreenshotTabContainer<Content: View>: View {
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

/// 端末の実データ（ローカルストア・要約キャッシュ）に触れないよう一時ファイルのストアを使う
@MainActor
struct ScreenshotFixture {
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

        // ホームの3日分は月をまたぐ場合があるため今月分とは別に入れる
        let monthDates = Self.allDatesInMonth(of: now, calendar: calendar)
        for (index, date) in (monthDates + homeDates).enumerated() {
            store.applyServerEntry(Self.entry(
                date: date,
                content: Self.sampleContents[index % Self.sampleContents.count]
            ))
        }
    }

    /// serverID を付けて同期待ち（needsSync）扱いにならないようにする
    private static func entry(date: Diary_YMD, content: String) -> Diary_DiaryEntry {
        var entry = Diary_DiaryEntry()
        entry.id = "screenshot-\(LocalDiaryEntry.dateKey(date))"
        entry.date = date
        entry.content = content
        entry.updatedAt = Int64(Date().timeIntervalSince1970)
        return entry
    }

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

    private static func ymd(from date: Date, calendar: Calendar) -> Diary_YMD {
        var ymd = Diary_YMD()
        ymd.year = UInt32(calendar.component(.year, from: date))
        ymd.month = UInt32(calendar.component(.month, from: date))
        ymd.day = UInt32(calendar.component(.day, from: date))
        return ymd
    }

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

    func makeSettingsViewModel() -> SettingsViewModel {
        let viewModel = SettingsViewModel(authViewModel: authViewModel, notificationManager: MemoryNotificationManager.shared)
        viewModel.userName = "うみみかん"
        viewModel.email = "demo@example.com"
        return viewModel
    }
}

// MARK: - 描画

/// Liquid Glass やマテリアルは layer.render では描画されないため、実ウィンドウを出して drawHierarchy で撮る
@MainActor
enum ScreenshotRenderer {
    static func capture(_ view: some View, wait: Duration) async throws -> UIImage {
        let scene = try #require(
            UIApplication.shared.connectedScenes.compactMap { $0 as? UIWindowScene }.first,
            "テストホストのウィンドウシーンが見つかりません"
        )
        let window = UIWindow(windowScene: scene)
        window.frame = scene.effectiveGeometry.coordinateSpace.bounds
        window.windowLevel = .alert + 1
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
