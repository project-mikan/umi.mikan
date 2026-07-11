import Foundation

extension Calendar {
    /// バックエンドが JST 基準で日付を管理するため、日付計算は常にこの JST 固定カレンダーを使う。
    /// 生成コストがあるため、呼び出し側では毎回生成せずこのプロパティを参照する。
    static let jst: Calendar = {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "Asia/Tokyo")!
        return calendar
    }()
}

extension TimeZone {
    /// バックエンドが JST 基準で日付を管理するため、日時表示は常にこの JST 固定タイムゾーンを使う。
    static let jst = TimeZone(identifier: "Asia/Tokyo")!
}
