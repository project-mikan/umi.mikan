import UserNotifications

/// アプリがフォアグラウンドにいる間もローカル通知をバナー表示するためのデリゲート。
///
/// UNUserNotificationCenterDelegate 未設定の場合、iOSはフォアグラウンド中の通知を
/// デフォルトで抑制する（バナー等は出ないが通知センターには配信済みとして残る）。
/// 明示的にハンドリングしないままユーザーがバックグラウンド/フォアグラウンドを行き来すると
/// 「表示されなかったはずの通知が後から重なって見える」体感を生みやすいため、
/// ここで一貫してバナー・サウンドを表示する。
@MainActor
final class NotificationDelegate: NSObject, UNUserNotificationCenterDelegate {
    static let shared = NotificationDelegate()

    /// フォアグラウンド中に通知を受け取った時、バナー・サウンドを表示する
    nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        willPresent notification: UNNotification,
        withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
    ) {
        completionHandler([.banner, .sound, .badge])
    }
}
