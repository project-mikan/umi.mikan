import SwiftUI

/// 新規登録画面 - Liquid Glassデザイン対応
struct RegisterView: View {
    @State private var email = ""
    @State private var password = ""
    @State private var name = ""
    @State private var registerKey = ""

    @Bindable var viewModel: AuthViewModel

    var body: some View {
        ZStack {
            registerBackground
            ScrollView {
                VStack(spacing: 32) {
                    Spacer().frame(height: 40)
                    registerForm
                    registerErrorMessage
                    registerButton
                    Spacer().frame(height: 60)
                }
            }
        }
        .navigationTitle("アカウント作成")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(.hidden, for: .navigationBar)
    }

    private var registerBackground: some View {
        LinearGradient(
            colors: [.purple.opacity(0.3), .pink.opacity(0.3)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
        .ignoresSafeArea()
    }

    private var registerForm: some View {
        GlassEffectContainer(spacing: 20.0) {
            VStack(spacing: 20) {
                registerNameField
                registerEmailField
                registerPasswordField
                registerKeyField
            }
            .padding(.horizontal, 24)
        }
    }

    private var registerNameField: some View {
        VStack(alignment: .leading, spacing: 8) {
            fieldLabel("名前")
            TextField("田中太郎", text: $name)
                .textContentType(.name)
                .glassTextField()
        }
    }

    private var registerEmailField: some View {
        VStack(alignment: .leading, spacing: 8) {
            fieldLabel("メールアドレス")
            TextField("mail@example.com", text: $email)
                .textContentType(.emailAddress)
                .keyboardType(.emailAddress)
                .autocapitalization(.none)
                .textInputAutocapitalization(.never)
                .glassTextField()
        }
    }

    private var registerPasswordField: some View {
        VStack(alignment: .leading, spacing: 8) {
            fieldLabel("パスワード")
            SecureField("••••••••", text: $password)
                .textContentType(.newPassword)
                .glassTextField()
        }
    }

    private var registerKeyField: some View {
        VStack(alignment: .leading, spacing: 8) {
            fieldLabel("登録キー（任意）")
            TextField("登録キー", text: $registerKey)
                .autocapitalization(.none)
                .textInputAutocapitalization(.never)
                .glassTextField()
        }
    }

    @ViewBuilder private var registerErrorMessage: some View {
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

    private var registerButton: some View {
        Button {
            Task {
                await viewModel.register(
                    email: email,
                    password: password,
                    name: name,
                    registerKey: registerKey
                )
                // isLoggedIn=trueになるとContentViewがMainViewに切り替えるため
                // dismissは不要（破棄済みNavigationStack上での呼び出しを避ける）
            }
        } label: {
            Group {
                if viewModel.isLoading {
                    ProgressView().tint(.white)
                } else {
                    Text("アカウントを作成").fontWeight(.semibold)
                }
            }
            .frame(maxWidth: .infinity)
            .frame(height: 52)
        }
        .buttonStyle(.glassProminent)
        .disabled(viewModel.isLoading || email.isEmpty || password.isEmpty || name.isEmpty)
        .padding(.horizontal, 24)
    }

    private func fieldLabel(_ text: String) -> some View {
        Text(text)
            .font(.caption)
            .foregroundStyle(.white.opacity(0.8))
            .padding(.leading, 4)
    }
}

/// ガラス風テキストフィールドのスタイルを適用するViewExtension
private extension View {
    func glassTextField() -> some View {
        padding()
            .background(.white.opacity(0.15))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
    }
}

#Preview {
    NavigationStack {
        RegisterView(viewModel: AuthViewModel())
    }
}
