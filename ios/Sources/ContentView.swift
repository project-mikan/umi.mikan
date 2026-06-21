import SwiftUI

/// 認証状態に応じてログイン画面とメイン画面を切り替えるルートビュー
struct ContentView: View {
    @State private var authViewModel = AuthViewModel()

    var body: some View {
        if authViewModel.isLoggedIn {
            // TODO: 日記一覧画面に置き換える
            Text("ログイン済み")
                .toolbar {
                    ToolbarItem(placement: .topBarTrailing) {
                        Button("ログアウト") {
                            authViewModel.logout()
                        }
                    }
                }
        } else {
            LoginView(viewModel: authViewModel)
        }
    }
}
