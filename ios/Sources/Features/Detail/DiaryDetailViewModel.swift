import Connect
import Foundation

/// 日付ごとの日記詳細・編集画面のViewModel。
///
/// オフライン対応のため、ローカルストアを優先して表示し、
/// 保存はローカルへ書き込んだ後にSyncManagerがサーバーへ同期する。
@MainActor
@Observable
final class DiaryDetailViewModel {
    let date: Diary_YMD
    var entry: Diary_DiaryEntry?
    var content: String = ""
    /// この日記の概要（チャンク一覧、RAGインデックス済みの場合のみ）
    var embeddingStatus: Diary_GetDiaryEmbeddingStatusResponse?
    var isLoading: Bool = false
    var isSaving: Bool = false
    var isSaved: Bool = false
    var errorMessage: String?

    private let authViewModel: AuthViewModel
    private let syncManager: SyncManager
    private let store: LocalDiaryStore
    /// 最後にストアから読み込んだ内容（ユーザーの入力途中の上書きを防ぐ）
    private var lastLoadedContent: String = ""
    /// isSaved を 2 秒後にリセットするタスク（ViewModel 破棄時にキャンセルできるよう保持する）
    private var savedResetTask: Task<Void, Never>?

    /// 編集中の本文に未保存の変更があるかどうか（フォーカスが外れた時の自動保存の判定に使う）
    var hasUnsavedChanges: Bool {
        content != lastLoadedContent
    }

    init(date: Diary_YMD, authViewModel: AuthViewModel, syncManager: SyncManager, store: LocalDiaryStore = .shared) {
        self.date = date
        self.authViewModel = authViewModel
        self.syncManager = syncManager
        self.store = store
    }

    /// ローカルを優先して表示し、サーバーから最新を取得して反映する
    func fetch() async {
        errorMessage = nil

        // ローカルストアから即座に表示する
        if let local = store.entry(for: date) {
            entry = local.toProto()
            content = local.content
            lastLoadedContent = local.content
        }
        isLoading = entry == nil

        // サーバーから最新を取得する
        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Diary_GetDiaryEntryRequest()
        request.date = date

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.getDiaryEntry(request: request, headers: ConnectClient.shared.headers())
        }

        if let error = response.error {
            // NOT_FOUND は未作成、ネットワークエラーはオフラインとして正常扱い
            if error.code != .notFound, !APIHelper.isNetworkError(error) {
                errorMessage = APIHelper.errorMessage(error)
            }
        } else if let message = response.message, message.hasEntry, !message.entry.id.isEmpty {
            store.applyServerEntry(message.entry)
            if let local = store.entry(for: date) {
                entry = local.toProto()
                // ユーザーが編集を始めていない場合のみテキストを更新する
                if content == lastLoadedContent {
                    content = local.content
                    lastLoadedContent = local.content
                }
            }
        }
        isLoading = false

        // 概要は取得できなくても画面表示に影響させない
        await fetchEmbeddingStatus()
    }

    /// 日記をローカルへ保存し、サーバーへの同期を試みる。オフラインでも保存は成功する。
    func save() async {
        isSaving = true
        errorMessage = nil
        defer { isSaving = false }

        store.saveLocalEdit(date: date, content: content)
        lastLoadedContent = content
        syncManager.refreshPendingCount()
        if let local = store.entry(for: date) {
            entry = local.toProto()
        }
        isSaved = true

        // 同期はバックグラウンドで実行し、保存操作自体は同期完了を待たない
        // （ネットワーク不調時に同期が長引いて保存操作がブロックされるのを防ぐため）
        Task { await syncManager.syncPending() }

        // 2秒後に保存済み表示をリセット（協調キャンセル可能なタスクとして保持する）
        savedResetTask?.cancel()
        savedResetTask = Task {
            do {
                try await Task.sleep(for: .seconds(2))
                isSaved = false
            } catch {
                // キャンセル時は何もしない
            }
        }
    }

    /// この日記の概要（チャンク一覧）を取得する（未生成・エラー時は非表示にするだけ）
    func fetchEmbeddingStatus() async {
        embeddingStatus = nil
        guard let entry, !entry.id.isEmpty else { return }

        let client = Diary_DiaryServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Diary_GetDiaryEmbeddingStatusRequest()
        request.diaryID = entry.id

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.getDiaryEmbeddingStatus(request: request, headers: ConnectClient.shared.headers())
        }

        guard let message = response.message, message.indexed, !message.chunkSummaries.isEmpty else { return }
        embeddingStatus = message
    }
}
