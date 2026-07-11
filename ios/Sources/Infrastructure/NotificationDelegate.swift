import Foundation
import UserNotifications

extension Notification.Name {
    /// 「おもいで」通知がタップされた時にHome画面へ遷移（スクロール）を要求する通知
    nonisolated static let memoryNotificationTapped = Notification.Name("net.usuyuki.umi-mikan.memoryNotificationTapped")
}

/// アプリがフォアグラウンドにいる間もローカル通知をバナー表示し、
/// 通知タップ時にHome画面の「おもいで」セクションへ遷移させるためのデリゲート。
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

    /// 通知タップ時、Home画面の「おもいで」セクションへ遷移させる。
    /// ADR 0014 の最小実装方針（スクロール位置合わせは後回し可）に従い、
    /// まずはHome画面を確実に開かせることを優先する。
    nonisolated func userNotificationCenter(
        _ center: UNUserNotificationCenter,
        didReceive response: UNNotificationResponse,
        withCompletionHandler completionHandler: @escaping () -> Void
    ) {
        NotificationCenter.default.post(name: .memoryNotificationTapped, object: nil)
        completionHandler()
    }
}
