import Connect
import Foundation

/// 今日・昨日・一昨日の日記データを保持する構造体
struct DiaryDayData {
    let date: Diary_YMD
    var entry: Diary_DiaryEntry?
}

/// ホーム画面の日記データと保存操作を管理するViewModel。
///
/// オフライン対応のため、ローカルストアを優先して表示し、
/// 保存はローカルへ書き込んだ後にSyncManagerがサーバーへ同期する。
@MainActor
@Observable
final class DiaryViewModel {
    var today: DiaryDayData = .init(date: Diary_YMD())
    var yesterday: DiaryDayData = .init(date: Diary_YMD())
    var dayBeforeYesterday: DiaryDayData = .init(date: Diary_YMD())

    var isLoading: Bool = false
    var errorMessage: String?

    /// 保存中・保存完了フラグ（今日・昨日・一昨日）
    var todaySaving: Bool = false
    var yesterdaySaving: Bool = false
    var dayBeforeYesterdaySaving: Bool = false

    var todaySaved: Bool = false
    var yesterdaySaved: Bool = false
    var dayBeforeYesterdaySaved: Bool = false

    /// ローカルにデータが1件でもあるかどうか（スプラッシュ解除の判断に使う）
    var hasLocalData: Bool {
        today.entry != nil || yesterday.entry != nil || dayBeforeYesterday.entry != nil
    }

    private let authViewModel: AuthViewModel
    private let syncManager: SyncManager
    private let store: LocalDiaryStore
    /// 保存済みフラグリセットタスク（協調キャンセルのため保持する）
    private var todaySavedResetTask: Task<Void, Never>?
    private var yesterdaySavedResetTask: Task<Void, Never>?
    private var dayBeforeSavedResetTask: Task<Void, Never>?

    init(authViewModel: AuthViewModel, syncManager: SyncManager, store: LocalDiaryStore = .shared) {
        self.authViewModel = authViewModel
        self.syncManager = syncManager
        self.store = store
    }

    /// 今日・昨日・一昨日の日付を計算して初期化する。
    /// バックエンドが JST 基準で日付を管理するため、カレンダーも JST 固定にする。
    func setup() {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "Asia/Tokyo")!
        let now = Date()
        today = DiaryDayData(date: ymd(from: now, calendar: calendar))
        yesterday = DiaryDayData(date: ymd(from: calendar.date(byAdding: .day, value: -1, to: now)!, calendar: calendar))
        dayBeforeYesterday = DiaryDayData(date: ymd(from: calendar.date(byAdding: .day, value: -2, to: now)!, calendar: calendar))
        reloadFromStore()
    }

    /// ローカルストアから即座に読み込む（オフラインでも表示できる）
    func loadLocal() {
        setup()
    }

    /// サーバーから最新の日記を取得してローカルへ反映する。
    /// オフライン時は静かに失敗しローカルデータを表示し続ける。
    func refreshFromServer() async {
        errorMessage = nil
        isLoading = !hasLocalData

        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Diary_GetDiaryEntriesRequest()
        request.dates = [today.date, yesterday.date, dayBeforeYesterday.date]

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.getDiaryEntries(request: request, headers: ConnectClient.shared.headers())
        }

        if let error = response.error {
            // オフラインはエラー扱いしない（ローカルデータで動作継続）
            if !APIHelper.isNetworkError(error) {
                errorMessage = APIHelper.errorMessage(error)
            }
            isLoading = false
            return
        }

        for entry in response.message?.entries ?? [] {
            store.applyServerEntry(entry)
        }
        reloadFromStore()
        isLoading = false
    }

    /// 今日の日記を保存する
    func saveToday(content: String) async {
        todaySaving = true
        defer { todaySaving = false }
        today.entry = await saveLocally(date: today.date, content: content)
        todaySaved = true
        resetSavedFlagLater(task: &todaySavedResetTask) { self.todaySaved = false }
    }

    /// 昨日の日記を保存する
    func saveYesterday(content: String) async {
        yesterdaySaving = true
        defer { yesterdaySaving = false }
        yesterday.entry = await saveLocally(date: yesterday.date, content: content)
        yesterdaySaved = true
        resetSavedFlagLater(task: &yesterdaySavedResetTask) { self.yesterdaySaved = false }
    }

    /// 一昨日の日記を保存する
    func saveDayBeforeYesterday(content: String) async {
        dayBeforeYesterdaySaving = true
        defer { dayBeforeYesterdaySaving = false }
        dayBeforeYesterday.entry = await saveLocally(date: dayBeforeYesterday.date, content: content)
        dayBeforeYesterdaySaved = true
        resetSavedFlagLater(task: &dayBeforeSavedResetTask) { self.dayBeforeYesterdaySaved = false }
    }

    // MARK: - Private

    /// ローカルへ保存して同期を試みる。オフラインでも保存自体は必ず成功する。
    private func saveLocally(date: Diary_YMD, content: String) async -> Diary_DiaryEntry? {
        store.saveLocalEdit(date: date, content: content)
        syncManager.refreshPendingCount()
        await syncManager.syncPending()
        return store.entry(for: date)?.toProto()
    }

    /// ストアから3日分のエントリを再読込する
    private func reloadFromStore() {
        today.entry = store.entry(for: today.date)?.toProto()
        yesterday.entry = store.entry(for: yesterday.date)?.toProto()
        dayBeforeYesterday.entry = store.entry(for: dayBeforeYesterday.date)?.toProto()
    }

    /// 2秒後に保存済みフラグをリセットする（協調キャンセル可能なタスクとして保持する）
    private func resetSavedFlagLater(task: inout Task<Void, Never>?, _ reset: @escaping @MainActor () -> Void) {
        task?.cancel()
        task = Task {
            do {
                try await Task.sleep(for: .seconds(2))
                reset()
            } catch {
                // キャンセル時は何もしない
            }
        }
    }

    /// Date を Diary_YMD に変換する
    private func ymd(from date: Date, calendar: Calendar) -> Diary_YMD {
        var ymd = Diary_YMD()
        ymd.year = UInt32(calendar.component(.year, from: date))
        ymd.month = UInt32(calendar.component(.month, from: date))
        ymd.day = UInt32(calendar.component(.day, from: date))
        return ymd
    }
}
