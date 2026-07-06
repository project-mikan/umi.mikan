import Connect
import Foundation

/// ConnectRPC API呼び出しの共通ヘルパー
enum APIHelper {
    /// アクセストークン期限切れ（UNAUTHENTICATED）時にリフレッシュしてリトライする。
    /// リフレッシュ自体が失敗した場合（リフレッシュトークンも期限切れ等）はリトライせず元のエラーを返す。
    @MainActor
    static func withTokenRefresh<T>(
        _ authViewModel: AuthViewModel,
        _ operation: () async -> ResponseMessage<T>
    ) async -> ResponseMessage<T> {
        let response = await operation()
        guard response.error?.code == .unauthenticated else { return response }
        do {
            try await authViewModel.refreshAccessToken()
        } catch {
            // リフレッシュ失敗（リフレッシュトークン期限切れ等）はログアウト済みのため再試行しない
            return response
        }
        return await operation()
    }

    /// ネットワーク起因のエラーかどうかを判定する。
    /// オフライン時は正常系として扱うため、エラーバナーを表示しない判断に使う。
    static func isNetworkError(_ error: ConnectError) -> Bool {
        error.code == .unavailable || error.code == .deadlineExceeded || error.code == .unknown
    }

    /// ConnectRPC エラーを日本語メッセージに変換する
    static func errorMessage(_ error: ConnectError) -> String {
        switch error.code {
        case .unauthenticated:
            return "セッションの有効期限が切れました。再ログインしてください。"

        case .notFound:
            return "日記が見つかりません"

        case .alreadyExists:
            return "この日付の日記は既に存在します"

        case .invalidArgument:
            return "入力内容を確認してください"

        case .failedPrecondition:
            return "この機能を利用するには設定が必要です"

        default:
            return "エラーが発生しました"
        }
    }
}
