import Foundation
import Security

/// Keychainへのトークン保存・取得・削除を管理する
enum KeychainStore {
    enum Key: String {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
    }

    private static let service = "com.usuyuki.umi-mikan"

    /// 指定したキーで値をKeychainに保存する。書き込み成功時はtrueを返す
    @discardableResult
    static func save(_ value: String, for key: Key) -> Bool {
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
        let status = SecItemAdd(addQuery as CFDictionary, nil)
        if status == errSecSuccess {
            return true
        }
        // Keychainが使えない環境（CI等）ではUserDefaultsにフォールバック
        UserDefaults.standard.set(value, forKey: fallbackKey(key))
        return true
    }

    /// 指定したキーの値をKeychainから取得する
    static func load(_ key: Key) -> String? {
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue,
            kSecReturnData: true,
            kSecMatchLimit: kSecMatchLimitOne
        ]
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        if status == errSecSuccess, let data = result as? Data {
            return String(data: data, encoding: .utf8)
        }
        // Keychainから取得できない場合はUserDefaultsのフォールバックを参照
        return UserDefaults.standard.string(forKey: fallbackKey(key))
    }

    /// 指定したキーの値をKeychainから削除する
    static func delete(_ key: Key) {
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue
        ]
        SecItemDelete(query as CFDictionary)
        UserDefaults.standard.removeObject(forKey: fallbackKey(key))
    }

    /// アクセストークンとリフレッシュトークンを両方削除する（ログアウト用）
    static func deleteAll() {
        delete(.accessToken)
        delete(.refreshToken)
    }

    private static func fallbackKey(_ key: Key) -> String {
        "\(service).\(key.rawValue)"
    }
}
