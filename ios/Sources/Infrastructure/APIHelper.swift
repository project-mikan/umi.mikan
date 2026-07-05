import Connect
import Foundation

/// ConnectRPC API呼び出しの共通ヘルパー
enum APIHelper {
    /// アクセストークン期限切れ（UNAUTHENTICATED）時にリフレッシュしてリトライする
    @MainActor
    static func withTokenRefresh<T>(
        _ authViewModel: AuthViewModel,
        _ operation: () async -> ResponseMessage<T>
    ) async -> ResponseMessage<T> {
        let response = await operation()
        if response.error?.code == .unauthenticated {
            try? await authViewModel.refreshAccessToken()
            return await operation()
        }
        return response
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
