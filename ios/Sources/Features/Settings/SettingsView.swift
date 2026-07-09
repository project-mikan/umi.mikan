import SwiftUI

/// 設定ページ - ユーザー情報の表示・変更とログアウトを行う
struct SettingsView: View {
    @State private var viewModel: SettingsViewModel
    @State private var showLogoutConfirm = false

    // swiftlint:disable:next type_contents_order
    init(authViewModel: AuthViewModel) {
        _viewModel = State(initialValue: SettingsViewModel(
            authViewModel: authViewModel,
            notificationManager: MemoryNotificationManager.shared
        ))
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                if viewModel.isLoading {
                    loadingView
                } else {
                    accountCard
                    notificationCard
                    logoutCard
                }
            }
            .padding(16)
        }
        .task {
            await viewModel.fetch()
            await viewModel.refreshNotificationAuthorizationState()
        }
        .refreshable {
            await viewModel.fetch()
            await viewModel.refreshNotificationAuthorizationState()
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
                ErrorBannerView(message: error) { viewModel.errorMessage = nil }
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

    /// 「おもいで」通知のON/OFFトグル
    private var notificationCard: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text("通知")
                .font(.subheadline)
                .fontWeight(.semibold)
                .foregroundStyle(Color.twHeading)
            Toggle(isOn: Binding(
                get: { viewModel.memoryNotificationEnabled },
                set: { newValue in
                    Task { await viewModel.setMemoryNotificationEnabled(newValue) }
                }
            )) {
                VStack(alignment: .leading, spacing: 2) {
                    Text("おもいで 通知")
                        .font(.subheadline)
                        .foregroundStyle(Color.twBody)
                    Text("毎日決まった時刻に振り返りを通知します")
                        .font(.caption2)
                        .foregroundStyle(Color.twSecondary)
                }
            }
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
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
}
