import SwiftUI

/// 設定ページ - ユーザー情報の表示・変更とログアウトを行う
struct SettingsView: View {
    @State private var viewModel: SettingsViewModel
    @State private var showLogoutConfirm = false

    // swiftlint:disable:next type_contents_order
    init(authViewModel: AuthViewModel) {
        _viewModel = State(initialValue: SettingsViewModel(authViewModel: authViewModel))
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                if viewModel.isLoading {
                    loadingView
                } else {
                    accountCard
                    logoutCard
                }
            }
            .padding(16)
        }
        .task {
            await viewModel.fetch()
        }
        .refreshable {
            await viewModel.fetch()
        }
        .confirmationDialog("ログアウトしますか？", isPresented: $showLogoutConfirm, titleVisibility: .visible) {
            Button("ログアウトする", role: .destructive) {
                viewModel.logout()
            }
        } message: {
            Text("端末に保存された日記データも削除されます")
        }
        .overlay(alignment: .bottom) {
            if let error = viewModel.errorMessage {
                errorBanner(message: error)
            }
        }
    }

    // MARK: - コンポーネント

    private var accountCard: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("アカウント")
                .font(.subheadline)
                .fontWeight(.semibold)
                .foregroundStyle(Color.twHeading)
            nameField
            emailField
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
    }

    /// なまえの編集フィールドと保存ボタン
    private var nameField: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("なまえ")
                .font(.caption)
                .foregroundStyle(Color.twSecondary)
            HStack(spacing: 8) {
                TextField("なまえ", text: $viewModel.userName)
                    .textFieldStyle(.roundedBorder)
                Button {
                    Task { await viewModel.updateUserName() }
                } label: {
                    if viewModel.isSavingName {
                        ProgressView().controlSize(.small)
                    } else if viewModel.nameSaved {
                        Image(systemName: "checkmark")
                    } else {
                        Text("保存")
                            .font(.caption)
                    }
                }
                .buttonStyle(.glassProminent)
                .disabled(viewModel.isSavingName || viewModel.userName.trimmingCharacters(in: .whitespaces).isEmpty)
            }
        }
    }

    /// メールアドレスの表示
    private var emailField: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("メールアドレス")
                .font(.caption)
                .foregroundStyle(Color.twSecondary)
            Text(viewModel.email)
                .font(.subheadline)
                .foregroundStyle(Color.twBody)
        }
    }

    private var logoutCard: some View {
        Button(role: .destructive) {
            showLogoutConfirm = true
        } label: {
            Label("ログアウト", systemImage: "rectangle.portrait.and.arrow.right")
                .frame(maxWidth: .infinity)
                .frame(height: 44)
        }
        .buttonStyle(.glass)
        .tint(.red)
    }

    private var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .controlSize(.large)
            Text("読み込み中...")
                .font(.subheadline)
                .foregroundStyle(Color.twSecondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 60)
    }

    private func errorBanner(message: String) -> some View {
        HStack(spacing: 8) {
            Image(systemName: "exclamationmark.circle.fill")
            Text(message)
                .font(.subheadline)
            Spacer()
            Button {
                viewModel.errorMessage = nil
            } label: {
                Image(systemName: "xmark")
                    .font(.caption)
            }
        }
        .padding(16)
        .background(.red.opacity(0.15))
        .foregroundStyle(Color.twRed)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .glassEffect(.regular.tint(.red), in: .rect(cornerRadius: 12))
        .padding(16)
    }
}
