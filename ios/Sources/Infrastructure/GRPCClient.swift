import GRPCCore
import GRPCNIOTransportHTTP2TransportServices

/// gRPCチャンネルとサービスクライアントを管理するシングルトン
final class GRPCClient: Sendable {
    static let shared = GRPCClient()

    private let host = "umi-mikan-api.usuyuki.net"
    private let port = 443

    private init() {}

    /// gRPCコールに使うメタデータを生成する（Authorizationヘッダー付き）
    func metadata(accessToken: String? = nil) -> Metadata {
        var metadata = Metadata()
        if let token = accessToken ?? KeychainStore.load(.accessToken) {
            metadata.addString("Bearer \(token)", forKey: "authorization")
        }
        return metadata
    }

    /// 指定したクロージャ内でgRPCクライアントを使用する
    func withClient<T: Sendable>(
        _ body: @Sendable (GRPCCore.GRPCClient<HTTP2ClientTransport.TransportServices>) async throws -> T
    ) async throws -> T {
        let transport = try HTTP2ClientTransport.TransportServices(
            target: .dns(host: host, port: port),
            transportSecurity: .tls
        )
        let client = GRPCCore.GRPCClient(transport: transport)
        return try await withThrowingDiscardingTaskGroup { group in
            group.addTask { try await client.runConnections() }
            let result = try await body(client)
            client.beginGracefulShutdown()
            return result
        }
    }
}
