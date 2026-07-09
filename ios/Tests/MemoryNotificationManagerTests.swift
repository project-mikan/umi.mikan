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
}
