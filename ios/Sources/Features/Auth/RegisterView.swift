import SwiftUI

/// 新規登録画面 - Liquid Glassデザイン対応
struct RegisterView: View {
    @State private var email = ""
    @State private var password = ""
    @State private var name = ""
    @State private var registerKey = ""
    @Environment(\.dismiss) private var dismiss
    @Bindable var viewModel: AuthViewModel

    var body: some View {
        ZStack {
            // 背景のグラデーション
            LinearGradient(
                colors: [.purple.opacity(0.3), .pink.opacity(0.3)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            ScrollView {
                VStack(spacing: 32) {
                    Spacer()
                        .frame(height: 40)

                    // 登録フォーム - GlassEffectContainer使用
                    GlassEffectContainer(spacing: 20.0) {
                        VStack(spacing: 20) {
                            // 名前入力フィールド
                            VStack(alignment: .leading, spacing: 8) {
                                Text("名前")
                                    .font(.caption)
                                    .foregroundStyle(.white.opacity(0.8))
                                    .padding(.leading, 4)

                                TextField("田中太郎", text: $name)
                                    .textContentType(.name)
                                    .padding()
                                    .background(.white.opacity(0.15))
                                    .clipShape(RoundedRectangle(cornerRadius: 12))
                                    .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
                            }

                            // メールアドレス入力フィールド
                            VStack(alignment: .leading, spacing: 8) {
                                Text("メールアドレス")
                                    .font(.caption)
                                    .foregroundStyle(.white.opacity(0.8))
                                    .padding(.leading, 4)

                                TextField("mail@example.com", text: $email)
                                    .textContentType(.emailAddress)
                                    .keyboardType(.emailAddress)
                                    .autocapitalization(.none)
                                    .textInputAutocapitalization(.never)
                                    .padding()
                                    .background(.white.opacity(0.15))
                                    .clipShape(RoundedRectangle(cornerRadius: 12))
                                    .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
                            }

                            // パスワード入力フィールド
                            VStack(alignment: .leading, spacing: 8) {
                                Text("パスワード")
                                    .font(.caption)
                                    .foregroundStyle(.white.opacity(0.8))
                                    .padding(.leading, 4)

                                SecureField("••••••••", text: $password)
                                    .textContentType(.newPassword)
                                    .padding()
                                    .background(.white.opacity(0.15))
                                    .clipShape(RoundedRectangle(cornerRadius: 12))
                                    .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
                            }

                            // 登録キー入力フィールド
                            VStack(alignment: .leading, spacing: 8) {
                                Text("登録キー（任意）")
                                    .font(.caption)
                                    .foregroundStyle(.white.opacity(0.8))
                                    .padding(.leading, 4)

                                TextField("登録キー", text: $registerKey)
                                    .autocapitalization(.none)
                                    .textInputAutocapitalization(.never)
                                    .padding()
                                    .background(.white.opacity(0.15))
                                    .clipShape(RoundedRectangle(cornerRadius: 12))
                                    .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
                            }
                        }
                        .padding(.horizontal, 24)
                    }

                    // エラーメッセージ
                    if let error = viewModel.errorMessage {
                        Text(error)
                            .foregroundStyle(.red)
                            .font(.callout)
                            .padding(.horizontal, 24)
                            .padding(.vertical, 12)
                            .background(.red.opacity(0.15))
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                            .glassEffect(.regular.tint(.red), in: .rect(cornerRadius: 12))
                            .padding(.horizontal, 24)
                    }

                    // 登録ボタン
                    Button {
                        Task {
                            await viewModel.register(
                                email: email,
                                password: password,
                                name: name,
                                registerKey: registerKey
                            )
                            if viewModel.isLoggedIn { dismiss() }
                        }
                    } label: {
                        Group {
                            if viewModel.isLoading {
                                ProgressView()
                                    .tint(.white)
                            } else {
                                Text("アカウントを作成")
                                    .fontWeight(.semibold)
                            }
                        }
                        .frame(maxWidth: .infinity)
                        .frame(height: 52)
                    }
                    .buttonStyle(.glassProminent)
                    .disabled(viewModel.isLoading || email.isEmpty || password.isEmpty || name.isEmpty)
                    .padding(.horizontal, 24)

                    Spacer()
                        .frame(height: 60)
                }
            }
        }
        .navigationTitle("アカウント作成")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(.hidden, for: .navigationBar)
    }
}

#Preview {
    NavigationStack {
        RegisterView(viewModel: AuthViewModel())
    }
}
