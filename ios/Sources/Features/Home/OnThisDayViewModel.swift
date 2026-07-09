import Connect
import Foundation

/// 「n年前の今日」の1件分の表示データ
struct OnThisDayItem: Identifiable {
    let yearsAgo: Int
    let entry: Diary_DiaryEntry

    var id: String {
        entry.id
    }
}

/// ホーム画面の「n年前の今日」セクション用ViewModel。
///
/// 過去の同月同日（当年を除く）の日記を GetDiaryEntries でまとめて取得する。
/// 付加的な導線のため、オフライン時・取得失敗時はローカルキャッシュを持たず単に空表示にする。
@MainActor
@Observable
final class OnThisDayViewModel {
    /// 遡る年数の上限（何年前まで問い合わせるか）。
    /// デフォルト引数（nonisolated コンテキスト）から参照できるよう nonisolated にする。
    nonisolated static let maxYearsToLookBack = 30

    var items: [OnThisDayItem] = []

    private let authViewModel: AuthViewModel

    init(authViewModel: AuthViewModel) {
        self.authViewModel = authViewModel
    }

    /// 今日の日付を基準に、過去（当年を除く）の同月同日の Diary_YMD を年降順（直近年から）で列挙する。
    /// 2/29 は該当年にその日付が存在しないため、Calendar が nil を返しスキップされる。
    /// 純粋な日付計算のみで @MainActor 依存がないため、テストから同期的に呼び出せるよう nonisolated にする。
    nonisolated static func pastYearsSameDayDates(
        today: Date,
        calendar: Calendar,
        maxYearsToLookBack: Int = maxYearsToLookBack
    ) -> [(yearsAgo: Int, ymd: Diary_YMD)] {
        guard maxYearsToLookBack > 0 else { return [] }
        return (1 ... maxYearsToLookBack).compactMap { yearsAgo -> (Int, Diary_YMD)? in
            guard let pastDate = calendar.date(byAdding: .year, value: -yearsAgo, to: today) else { return nil }
            // うるう年の2/29はcalendar.date(byAdding:)が2/28や3/1へ丸めず、
            // 該当日が存在しない年はnilを返すため、月日が一致するかを明示的に確認する
            let todayComponents = calendar.dateComponents([.month, .day], from: today)
            let pastComponents = calendar.dateComponents([.year, .month, .day], from: pastDate)
            guard pastComponents.month == todayComponents.month, pastComponents.day == todayComponents.day else {
                return nil
            }
            var ymd = Diary_YMD()
            ymd.year = UInt32(pastComponents.year ?? 0)
            ymd.month = UInt32(pastComponents.month ?? 0)
            ymd.day = UInt32(pastComponents.day ?? 0)
            return (yearsAgo, ymd)
        }
    }

    /// サーバーから「n年前の今日」の日記を取得する。
    /// オフライン時・エラー時は静かに失敗し items を空のままにする（付加的な導線のため通常表示を妨げない）。
    func load(today: Date = Date()) async {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "Asia/Tokyo")!

        let dates = Self.pastYearsSameDayDates(today: today, calendar: calendar)
        guard !dates.isEmpty else {
            items = []
            return
        }

        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Diary_GetDiaryEntriesRequest()
        request.dates = dates.map(\.ymd)

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.getDiaryEntries(request: request, headers: ConnectClient.shared.headers())
        }

        guard response.error == nil, let entries = response.message?.entries else {
            items = []
            return
        }

        // yearsAgo を引けるようにYMDでの逆引きマップを作る
        let yearsAgoByYear = Dictionary(uniqueKeysWithValues: dates.map { ($0.ymd.year, $0.yearsAgo) })

        items = entries
            .compactMap { entry -> OnThisDayItem? in
                guard let yearsAgo = yearsAgoByYear[entry.date.year] else { return nil }
                return OnThisDayItem(yearsAgo: yearsAgo, entry: entry)
            }
            .sorted { $0.yearsAgo < $1.yearsAgo }
    }
}
