import Connect
import Foundation

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

        let authClient = Auth_AuthServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Auth_LoginByPasswordRequest()
        request.email = email
        request.password = password

        let response = await authClient.loginByPassword(request: request, headers: ConnectClient.shared.headers())
        if let error = response.error {
            errorMessage = errorDescription(error, isAuthEndpoint: true)
            isLoading = false
            return
        }
        guard let message = response.message else {
            errorMessage = "レスポンスが空です"
            isLoading = false
            return
        }

        guard
            KeychainStore.save(message.accessToken, for: .accessToken),
            KeychainStore.save(message.refreshToken, for: .refreshToken)
        else {
            errorMessage = "トークンの保存に失敗しました。再度お試しください。"
            isLoading = false
            return
        }
        isLoading = false
        isLoggedIn = true
    }

    /// メールアドレス・パスワード・名前で新規登録する
    func register(email: String, password: String, name: String, registerKey: String) async {
        isLoading = true
        errorMessage = nil

        let authClient = Auth_AuthServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Auth_RegisterByPasswordRequest()
        request.email = email
        request.password = password
        request.name = name
        request.registerKey = registerKey

        let response = await authClient.registerByPassword(request: request, headers: ConnectClient.shared.headers())
        if let error = response.error {
            errorMessage = errorDescription(error, isAuthEndpoint: true)
            isLoading = false
            return
        }
        guard let message = response.message else {
            errorMessage = "レスポンスが空です"
            isLoading = false
            return
        }

        guard
            KeychainStore.save(message.accessToken, for: .accessToken),
            KeychainStore.save(message.refreshToken, for: .refreshToken)
        else {
            errorMessage = "トークンの保存に失敗しました。再度お試しください。"
            isLoading = false
            return
        }
        isLoading = false
        isLoggedIn = true
    }

    /// リフレッシュトークンを使ってアクセストークンを更新する
    /// アクセストークン期限切れ（UNAUTHENTICATED）発生時に呼び出す
    func refreshAccessToken() async throws {
        guard let refreshToken = KeychainStore.load(.refreshToken) else {
            // リフレッシュトークンがない場合はログアウトして再ログインを促す
            logout()
            return
        }

        let authClient = Auth_AuthServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Auth_RefreshAccessTokenRequest()
        request.refreshToken = refreshToken

        let response = await authClient.refreshAccessToken(request: request, headers: ConnectClient.shared.headers())
        if let error = response.error {
            logout()
            throw error
        }
        guard let message = response.message else {
            logout()
            return
        }

        // アクセストークンを更新する。
        // RefreshAccessToken レスポンスの refresh_token は空文字列の場合があるため、
        // 空のときは既存のリフレッシュトークンをそのまま維持する。
        var accessTokenSaved = KeychainStore.save(message.accessToken, for: .accessToken)
        if !message.refreshToken.isEmpty {
            accessTokenSaved = accessTokenSaved && KeychainStore.save(message.refreshToken, for: .refreshToken)
        }
        guard accessTokenSaved else {
            logout()
            return
        }
    }

    /// ログアウトしてKeychainのトークンを削除する
    func logout() {
        KeychainStore.deleteAll()
        isLoggedIn = false
    }

    /// ConnectRPC エラーを日本語メッセージに変換する
    /// - Parameter isAuthEndpoint: ログイン・登録エンドポイントかどうか（UNAUTHENTICATEDの解釈が変わる）
    func errorDescription(_ error: ConnectError, isAuthEndpoint: Bool = false) -> String {
        switch error.code {
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
}
