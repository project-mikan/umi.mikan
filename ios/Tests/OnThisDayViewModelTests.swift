import Foundation
import Testing
@testable import umi_mikan

/// OnThisDayViewModel.pastYearsSameDayDates のテスト
struct OnThisDayViewModelTests {
    /// JST固定のカレンダーを生成する
    private func jstCalendar() -> Calendar {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "Asia/Tokyo")!
        return calendar
    }

    /// JSTの年月日からDateを生成する
    private func jstDate(_ year: Int, _ month: Int, _ day: Int) -> Date {
        var components = DateComponents()
        components.year = year
        components.month = month
        components.day = day
        components.hour = 12
        return jstCalendar().date(from: components)!
    }

    @Test("正常系: 通常日は指定年数分すべて過去の同月同日を返す")
    func normalDayReturnsAllYears() {
        let today = jstDate(2026, 7, 10)
        let result = OnThisDayViewModel.pastYearsSameDayDates(
            today: today,
            calendar: jstCalendar(),
            maxYearsToLookBack: 3
        )

        #expect(result.count == 3)
        #expect(result.map(\.yearsAgo) == [1, 2, 3])
        #expect(result[0].ymd.year == 2025)
        #expect(result[0].ymd.month == 7)
        #expect(result[0].ymd.day == 10)
        #expect(result[1].ymd.year == 2024)
        #expect(result[2].ymd.year == 2023)
    }

    @Test("異常系: 2/29が今日の場合、うるう年でない過去年はスキップされる")
    func leapDaySkipsNonLeapYears() {
        // 2024/2/29 を基準に3年分遡る（2023, 2022, 2021はいずれも非うるう年）
        let today = jstDate(2024, 2, 29)
        let result = OnThisDayViewModel.pastYearsSameDayDates(
            today: today,
            calendar: jstCalendar(),
            maxYearsToLookBack: 4
        )

        // 4年前の2020年はうるう年なので該当する。2021-2023は非うるう年なので除外される
        #expect(result.count == 1)
        #expect(result[0].yearsAgo == 4)
        #expect(result[0].ymd.year == 2020)
        #expect(result[0].ymd.month == 2)
        #expect(result[0].ymd.day == 29)
    }

    @Test("正常系: maxYearsToLookBackが0の場合は空配列を返す")
    func zeroLookBackReturnsEmpty() {
        let today = jstDate(2026, 7, 10)
        let result = OnThisDayViewModel.pastYearsSameDayDates(
            today: today,
            calendar: jstCalendar(),
            maxYearsToLookBack: 0
        )

        #expect(result.isEmpty)
    }

    @Test("正常系: 結果は年降順（直近年から）で並ぶ")
    func resultsAreOrderedByMostRecentYearFirst() {
        let today = jstDate(2026, 1, 1)
        let result = OnThisDayViewModel.pastYearsSameDayDates(
            today: today,
            calendar: jstCalendar(),
            maxYearsToLookBack: 5
        )

        let years = result.map(\.ymd.year)
        #expect(years == years.sorted(by: >))
    }
}
