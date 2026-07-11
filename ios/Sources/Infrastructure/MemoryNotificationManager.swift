import Foundation
import UserNotifications

/// 「おもいで」（n年前の今日）ローカル通知の設定・スケジューリングを管理する。
///
/// 通知はデフォルトOFFで、設定画面のトグルからのみONにできる（ユーザーの明示的なオプトインを必須にする）。
/// サーバー側で該当日記の有無を判定しないため、通知本文は誘導文言に留める。
@MainActor
final class MemoryNotificationManager {
    static let shared = MemoryNotificationManager()

    /// 通知時刻のデフォルト値（設定未保存時に使用）
    static let defaultNotificationHour = 8
    static let defaultNotificationMinute = 0

    /// スケジュール済み通知の識別子
    private static let notificationIdentifier = "memory_daily_notification"
    /// 設定トグルの永続化キー
    private static let enabledDefaultsKey = "memoryNotificationEnabled"
    /// 通知時刻（時）の永続化キー
    private static let hourDefaultsKey = "memoryNotificationHour"
    /// 通知時刻（分）の永続化キー
    private static let minuteDefaultsKey = "memoryNotificationMinute"

    /// 過去にリネーム・廃止された通知識別子の一覧。
    /// アプリ更新前にこれらの識別子で登録されたローカル通知はOS側に残り続けるため、
    /// 起動時に明示的に削除しないと新通知と二重に発火してしまう。
    /// 今後 identifier を変更する場合は、旧値をこの配列に追加するだけで後片付けが効くようにする
    /// （CLAUDE.md: 識別子/UserDefaultsキー変更時は必ず旧値の後片付け処理もセットで追加すること）。
    private static let legacyNotificationIdentifiers = [
        // OnThisDayNotificationManager → MemoryNotificationManager へのリネームで使われなくなった識別子
        "on_this_day_daily_notification"
    ]

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

    /// 通知時刻（時）。未設定時はデフォルト値を返す。
    var notificationHour: Int {
        get { userDefaults.object(forKey: Self.hourDefaultsKey) as? Int ?? Self.defaultNotificationHour }
        set { userDefaults.set(newValue, forKey: Self.hourDefaultsKey) }
    }

    /// 通知時刻（分）。未設定時はデフォルト値を返す。
    var notificationMinute: Int {
        get { userDefaults.object(forKey: Self.minuteDefaultsKey) as? Int ?? Self.defaultNotificationMinute }
        set { userDefaults.set(newValue, forKey: Self.minuteDefaultsKey) }
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

    /// トグルをOFFにし、スケジュール済み・配信済みの通知を取り消す。
    func disable() {
        isEnabled = false
        notificationCenter.removePendingNotificationRequests(withIdentifiers: [Self.notificationIdentifier])
        notificationCenter.removeDeliveredNotifications(withIdentifiers: [Self.notificationIdentifier])
    }

    /// 通知時刻を変更する。ONの場合は新しい時刻で即座に再スケジュールする。
    func updateNotificationTime(hour: Int, minute: Int) {
        notificationHour = hour
        notificationMinute = minute
        guard isEnabled else { return }
        scheduleDailyNotification()
    }

    /// システム側の通知許可状態を確認する。
    /// トグルがONなのにシステム側で拒否されている場合は、呼び出し側でトグル表示を補正するために使う。
    func isSystemAuthorized() async -> Bool {
        let settings = await notificationCenter.notificationSettings()
        return settings.authorizationStatus == .authorized
    }

    /// 過去にリネーム・廃止された識別子が残した孤児通知を削除する。
    /// アプリ起動時に一度だけ呼び出すこと。旧バージョンで通知をONにしていたユーザーの端末には
    /// 新識別子と無関係に毎日発火する古い通知リクエストが残ったままになり、
    /// 新通知と合わせて「朝に2回通知が来る」原因になっていた。
    func migrateLegacyNotification() {
        notificationCenter.removePendingNotificationRequests(withIdentifiers: Self.legacyNotificationIdentifiers)
        notificationCenter.removeDeliveredNotifications(withIdentifiers: Self.legacyNotificationIdentifiers)
    }

    /// 毎日決まった時刻に繰り返し発火するローカル通知をスケジュールする
    private func scheduleDailyNotification() {
        let content = UNMutableNotificationContent()
        content.title = "おもいで"
        content.body = "n年前の今日の日記を振り返りましょう"
        content.sound = .default

        var dateComponents = DateComponents()
        dateComponents.hour = notificationHour
        dateComponents.minute = notificationMinute
        let trigger = UNCalendarNotificationTrigger(dateMatching: dateComponents, repeats: true)

        let request = UNNotificationRequest(identifier: Self.notificationIdentifier, content: content, trigger: trigger)
        notificationCenter.removePendingNotificationRequests(withIdentifiers: [Self.notificationIdentifier])
        notificationCenter.removeDeliveredNotifications(withIdentifiers: [Self.notificationIdentifier])
        notificationCenter.add(request)
    }
}
