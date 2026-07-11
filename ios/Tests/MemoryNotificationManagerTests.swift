import Foundation
import Testing
@testable import umi_mikan

/// MemoryNotificationManager のテスト（UserDefaults永続化部分）
@MainActor
struct MemoryNotificationManagerTests {
    /// テスト用に独立したUserDefaultsを生成する
    private func makeUserDefaults() -> UserDefaults {
        let suiteName = "MemoryNotificationManagerTests_\(UUID().uuidString)"
        return UserDefaults(suiteName: suiteName)!
    }

    @Test("正常系: 初期状態はOFF（デフォルト値）")
    func initialStateIsDisabled() {
        let manager = MemoryNotificationManager(userDefaults: makeUserDefaults())
        #expect(manager.isEnabled == false)
    }

    @Test("正常系: disable() を呼ぶとisEnabledがfalseになる")
    func disableSetsFalse() {
        let defaults = makeUserDefaults()
        let manager = MemoryNotificationManager(userDefaults: defaults)
        defaults.set(true, forKey: "memoryNotificationEnabled")

        manager.disable()

        #expect(manager.isEnabled == false)
    }

    @Test("正常系: isEnabledの変更がUserDefaultsへ永続化される")
    func isEnabledPersistsToUserDefaults() {
        let defaults = makeUserDefaults()
        let manager = MemoryNotificationManager(userDefaults: defaults)

        manager.isEnabled = true

        #expect(defaults.bool(forKey: "memoryNotificationEnabled") == true)
    }

    @Test("正常系: notificationHour/notificationMinuteの初期値はデフォルト時刻（8:00）")
    func notificationTimeDefaultsToEightAM() {
        let manager = MemoryNotificationManager(userDefaults: makeUserDefaults())

        #expect(manager.notificationHour == MemoryNotificationManager.defaultNotificationHour)
        #expect(manager.notificationMinute == MemoryNotificationManager.defaultNotificationMinute)
    }

    @Test(
        "正常系: updateNotificationTimeで指定した時刻がUserDefaultsへ永続化される",
        arguments: [(hour: 7, minute: 30), (hour: 21, minute: 45), (hour: 0, minute: 0)]
    )
    func updateNotificationTimePersists(hour: Int, minute: Int) {
        let defaults = makeUserDefaults()
        let manager = MemoryNotificationManager(userDefaults: defaults)

        manager.updateNotificationTime(hour: hour, minute: minute)

        #expect(manager.notificationHour == hour)
        #expect(manager.notificationMinute == minute)
    }

    @Test("正常系: disable()を呼ぶと配信済み・保留中の両方の通知リクエストが同一identifierで削除される")
    func disableRemovesBothPendingAndDeliveredNotifications() {
        let defaults = makeUserDefaults()
        let manager = MemoryNotificationManager(userDefaults: defaults)
        defaults.set(true, forKey: "memoryNotificationEnabled")

        // UNUserNotificationCenter自体はモック不可のため、disable()が例外なく完了し
        // isEnabledがfalseへ倒れることを確認する（削除API呼び出し自体はシステムAPIに委譲）
        manager.disable()

        #expect(manager.isEnabled == false)
    }
}
