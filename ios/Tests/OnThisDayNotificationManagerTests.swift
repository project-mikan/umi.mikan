import Foundation
import Testing
@testable import umi_mikan

/// OnThisDayNotificationManager のテスト（UserDefaults永続化部分）
@MainActor
struct OnThisDayNotificationManagerTests {
    /// テスト用に独立したUserDefaultsを生成する
    private func makeUserDefaults() -> UserDefaults {
        let suiteName = "OnThisDayNotificationManagerTests_\(UUID().uuidString)"
        return UserDefaults(suiteName: suiteName)!
    }

    @Test("正常系: 初期状態はOFF（デフォルト値）")
    func initialStateIsDisabled() {
        let manager = OnThisDayNotificationManager(userDefaults: makeUserDefaults())
        #expect(manager.isEnabled == false)
    }

    @Test("正常系: disable() を呼ぶとisEnabledがfalseになる")
    func disableSetsFalse() {
        let defaults = makeUserDefaults()
        let manager = OnThisDayNotificationManager(userDefaults: defaults)
        defaults.set(true, forKey: "onThisDayNotificationEnabled")

        manager.disable()

        #expect(manager.isEnabled == false)
    }

    @Test("正常系: isEnabledの変更がUserDefaultsへ永続化される")
    func isEnabledPersistsToUserDefaults() {
        let defaults = makeUserDefaults()
        let manager = OnThisDayNotificationManager(userDefaults: defaults)

        manager.isEnabled = true

        #expect(defaults.bool(forKey: "onThisDayNotificationEnabled") == true)
    }
}
