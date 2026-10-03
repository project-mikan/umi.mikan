import Connect
import Foundation
import Synchronization

/// ConnectRPC クライアントを管理するシングルトン。
///
/// connect-swift の ProtocolClient を使って iOS から Cloudflare Tunnel 経由でバックエンドと通信する。
/// ConnectRPC は Content-Type: application/proto を使用するため Cloudflare にブロックされない。
/// 参照: adr/0012-ios-grpc-transport.md
final class ConnectClient: Sendable {
    static let shared = ConnectClient()

    static let defaultHost = "https://umi-mikan-api.usuyuki.net"

    /// スクリーンショットテストで接続先を差し替えるため Mutex で保持する
    private let client: Mutex<ProtocolClientInterface>

    /// ConnectRPC プロトコルクライアント（接続設定を保持）
    var protocolClient: ProtocolClientInterface {
        client.withLock { $0 }
    }

    private init() {
        client = Mutex(Self.makeProtocolClient(host: Self.defaultHost))
    }

    private static func makeProtocolClient(host: String) -> ProtocolClientInterface {
        ProtocolClient(
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
    }

    /// スクリーンショットテストで本番へ通信させない（実データ表示やログアウトを防ぐ）ために使う
    func replaceHost(_ host: String) {
        let newClient = Self.makeProtocolClient(host: host)
        client.withLock { $0 = newClient }
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
