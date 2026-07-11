import Connect
import Foundation
import Network
#if canImport(UIKit)
    import UIKit
#endif

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
    #if canImport(UIKit)
        /// syncPending 実行中のバックグラウンドタスクID。期限切れハンドラと defer の二重解放を防ぐため、
        /// 解放時は必ずこのプロパティ経由でIDを取り出し .invalid にリセットしてから endBackgroundTask する。
        private var backgroundTaskID: UIBackgroundTaskIdentifier = .invalid
    #endif
    /// バックグラウンド実行時間が期限切れになったかどうか。
    /// OSの期限切れハンドラから true にセットし、syncPending の逐次ループを次のエントリへ進める前に
    /// チェックすることで、アサーション失効後も新しいネットワークリクエストを送り続けるのを防ぐ。
    private var isBackgroundTaskExpiring = false

    init(authViewModel: AuthViewModel, store: LocalDiaryStore = .shared) {
        self.authViewModel = authViewModel
        self.store = store
        pendingCount = store.pendingEntries().count
        startMonitoring()
        startSyncRequestObserver()
        updateLiveActivity()
    }

    /// バックグラウンド実行時間を延長した状態で任意の非同期処理を実行する。
    /// アプリのバックグラウンド移行直後に起動される保存処理などが、
    /// Taskの初回実行機会を得る前にOSにサスペンドされて失われるのを防ぐために使う
    /// （SyncManager自身の同期処理と同じ保護をHomeView/DiaryDetailViewの
    /// バックグラウンド保存にも適用する目的の共通ヘルパー）。
    /// 注意: ここで保護した operation（ローカル保存）の内部から syncPending() が
    /// 別Taskで起動される場合、syncPending() は自身の beginBackgroundTask で
    /// 独立したアサーションを別途取得する（このメソッドのアサーションとは連動しない）。
    /// ローカル保存自体は同期的にすぐ完了するため実害はないが、
    /// 両者は意図的に独立した2つのアサーションであり、片方の期限切れがもう片方の
    /// 処理を中断させるわけではない点に注意する。
    @MainActor
    static func withBackgroundTask(_ operation: () async -> Void) async {
        let holder = BackgroundTaskHolder()
        #if canImport(UIKit)
            holder.taskID = UIApplication.shared.beginBackgroundTask(withName: "DiaryBackgroundSave") {
                // expirationHandler は任意のスレッドから呼ばれうるため、@MainActor へホップしてから解放する
                Task { @MainActor in
                    holder.end()
                }
            }
        #endif
        await operation()
        #if canImport(UIKit)
            holder.end()
        #endif
    }

    /// 同期待ちのエントリをすべてサーバーへ送信する。
    /// 保存操作からは detached Task で fire-and-forget 起動されるため、
    /// アプリがバックグラウンドへ移行した直後に呼ばれてもOSに即座にサスペンドされないよう
    /// バックグラウンド実行時間を延長するアサーションを取得してから同期する。
    func syncPending() async {
        guard !isSyncing else { return }
        isSyncing = true
        isBackgroundTaskExpiring = false
        updateLiveActivity()
        beginBackgroundTask()
        defer {
            isSyncing = false
            pendingCount = store.pendingEntries().count
            updateLiveActivity()
            endBackgroundTask()
        }

        for local in store.pendingEntries() {
            // バックグラウンド実行時間の期限が切れた後は、アサーションを持たない状態で
            // 新しいネットワークリクエストを送り続けないよう、残りのエントリを打ち切る
            // （次回のオンライン復帰時や次回保存時に再度 syncPending が呼ばれ再開される）
            guard !isBackgroundTaskExpiring else { break }
            await syncEntry(local)
        }
    }

    /// バックグラウンド実行時間の延長をリクエストする（UIKitが使えない環境では何もしない）
    private func beginBackgroundTask() {
        #if canImport(UIKit)
            backgroundTaskID = UIApplication.shared.beginBackgroundTask(withName: "DiarySync") { [weak self] in
                // expirationHandler は任意のスレッドから呼ばれうるため、@MainActor へホップしてから解放する
                Task { @MainActor in
                    self?.isBackgroundTaskExpiring = true
                    self?.endBackgroundTask()
                }
            }
        #endif
    }

    /// バックグラウンド実行時間の延長アサーションを解放する。
    /// 期限切れハンドラと defer の両方から呼ばれうるため、解放済みなら何もしないようIDをリセットする。
    private func endBackgroundTask() {
        #if canImport(UIKit)
            guard backgroundTaskID != .invalid else { return }
            let taskID = backgroundTaskID
            backgroundTaskID = .invalid
            UIApplication.shared.endBackgroundTask(taskID)
        #endif
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
                    await resolveConflictAndUpdate(local: local, pushedContent: pushedContent, date: date)
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

    /// 作成時にサーバー側に既存エントリがあった場合、IDを取得して更新で上書きする。
    /// pushedContent は syncEntry でスナップショットした内容を引き継ぎ、
    /// 同期中の再編集チェックが正しく機能するようにする。
    private func resolveConflictAndUpdate(local: LocalDiaryEntry, pushedContent: String, date: Diary_YMD) async {
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
        updateRequest.content = pushedContent
        updateRequest.date = date

        let updateResponse = await APIHelper.withTokenRefresh(authViewModel) {
            await client.updateDiaryEntry(request: updateRequest, headers: ConnectClient.shared.headers())
        }
        guard updateResponse.error == nil, let updated = updateResponse.message?.entry else { return }
        // pushedContent を渡すことで、同期中にユーザーが再編集した場合に needsSync が正しく維持される
        store.completeSync(dateKey: local.dateKey, pushedContent: pushedContent, serverEntry: updated)
    }
}

/// SyncManager.withBackgroundTask のバックグラウンドタスクIDを保持し、期限切れハンドラと
/// 通常完了パスの両方から呼ばれても二重解放しないよう管理する。
/// expirationHandler は @MainActor へホップしてから end() を呼ぶため、
/// このクラス自体はアクター分離しなくても競合しない。
private final class BackgroundTaskHolder: @unchecked Sendable {
    #if canImport(UIKit)
        var taskID: UIBackgroundTaskIdentifier = .invalid
    #endif

    func end() {
        #if canImport(UIKit)
            guard taskID != .invalid else { return }
            let endingID = taskID
            taskID = .invalid
            UIApplication.shared.endBackgroundTask(endingID)
        #endif
    }
}
