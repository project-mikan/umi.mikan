import ActivityKit
import Foundation

/// 未同期の日記件数と書きかけの日記を Live Activity として表示・更新するマネージャー。
///
/// - 未同期件数: SyncManager から通知され、1件以上で表示する
/// - 書きかけ: 編集途中でアプリがバックグラウンドへ移行した時に表示する
///
/// どちらも無くなったら Live Activity を終了する。
@MainActor
final class LiveActivityManager {
    static let shared = LiveActivityManager()

    /// 表示中のLive Activity
    private var activity: Activity<DiaryActivityAttributes>?
    /// 最後に通知された未同期件数
    private var pendingCount = 0
    /// 最後に通知された同期実行中フラグ
    private var isSyncing = false
    /// 書きかけの日記があるかどうか（バックグラウンド移行時にtrue、復帰時にfalse）
    private var hasDraft = false

    /// 未同期件数の変化を反映する（SyncManagerから呼ばれる）
    func update(pendingCount: Int, isSyncing: Bool) {
        self.pendingCount = pendingCount
        self.isSyncing = isSyncing
        refresh()
    }

    /// 書きかけの日記の有無を反映する。
    /// バックグラウンド移行直前（inactive）にtrue、フォアグラウンド復帰時にfalseを渡す。
    func setDraft(_ hasDraft: Bool) {
        guard self.hasDraft != hasDraft else { return }
        self.hasDraft = hasDraft
        refresh()
    }

    /// Live Activityを即座に終了する
    func end() {
        guard let activity else { return }
        self.activity = nil
        let content = ActivityContent(
            state: DiaryActivityAttributes.ContentState(pendingCount: 0, isSyncing: false, hasDraft: false),
            staleDate: nil
        )
        Task { await activity.end(content, dismissalPolicy: .immediate) }
    }

    // MARK: - Private

    /// 現在の状態に応じてLive Activityを開始・更新・終了する
    private func refresh() {
        guard ActivityAuthorizationInfo().areActivitiesEnabled else { return }

        if pendingCount <= 0, !hasDraft {
            end()
            return
        }

        let content = ActivityContent(
            state: DiaryActivityAttributes.ContentState(
                pendingCount: pendingCount,
                isSyncing: isSyncing,
                hasDraft: hasDraft
            ),
            staleDate: nil
        )

        // アプリ再起動時などは既存のアクティビティを引き継ぐ
        if activity == nil {
            activity = Activity<DiaryActivityAttributes>.activities.first
        }

        if let activity {
            Task { await activity.update(content) }
        } else {
            // 開始できない場合（設定で無効など）は静かに諦める
            activity = try? Activity.request(attributes: DiaryActivityAttributes(), content: content)
        }
    }
}
