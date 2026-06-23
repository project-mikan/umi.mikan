import SwiftUI

/// ログイン画面 - Liquid Glassデザイン対応
struct LoginView: View {
    @State private var email = ""
    @State private var password = ""
    @State private var showRegister = false
    @Bindable var viewModel: AuthViewModel

    var body: some View {
        NavigationStack {
            loginContent
                .navigationDestination(isPresented: $showRegister) {
                    RegisterView(viewModel: viewModel)
                }
                .onAppear {
                    // 登録画面から戻った際に前のエラーをクリアする
                    viewModel.errorMessage = nil
                }
        }
    }

    private var loginContent: some View {
        ZStack {
            loginBackground
            ScrollView {
                VStack(spacing: 32) {
                    Spacer().frame(height: 60)
                    loginLogo
                    loginForm
                    loginErrorMessage
                    loginButtons
                    Spacer().frame(height: 60)
                }
            }
        }
    }

    private var loginBackground: some View {
        LinearGradient(
            colors: [.blue.opacity(0.3), .purple.opacity(0.3)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
        .ignoresSafeArea()
    }

    private var loginLogo: some View {
        VStack(spacing: 16) {
            Image(systemName: "drop.circle.fill")
                .font(.system(size: 80))
                .foregroundStyle(
                    .linearGradient(
                        colors: [.blue, .cyan],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
                .shadow(color: .blue.opacity(0.5), radius: 10)

            Text("umi.mikan")
                .font(.system(size: 48, weight: .bold, design: .rounded))
                .foregroundStyle(.white)
        }
        .padding(.horizontal, 32)
        .padding(.vertical, 20)
        .glassEffect(.regular.tint(.blue).interactive())
    }

    private var loginForm: some View {
        GlassEffectContainer(spacing: 20.0) {
            VStack(spacing: 20) {
                labeledTextField("メールアドレス", placeholder: "mail@example.com", text: $email)
                    .textContentType(.emailAddress)
                    .keyboardType(.emailAddress)
                    .autocapitalization(.none)
                    .textInputAutocapitalization(.never)

                labeledSecureField("パスワード", placeholder: "••••••••", text: $password)
                    .textContentType(.password)
            }
            .padding(.horizontal, 24)
        }
    }

    @ViewBuilder private var loginErrorMessage: some View {
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
    }

    private var loginButtons: some View {
        VStack(spacing: 16) {
            Button {
                Task { await viewModel.login(email: email, password: password) }
            } label: {
                loginButtonLabel
            }
            .buttonStyle(.glassProminent)
            .disabled(viewModel.isLoading || email.isEmpty || password.isEmpty)

            Button { showRegister = true } label: {
                Text("アカウントを作成")
                    .fontWeight(.medium)
                    .frame(maxWidth: .infinity)
                    .frame(height: 48)
            }
            .buttonStyle(.glass)
        }
        .padding(.horizontal, 24)
    }

    private var loginButtonLabel: some View {
        Group {
            if viewModel.isLoading {
                ProgressView().tint(.white)
            } else {
                Text("ログイン").fontWeight(.semibold)
            }
        }
        .frame(maxWidth: .infinity)
        .frame(height: 52)
    }
}

/// ラベル付きテキストフィールド
private func labeledTextField(
    _ label: String,
    placeholder: String,
    text: Binding<String>
) -> some View {
    VStack(alignment: .leading, spacing: 8) {
        Text(label)
            .font(.caption)
            .foregroundStyle(.white.opacity(0.8))
            .padding(.leading, 4)

        TextField(placeholder, text: text)
            .padding()
            .background(.white.opacity(0.15))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
    }
}

/// ラベル付きセキュアフィールド
private func labeledSecureField(
    _ label: String,
    placeholder: String,
    text: Binding<String>
) -> some View {
    VStack(alignment: .leading, spacing: 8) {
        Text(label)
            .font(.caption)
            .foregroundStyle(.white.opacity(0.8))
            .padding(.leading, 4)

        SecureField(placeholder, text: text)
            .padding()
            .background(.white.opacity(0.15))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
    }
}

#Preview {
    LoginView(viewModel: AuthViewModel())
}
