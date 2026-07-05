import ActivityKit
import Foundation

/// 未同期の日記件数を Live Activity として表示・更新するマネージャー。
/// 件数が1件以上で開始、変化で更新、0件になったら終了する。
@MainActor
final class LiveActivityManager {
    static let shared = LiveActivityManager()

    /// 表示中のLive Activity
    private var activity: Activity<DiaryActivityAttributes>?

    /// 未同期件数に応じてLive Activityを開始・更新・終了する
    func update(pendingCount: Int, isSyncing: Bool) {
        guard ActivityAuthorizationInfo().areActivitiesEnabled else { return }

        if pendingCount <= 0 {
            end()
            return
        }

        let content = ActivityContent(
            state: DiaryActivityAttributes.ContentState(pendingCount: pendingCount, isSyncing: isSyncing),
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

    /// Live Activityを即座に終了する
    func end() {
        guard let activity else { return }
        self.activity = nil
        let content = ActivityContent(
            state: DiaryActivityAttributes.ContentState(pendingCount: 0, isSyncing: false),
            staleDate: nil
        )
        Task { await activity.end(content, dismissalPolicy: .immediate) }
    }
}
