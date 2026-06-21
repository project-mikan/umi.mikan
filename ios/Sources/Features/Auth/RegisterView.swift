import SwiftUI

/// 新規登録画面
struct RegisterView: View {
    @State private var email = ""
    @State private var password = ""
    @State private var name = ""
    @State private var registerKey = ""
    @Environment(\.dismiss) private var dismiss
    @Bindable var viewModel: AuthViewModel

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            VStack(spacing: 16) {
                TextField("名前", text: $name)
                    .textContentType(.name)
                    .textFieldStyle(.roundedBorder)

                TextField("メールアドレス", text: $email)
                    .textContentType(.emailAddress)
                    .keyboardType(.emailAddress)
                    .autocapitalization(.none)
                    .textFieldStyle(.roundedBorder)

                SecureField("パスワード", text: $password)
                    .textContentType(.newPassword)
                    .textFieldStyle(.roundedBorder)

                TextField("登録キー（任意）", text: $registerKey)
                    .autocapitalization(.none)
                    .textFieldStyle(.roundedBorder)
            }
            .padding(.horizontal)

            if let error = viewModel.errorMessage {
                Text(error)
                    .foregroundStyle(.red)
                    .font(.caption)
            }

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
                if viewModel.isLoading {
                    ProgressView()
                        .frame(maxWidth: .infinity)
                } else {
                    Text("登録")
                        .frame(maxWidth: .infinity)
                }
            }
            .buttonStyle(.borderedProminent)
            .padding(.horizontal)
            .disabled(viewModel.isLoading || email.isEmpty || password.isEmpty || name.isEmpty)

            Spacer()
        }
        .navigationTitle("アカウント作成")
        .navigationBarTitleDisplayMode(.inline)
    }
}
