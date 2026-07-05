import AppIntents
import Foundation

extension Notification.Name {
    /// Live Activityの同期ボタンから同期実行を要求する通知
    nonisolated static let syncPendingDiariesRequested = Notification.Name("net.usuyuki.umi-mikan.syncPendingDiariesRequested")
}

/// Live Activityの「今すぐ同期」ボタンで実行されるインテント。
/// LiveActivityIntent はアプリ本体のプロセスで実行されるため、
/// NotificationCenter 経由で SyncManager に同期を依頼する。
struct SyncPendingDiariesIntent: LiveActivityIntent {
    static let title: LocalizedStringResource = "未同期の日記を同期"

    /// 同期要求の通知を送信する（実際の同期はアプリ本体のSyncManagerが行う）
    func perform() async throws -> some IntentResult {
        NotificationCenter.default.post(name: .syncPendingDiariesRequested, object: nil)
        return .result()
    }
}
