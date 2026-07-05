import Connect
import Foundation

/// 月毎ページのViewModel - 指定月の日記一覧を管理する。
///
/// オフライン対応のため、ローカルストアを優先して表示し、
/// サーバーから取得できたらローカルへ反映する。
@MainActor
@Observable
final class MonthlyViewModel {
    var year: Int
    var month: Int
    /// 日（1〜31）をキーとした日記エントリのマップ
    var entryMap: [Int: Diary_DiaryEntry] = [:]
    /// 月間まとめ（生成済みの場合のみ）
    var monthlySummary: Diary_MonthlySummary?
    var isLoading: Bool = false
    var errorMessage: String?

    /// 表示中の月の日数
    var daysInMonth: Int {
        var components = DateComponents()
        components.year = year
        components.month = month
        let calendar = Calendar.current
        guard
            let date = calendar.date(from: components),
            let range = calendar.range(of: .day, in: .month, for: date)
        else {
            return 30
        }
        return range.count
    }

    private let authViewModel: AuthViewModel
    private let store: LocalDiaryStore

    init(authViewModel: AuthViewModel, store: LocalDiaryStore = .shared) {
        self.authViewModel = authViewModel
        self.store = store
        let now = Date()
        let calendar = Calendar.current
        year = calendar.component(.year, from: now)
        month = calendar.component(.month, from: now)
    }

    /// 表示中の月の日記エントリを取得する（ローカル優先＋サーバー同期）
    func fetch() async {
        errorMessage = nil

        // ローカルストアから即座に表示する
        loadLocalMonth()
        isLoading = entryMap.isEmpty

        // サーバーから最新を取得してローカルへ反映する
        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Diary_GetDiaryEntriesByMonthRequest()
        var ym = Diary_YM()
        ym.year = UInt32(year)
        ym.month = UInt32(month)
        request.month = ym

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.getDiaryEntriesByMonth(request: request, headers: ConnectClient.shared.headers())
        }

        if let error = response.error {
            // オフラインはエラー扱いしない（ローカルデータで動作継続）
            if !APIHelper.isNetworkError(error) {
                errorMessage = APIHelper.errorMessage(error)
            }
            isLoading = false
            return
        }

        for entry in response.message?.entries ?? [] where entry.hasDate {
            store.applyServerEntry(entry)
        }
        loadLocalMonth()
        isLoading = false

        // 月間まとめは取得できなくてもページ表示に影響させない
        await fetchMonthlySummary()
    }

    /// 月間まとめを取得する（未生成・エラー時は非表示にするだけ）
    func fetchMonthlySummary() async {
        monthlySummary = nil

        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Diary_GetMonthlySummaryRequest()
        var ym = Diary_YM()
        ym.year = UInt32(year)
        ym.month = UInt32(month)
        request.month = ym

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.getMonthlySummary(request: request, headers: ConnectClient.shared.headers())
        }

        guard
            let message = response.message,
            message.hasSummary,
            !message.summary.summary.isEmpty,
            message.summary.errorReason.isEmpty
        else {
            return
        }
        monthlySummary = message.summary
    }

    /// 前月へ移動する
    func previousMonth() async {
        if month == 1 {
            year -= 1
            month = 12
        } else {
            month -= 1
        }
        await fetch()
    }

    /// 翌月へ移動する
    func nextMonth() async {
        if month == 12 {
            year += 1
            month = 1
        } else {
            month += 1
        }
        await fetch()
    }

    /// 今月へ移動する
    func goToToday() async {
        let now = Date()
        let calendar = Calendar.current
        year = calendar.component(.year, from: now)
        month = calendar.component(.month, from: now)
        await fetch()
    }

    /// 指定した日の Diary_YMD を生成する
    func ymd(day: Int) -> Diary_YMD {
        var ymd = Diary_YMD()
        ymd.year = UInt32(year)
        ymd.month = UInt32(month)
        ymd.day = UInt32(day)
        return ymd
    }

    /// 指定した日の曜日名（例: "月"）を返す
    func weekdayName(day: Int) -> String {
        var components = DateComponents()
        components.year = year
        components.month = month
        components.day = day
        let calendar = Calendar.current
        guard let date = calendar.date(from: components) else { return "" }
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ja_JP")
        formatter.dateFormat = "E"
        return formatter.string(from: date)
    }

    // MARK: - Private

    /// ローカルストアから表示中の月のエントリを読み込む
    private func loadLocalMonth() {
        var map: [Int: Diary_DiaryEntry] = [:]
        for local in store.entries(year: year, month: month) {
            map[local.day] = local.toProto()
        }
        entryMap = map
    }
}
