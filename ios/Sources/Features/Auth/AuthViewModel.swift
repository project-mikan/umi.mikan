import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2TransportServices

/// ログイン・登録の状態と操作を管理するViewModel
@MainActor
@Observable
final class AuthViewModel {
    var isLoggedIn: Bool = false
    var isLoading: Bool = false
    var errorMessage: String?

    init() {
        // 起動時にKeychainにトークンがあればログイン済みとみなす
        isLoggedIn = KeychainStore.load(.accessToken) != nil
    }

    /// メールアドレスとパスワードでログインする
    func login(email: String, password: String) async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            try await GRPCClient.shared.withClient { client in
                let authClient = Auth_AuthService.Client(wrapping: client)
                var request = Auth_LoginByPasswordRequest()
                request.email = email
                request.password = password

                let response = try await authClient.loginByPassword(request)
                await MainActor.run {
                    KeychainStore.save(response.accessToken, for: .accessToken)
                    KeychainStore.save(response.refreshToken, for: .refreshToken)
                    self.isLoggedIn = true
                }
            }
        } catch {
            errorMessage = errorDescription(error)
        }
    }

    /// メールアドレス・パスワード・名前で新規登録する
    func register(email: String, password: String, name: String, registerKey: String) async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            try await GRPCClient.shared.withClient { client in
                let authClient = Auth_AuthService.Client(wrapping: client)
                var request = Auth_RegisterByPasswordRequest()
                request.email = email
                request.password = password
                request.name = name
                request.registerKey = registerKey

                let response = try await authClient.registerByPassword(request)
                await MainActor.run {
                    KeychainStore.save(response.accessToken, for: .accessToken)
                    KeychainStore.save(response.refreshToken, for: .refreshToken)
                    self.isLoggedIn = true
                }
            }
        } catch {
            errorMessage = errorDescription(error)
        }
    }

    /// ログアウトしてKeychainのトークンを削除する
    func logout() {
        KeychainStore.deleteAll()
        isLoggedIn = false
    }

    /// gRPCエラーを日本語メッセージに変換する
    private func errorDescription(_ error: Error) -> String {
        if let rpcError = error as? RPCError {
            switch rpcError.code {
            case .unauthenticated:
                return "メールアドレスまたはパスワードが正しくありません"
            case .notFound:
                return "ユーザーが見つかりません"
            case .alreadyExists:
                return "このメールアドレスは既に登録されています"
            case .permissionDenied:
                return "登録キーが正しくありません"
            case .invalidArgument:
                return "入力内容を確認してください"
            default:
                return "エラーが発生しました"
            }
        }
        return "接続エラーが発生しました"
    }
}
