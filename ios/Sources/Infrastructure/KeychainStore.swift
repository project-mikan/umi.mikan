import Foundation
import Security

/// Keychainへのトークン保存・取得・削除を管理する
/// CI（CODE_SIGNING_ALLOWED=NO で entitlements 未設定）など Keychain が使えない環境では
/// UserDefaults にフォールバックする
enum KeychainStore {
    enum Key: String {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
    }

    private static let service = "com.usuyuki.umi-mikan"

    /// UserDefaults フォールバック保存時のキー接頭辞
    private static let fallbackKeyPrefix = "keychain_fallback_"

    /// Keychain が利用可能かどうかをプローブ書き込みで一度だけ判定してキャッシュする
    /// entitlements が無い環境では SecItemAdd が errSecMissingEntitlement (-34018) で失敗する
    private static let isKeychainAvailable: Bool = {
        let probeAccount = "__keychain_availability_probe__"
        let addQuery: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: probeAccount,
            kSecAttrAccessible: kSecAttrAccessibleAfterFirstUnlock,
            kSecValueData: Data("probe".utf8)
        ]
        let status = SecItemAdd(addQuery as CFDictionary, nil)
        guard status == errSecSuccess || status == errSecDuplicateItem else { return false }
        let deleteQuery: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: probeAccount
        ]
        SecItemDelete(deleteQuery as CFDictionary)
        return true
    }()

    /// 指定したキーで値をKeychainに保存する。書き込み成功時はtrueを返す
    @discardableResult
    static func save(_ value: String, for key: Key) -> Bool {
        guard isKeychainAvailable else {
            UserDefaults.standard.set(value, forKey: fallbackKeyPrefix + key.rawValue)
            return true
        }
        let data = Data(value.utf8)
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue
        ]
        SecItemDelete(query as CFDictionary)
        let addQuery: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue,
            // デバイスロック状態に依存しないアクセシビリティを設定し、CI環境でも動作させる
            kSecAttrAccessible: kSecAttrAccessibleAfterFirstUnlock,
            kSecValueData: data
        ]
        return SecItemAdd(addQuery as CFDictionary, nil) == errSecSuccess
    }

    /// 指定したキーの値をKeychainから取得する
    static func load(_ key: Key) -> String? {
        guard isKeychainAvailable else {
            return UserDefaults.standard.string(forKey: fallbackKeyPrefix + key.rawValue)
        }
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue,
            kSecReturnData: true,
            kSecMatchLimit: kSecMatchLimitOne
        ]
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        guard status == errSecSuccess, let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    /// 指定したキーの値をKeychainから削除する
    static func delete(_ key: Key) {
        guard isKeychainAvailable else {
            UserDefaults.standard.removeObject(forKey: fallbackKeyPrefix + key.rawValue)
            return
        }
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue
        ]
        SecItemDelete(query as CFDictionary)
    }

    /// アクセストークンとリフレッシュトークンを両方削除する（ログアウト用）
    static func deleteAll() {
        delete(.accessToken)
        delete(.refreshToken)
    }
}
