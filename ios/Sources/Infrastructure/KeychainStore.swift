import Foundation
import Security

/// Keychainへのトークン保存・取得・削除を管理する
enum KeychainStore {
    enum Key: String {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
    }

    private static let service = "com.usuyuki.umi-mikan"

    /// 指定したキーで値をKeychainに保存する
    static func save(_ value: String, for key: Key) {
        let data = Data(value.utf8)
        let query: [CFString: Any] = [
            kSecClass: kSecClassGenericPassword,
            kSecAttrService: service,
            kSecAttrAccount: key.rawValue
        ]
        SecItemDelete(query as CFDictionary)
        var addQuery = query
        addQuery[kSecValueData] = data
        SecItemAdd(addQuery as CFDictionary, nil)
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
        guard status == errSecSuccess, let data = result as? Data else { return nil }
        return String(data: data, encoding: .utf8)
    }

    /// 指定したキーの値をKeychainから削除する
    static func delete(_ key: Key) {
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
