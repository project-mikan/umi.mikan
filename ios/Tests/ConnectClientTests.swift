import Connect
import Foundation
import Testing
@testable import umi_mikan

@MainActor
struct ConnectClientTests {
    @Test("正常系: replaceHostで接続先を差し替えると、以降の通信は差し替え先へ向かう（到達不能な接続先ならネットワークエラーになる）")
    func replaceHost() async {
        ConnectClient.shared.replaceHost("http://127.0.0.1:9")
        defer { ConnectClient.shared.replaceHost(ConnectClient.defaultHost) }

        let client = User_UserServiceClient(client: ConnectClient.shared.protocolClient)
        let response = await client.getUserInfo(request: User_GetUserInfoRequest(), headers: [:])

        // 本番に繋がっていれば認証エラー等になり、ネットワークエラーにはならない
        #expect(response.error.map(APIHelper.isNetworkError) == true)
    }
}
