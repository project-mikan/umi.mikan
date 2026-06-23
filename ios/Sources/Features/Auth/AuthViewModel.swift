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
        // 起動時にKeychainにトークンがあればログイン済みとみなす。
        // アクセストークンは15分で期限切れになるため、次のAPI呼び出し時に
        // refreshAccessTokenで自動更新される。
        isLoggedIn = KeychainStore.load(.accessToken) != nil
    }

    /// メールアドレスとパスワードでログインする
    func login(email: String, password: String) async {
        isLoading = true
        errorMessage = nil

        do {
            let response = try await GRPCClient.shared.withClient { client in
                let authClient = Auth_AuthService.Client(wrapping: client)
                var request = Auth_LoginByPasswordRequest()
                request.email = email
                request.password = password

                return try await authClient.loginByPassword(request)
            }

            guard
                KeychainStore.save(response.accessToken, for: .accessToken),
                KeychainStore.save(response.refreshToken, for: .refreshToken)
            else {
                errorMessage = "トークンの保存に失敗しました。再度お試しください。"
                isLoading = false
                return
            }
            isLoading = false
            isLoggedIn = true
        } catch {
            errorMessage = errorDescription(error, isAuthEndpoint: true)
            isLoading = false
        }
    }

    /// メールアドレス・パスワード・名前で新規登録する
    func register(email: String, password: String, name: String, registerKey: String) async {
        isLoading = true
        errorMessage = nil

        do {
            let response = try await GRPCClient.shared.withClient { client in
                let authClient = Auth_AuthService.Client(wrapping: client)
                var request = Auth_RegisterByPasswordRequest()
                request.email = email
                request.password = password
                request.name = name
                request.registerKey = registerKey

                return try await authClient.registerByPassword(request)
            }

            guard
                KeychainStore.save(response.accessToken, for: .accessToken),
                KeychainStore.save(response.refreshToken, for: .refreshToken)
            else {
                errorMessage = "トークンの保存に失敗しました。再度お試しください。"
                isLoading = false
                return
            }
            isLoading = false
            isLoggedIn = true
        } catch {
            errorMessage = errorDescription(error, isAuthEndpoint: true)
            isLoading = false
        }
    }

    /// リフレッシュトークンを使ってアクセストークンを更新する
    /// アクセストークン期限切れ（UNAUTHENTICATED）発生時に呼び出す
    func refreshAccessToken() async throws {
        guard let refreshToken = KeychainStore.load(.refreshToken) else {
            // リフレッシュトークンがない場合はログアウトして再ログインを促す
            logout()
            return
        }

        let response = try await GRPCClient.shared.withClient { client in
            let authClient = Auth_AuthService.Client(wrapping: client)
            var request = Auth_RefreshAccessTokenRequest()
            request.refreshToken = refreshToken

            return try await authClient.refreshAccessToken(request)
        }

        guard
            KeychainStore.save(response.accessToken, for: .accessToken),
            KeychainStore.save(response.refreshToken, for: .refreshToken)
        else {
            logout()
            return
        }
    }

    /// ログアウトしてKeychainのトークンを削除する
    func logout() {
        KeychainStore.deleteAll()
        isLoggedIn = false
    }

    /// gRPCエラーを日本語メッセージに変換する
    /// - Parameter isAuthEndpoint: ログイン・登録エンドポイントかどうか（UNAUTHENTICATEDの解釈が変わる）
    func errorDescription(_ error: Error, isAuthEndpoint: Bool = false) -> String {
        if let rpcError = error as? RPCError {
            switch rpcError.code {
            case .unauthenticated:
                // ログイン・登録エンドポイントでは認証情報の誤り、それ以外はセッション期限切れ
                if isAuthEndpoint {
                    return "メールアドレスまたはパスワードが正しくありません"
                }
                return "セッションの有効期限が切れました。再ログインしてください。"

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
