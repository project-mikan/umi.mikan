import Foundation
import UserNotifications

/// 「n年前の今日」ローカル通知の設定・スケジューリングを管理する。
///
/// 通知はデフォルトOFFで、設定画面のトグルからのみONにできる（ユーザーの明示的なオプトインを必須にする）。
/// サーバー側で該当日記の有無を判定しないため、通知本文は誘導文言に留める。
@MainActor
final class OnThisDayNotificationManager {
    static let shared = OnThisDayNotificationManager()

    /// 通知を毎日発火させる時刻
    static let notificationHour = 8
    static let notificationMinute = 0

    /// スケジュール済み通知の識別子
    private static let notificationIdentifier = "on_this_day_daily_notification"
    /// 設定トグルの永続化キー
    private static let enabledDefaultsKey = "onThisDayNotificationEnabled"

    private let notificationCenter: UNUserNotificationCenter
    private let userDefaults: UserDefaults

    // swiftlint:disable:next type_contents_order
    init(notificationCenter: UNUserNotificationCenter = .current(), userDefaults: UserDefaults = .standard) {
        self.notificationCenter = notificationCenter
        self.userDefaults = userDefaults
    }

    /// アプリ内トグルの状態（UserDefaultsへ永続化）。デフォルトはOFF。
    var isEnabled: Bool {
        get { userDefaults.bool(forKey: Self.enabledDefaultsKey) }
        set { userDefaults.set(newValue, forKey: Self.enabledDefaultsKey) }
    }

    /// トグルをONにする。通知許可をリクエストし、許可された場合のみ毎日の通知をスケジュールする。
    /// 許可されなかった場合はトグルをOFFへ戻す。
    @discardableResult
    func enable() async -> Bool {
        do {
            let granted = try await notificationCenter.requestAuthorization(options: [.alert, .sound, .badge])
            guard granted else {
                isEnabled = false
                return false
            }
        } catch {
            isEnabled = false
            return false
        }
        isEnabled = true
        scheduleDailyNotification()
        return true
    }

    /// トグルをOFFにし、スケジュール済みの通知を取り消す。
    func disable() {
        isEnabled = false
        notificationCenter.removePendingNotificationRequests(withIdentifiers: [Self.notificationIdentifier])
    }

    /// システム側の通知許可状態を確認する。
    /// トグルがONなのにシステム側で拒否されている場合は、呼び出し側でトグル表示を補正するために使う。
    func isSystemAuthorized() async -> Bool {
        let settings = await notificationCenter.notificationSettings()
        return settings.authorizationStatus == .authorized
    }

    /// 毎日決まった時刻に繰り返し発火するローカル通知をスケジュールする
    private func scheduleDailyNotification() {
        let content = UNMutableNotificationContent()
        content.title = "n年前の今日"
        content.body = "n年前の今日の日記を振り返りましょう"
        content.sound = .default

        var dateComponents = DateComponents()
        dateComponents.hour = Self.notificationHour
        dateComponents.minute = Self.notificationMinute
        let trigger = UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: true)

        let request = UNNotificationRequest(identifier: Self.notificationIdentifier, content: content, trigger: trigger)
        notificationCenter.removePendingNotificationRequests(withIdentifiers: [Self.notificationIdentifier])
        notificationCenter.add(request)
    }
}
