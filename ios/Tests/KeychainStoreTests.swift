import Testing
@testable import umi_mikan

/// KeychainStoreのテスト（並列実行を避けるためにシリアル実行）
@Suite(.serialized)
struct KeychainStoreTests {
    /// 各テスト前にKeychainをクリアする
    init() {
        KeychainStore.deleteAll()
    }

    @Test("正常系: トークンを保存して取得できる")
    func saveAndLoad() {
        KeychainStore.save("test-access-token", for: .accessToken)
        #expect(KeychainStore.load(.accessToken) == "test-access-token")
    }

    @Test("正常系: 上書き保存が反映される")
    func overwrite() {
        KeychainStore.save("first", for: .accessToken)
        KeychainStore.save("second", for: .accessToken)
        #expect(KeychainStore.load(.accessToken) == "second")
    }

    @Test("正常系: deleteAllでアクセストークンとリフレッシュトークンが削除される")
    func deleteAll() {
        KeychainStore.save("access", for: .accessToken)
        KeychainStore.save("refresh", for: .refreshToken)
        KeychainStore.deleteAll()
        #expect(KeychainStore.load(.accessToken) == nil)
        #expect(KeychainStore.load(.refreshToken) == nil)
    }

    @Test("正常系: 存在しないキーはnilを返す")
    func loadNonExistent() {
        #expect(KeychainStore.load(.refreshToken) == nil)
    }

    @Test("正常系: accessTokenとrefreshTokenは独立して管理される")
    func independentKeys() {
        KeychainStore.save("access-value", for: .accessToken)
        KeychainStore.save("refresh-value", for: .refreshToken)
        #expect(KeychainStore.load(.accessToken) == "access-value")
        #expect(KeychainStore.load(.refreshToken) == "refresh-value")
    }
}
