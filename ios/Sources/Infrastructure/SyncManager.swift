import Connect
import Foundation
import Network

/// オフライン編集の同期を管理するマネージャー。
///
/// NWPathMonitor でネットワーク状態を監視し、オンライン復帰時に
/// LocalDiaryStore の同期待ちエントリをサーバーへ送信する。
@MainActor
@Observable
final class SyncManager {
    /// ネットワークに接続されているかどうか
    var isOnline: Bool = true
    /// 同期処理の実行中かどうか
    var isSyncing: Bool = false
    /// 同期待ちのエントリ数
    var pendingCount: Int = 0

    private let authViewModel: AuthViewModel
    private let store: LocalDiaryStore
    private let monitor = NWPathMonitor()

    init(authViewModel: AuthViewModel, store: LocalDiaryStore = .shared) {
        self.authViewModel = authViewModel
        self.store = store
        pendingCount = store.pendingEntries().count
        startMonitoring()
        startSyncRequestObserver()
        updateLiveActivity()
    }

    /// 同期待ちのエントリをすべてサーバーへ送信する
    func syncPending() async {
        guard !isSyncing else { return }
        isSyncing = true
        updateLiveActivity()
        defer {
            isSyncing = false
            pendingCount = store.pendingEntries().count
            updateLiveActivity()
        }

        for local in store.pendingEntries() {
            await syncEntry(local)
        }
    }

    /// 同期待ち件数を最新化する
    func refreshPendingCount() {
        pendingCount = store.pendingEntries().count
        updateLiveActivity()
    }

    // MARK: - Private

    /// ネットワーク監視を開始し、オンライン復帰時に同期を実行する
    private func startMonitoring() {
        monitor.pathUpdateHandler = { [weak self] path in
            let online = path.status == .satisfied
            Task { @MainActor [weak self] in
                guard let self else { return }
                let wasOffline = !isOnline
                isOnline = online
                // オフラインからオンラインに復帰したら同期する
                if online, wasOffline {
                    await syncPending()
                }
            }
        }
        monitor.start(queue: DispatchQueue(label: "net.usuyuki.umi-mikan.network-monitor"))
    }

    /// Live Activityの「今すぐ同期」ボタンからの同期要求を監視する
    private func startSyncRequestObserver() {
        NotificationCenter.default.addObserver(
            forName: .syncPendingDiariesRequested,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            Task { @MainActor [weak self] in
                await self?.syncPending()
            }
        }
    }

    /// 未同期件数と同期状態をLive Activityへ反映する
    private func updateLiveActivity() {
        LiveActivityManager.shared.update(pendingCount: pendingCount, isSyncing: isSyncing)
    }

    /// 1件のエントリをサーバーへ送信する
    private func syncEntry(_ local: LocalDiaryEntry) async {
        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
        var date = Diary_YMD()
        date.year = UInt32(local.year)
        date.month = UInt32(local.month)
        date.day = UInt32(local.day)
        let pushedContent = local.content

        if local.serverID.isEmpty {
            // 新規作成（既に存在する場合はサーバーIDを取得して更新に切り替える）
            var request = Diary_CreateDiaryEntryRequest()
            request.content = pushedContent
            request.date = date

            let response = await APIHelper.withTokenRefresh(authViewModel) {
                await client.createDiaryEntry(request: request, headers: ConnectClient.shared.headers())
            }
            if let error = response.error {
                if error.code == .alreadyExists {
                    await resolveConflictAndUpdate(local: local, date: date)
                }
                return
            }
            guard let entry = response.message?.entry else { return }
            store.completeSync(dateKey: local.dateKey, pushedContent: pushedContent, serverEntry: entry)
        } else {
            var request = Diary_UpdateDiaryEntryRequest()
            request.id = local.serverID
            request.content = pushedContent
            request.date = date

            let response = await APIHelper.withTokenRefresh(authViewModel) {
                await client.updateDiaryEntry(request: request, headers: ConnectClient.shared.headers())
            }
            guard response.error == nil, let entry = response.message?.entry else { return }
            store.completeSync(dateKey: local.dateKey, pushedContent: pushedContent, serverEntry: entry)
        }
    }

    /// 作成時にサーバー側に既存エントリがあった場合、IDを取得して更新で上書きする
    private func resolveConflictAndUpdate(local: LocalDiaryEntry, date: Diary_YMD) async {
        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
        var getRequest = Diary_GetDiaryEntryRequest()
        getRequest.date = date

        let getResponse = await APIHelper.withTokenRefresh(authViewModel) {
            await client.getDiaryEntry(request: getRequest, headers: ConnectClient.shared.headers())
        }
        guard
            getResponse.error == nil,
            let serverEntry = getResponse.message?.entry,
            !serverEntry.id.isEmpty
        else {
            return
        }

        var updateRequest = Diary_UpdateDiaryEntryRequest()
        updateRequest.id = serverEntry.id
        updateRequest.content = local.content
        updateRequest.date = date

        let updateResponse = await APIHelper.withTokenRefresh(authViewModel) {
            await client.updateDiaryEntry(request: updateRequest, headers: ConnectClient.shared.headers())
        }
        guard updateResponse.error == nil, let updated = updateResponse.message?.entry else { return }
        store.completeSync(dateKey: local.dateKey, pushedContent: local.content, serverEntry: updated)
    }
}
