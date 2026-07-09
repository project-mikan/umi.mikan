import Connect
import Foundation

/// ConnectRPC クライアントを管理するシングルトン。
///
/// connect-swift の ProtocolClient を使って iOS から Cloudflare Tunnel 経由でバックエンドと通信する。
/// ConnectRPC は Content-Type: application/proto を使用するため Cloudflare にブロックされない。
/// 参照: adr/0012-ios-grpc-transport.md
final class ConnectClient: Sendable {
    static let shared = ConnectClient()

    private let host = "https://umi-mikan-api.usuyuki.net"

    /// ConnectRPC プロトコルクライアント（接続設定を保持）
    let protocolClient: ProtocolClientInterface

    private init() {
        let client = ProtocolClient(
            httpClient: URLSessionHTTPClient(),
            config: ProtocolClientConfig(
                host: host,
                networkProtocol: .connect,
                codec: ProtoCodec(),
                // バックグラウンド放置後の復帰直後などネットワークが不安定な状況で
                // リクエストが長時間（デフォルトの約60秒）ハングするのを防ぐ
                timeout: 15
            )
        )
        protocolClient = client
    }

    /// Authorization ヘッダーを含む Headers を生成する。
    func headers(accessToken: String? = nil) -> Connect.Headers {
        var headers: Connect.Headers = [:]
        if let token = accessToken ?? KeychainStore.load(.accessToken) {
            headers["Authorization"] = ["Bearer \(token)"]
        }
        return headers
    }
}
