import Connect
import Foundation

/// 設定ページのViewModel - ユーザー情報の取得・更新を管理する
@MainActor
@Observable
final class SettingsViewModel {
    var userName: String = ""
    var email: String = ""
    var isLoading: Bool = false
    var isSavingName: Bool = false
    var nameSaved: Bool = false
    var errorMessage: String?

    /// 「おもいで」通知のトグル状態
    var memoryNotificationEnabled: Bool = false
    /// 「おもいで」通知の発火時刻（時分のみ使用するDate）
    var memoryNotificationTime: Date

    private let authViewModel: AuthViewModel
    private let notificationManager: MemoryNotificationManager

    init(authViewModel: AuthViewModel, notificationManager: MemoryNotificationManager) {
        self.authViewModel = authViewModel
        self.notificationManager = notificationManager
        memoryNotificationEnabled = notificationManager.isEnabled
        memoryNotificationTime = Self.makeTime(
            hour: notificationManager.notificationHour,
            minute: notificationManager.notificationMinute
        )
    }

    /// hour/minuteからPicker表示用のDateを組み立てる
    private static func makeTime(hour: Int, minute: Int) -> Date {
        var components = DateComponents()
        components.hour = hour
        components.minute = minute
        return Calendar.current.date(from: components) ?? Date()
    }

    /// 通知時刻が変更された時の処理。ONの場合は新しい時刻で再スケジュールする。
    func setMemoryNotificationTime(_ date: Date) {
        memoryNotificationTime = date
        let components = Calendar.current.dateComponents([.hour, .minute], from: date)
        guard let hour = components.hour, let minute = components.minute else { return }
        notificationManager.updateNotificationTime(hour: hour, minute: minute)
    }

    /// システム側の通知許可状態を確認し、拒否されていた場合はトグル表示をOFFへ補正する
    func refreshNotificationAuthorizationState() async {
        guard notificationManager.isEnabled else {
            memoryNotificationEnabled = false
            return
        }
        let authorized = await notificationManager.isSystemAuthorized()
        if !authorized {
            notificationManager.disable()
        }
        memoryNotificationEnabled = notificationManager.isEnabled
    }

    /// 「おもいで」通知のトグルが変更された時の処理
    func setMemoryNotificationEnabled(_ enabled: Bool) async {
        if enabled {
            let granted = await notificationManager.enable()
            memoryNotificationEnabled = granted
        } else {
            notificationManager.disable()
            memoryNotificationEnabled = false
        }
    }

    /// ユーザー情報を取得する
    func fetch() async {
        isLoading = userName.isEmpty
        errorMessage = nil

        let client = User_UserServiceClient(client: ConnectClient.shared.protocolClient)
        let request = User_GetUserInfoRequest()

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.getUserInfo(request: request, headers: ConnectClient.shared.headers())
        }

        if let error = response.error {
            if !APIHelper.isNetworkError(error) {
                errorMessage = APIHelper.errorMessage(error)
            }
            isLoading = false
            return
        }
        userName = response.message?.name ?? ""
        email = response.message?.email ?? ""
        isLoading = false
    }

    /// ユーザー名を更新する
    func updateUserName() async {
        let trimmed = userName.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }

        isSavingName = true
        nameSaved = false
        errorMessage = nil
        defer { isSavingName = false }

        let client = User_UserServiceClient(client: ConnectClient.shared.protocolClient)
        var request = User_UpdateUserNameRequest()
        request.newName = trimmed

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.updateUserName(request: request, headers: ConnectClient.shared.headers())
        }

        if let error = response.error {
            errorMessage = APIHelper.errorMessage(error)
            return
        }
        nameSaved = true
        // 2秒後に保存済み表示をリセット
        Task {
            try? await Task.sleep(for: .seconds(2))
            nameSaved = false
        }
    }

    /// ログアウトする
    func logout() {
        authViewModel.logout()
    }
}
