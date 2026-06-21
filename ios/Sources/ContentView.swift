import SwiftUI

/// 認証状態に応じてログイン画面とメイン画面を切り替えるルートビュー
struct ContentView: View {
    @State private var authViewModel = AuthViewModel()

    var body: some View {
        if authViewModel.isLoggedIn {
            MainView(authViewModel: authViewModel)
        } else {
            LoginView(viewModel: authViewModel)
        }
    }
}

/// メイン画面 - Liquid Glassデザイン対応
struct MainView: View {
    @Bindable var authViewModel: AuthViewModel

    var body: some View {
        NavigationStack {
            ZStack {
                // 背景のグラデーション
                LinearGradient(
                    colors: [.blue.opacity(0.2), .cyan.opacity(0.2)],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
                .ignoresSafeArea()

                VStack(spacing: 32) {
                    Spacer()

                    // ログイン成功メッセージ
                    VStack(spacing: 16) {
                        Image(systemName: "checkmark.circle.fill")
                            .font(.system(size: 80))
                            .foregroundStyle(.green)
                            .glassEffect(.regular.tint(.green).interactive())

                        Text("ログイン成功")
                            .font(.title)
                            .fontWeight(.bold)
                            .foregroundStyle(.white)

                        Text("日記一覧画面は準備中です")
                            .font(.body)
                            .foregroundStyle(.white.opacity(0.8))
                    }
                    .padding(40)
                    .glassEffect(.regular.tint(.blue), in: .rect(cornerRadius: 24))

                    Spacer()
                }
                .padding()
            }
            .navigationTitle("umi.mikan")
            .navigationBarTitleDisplayMode(.inline)
            .toolbarBackground(.hidden, for: .navigationBar)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        authViewModel.logout()
                    } label: {
                        Label("ログアウト", systemImage: "rectangle.portrait.and.arrow.right")
                            .foregroundStyle(.white)
                    }
                    .buttonStyle(.glass)
                }
            }
        }
    }
}
