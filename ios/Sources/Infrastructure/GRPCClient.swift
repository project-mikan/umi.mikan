import GRPCCore
import GRPCNIOTransportHTTP2TransportServices

/// gRPCチャンネルとサービスクライアントを管理するシングルトン
///
/// コネクションはシングルトン生成時に一度だけ確立し、HTTP/2多重化で全RPCを共有する。
/// withClient(_:) は永続クライアントに対してRPCを実行するだけで、毎回TLSハンドシェイクを行わない。
final class GRPCClient: Sendable {
    static let shared = GRPCClient()

    private let host = "umi-mikan-api.usuyuki.net"
    private let port = 443

    /// 永続的なgRPCクライアント（コネクションはrunConnectionsタスクが維持）
    private let client: GRPCCore.GRPCClient<HTTP2ClientTransport.TransportServices>
    /// コネクション維持タスク（アプリ終了まで動き続ける）
    private let connectionTask: Task<Void, Never>

    private init() {
        // swiftlint:disable:next force_try
        let transport = try! HTTP2ClientTransport.TransportServices(
            target: .dns(host: host, port: port),
            transportSecurity: .tls
        )
        let grpcClient = GRPCCore.GRPCClient(transport: transport)
        client = grpcClient
        // バックグラウンドでコネクションを維持し続ける
        connectionTask = Task {
            try? await grpcClient.runConnections()
        }
    }

    /// gRPCコールに使うメタデータを生成する（Authorizationヘッダー付き）
    func metadata(accessToken: String? = nil) -> Metadata {
        var metadata = Metadata()
        if let token = accessToken ?? KeychainStore.load(.accessToken) {
            metadata.addString("Bearer \(token)", forKey: "authorization")
        }
        return metadata
    }

    /// 永続コネクションのgRPCクライアントを使ってRPCを実行する
    func withClient<T: Sendable>(
        _ body: @Sendable (GRPCCore.GRPCClient<HTTP2ClientTransport.TransportServices>) async throws -> T
    ) async throws -> T {
        try await body(client)
    }
}
