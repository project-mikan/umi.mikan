import SwiftUI

/// scenePhase の変化に応じて、書きかけ（未保存の編集）をバックグラウンド移行時に自動保存し、
/// Live Activityの「書きかけ」表示を制御する共通ロジック。
///
/// HomeView（3枚のカードをまとめて保存）と DiaryDetailView（1件のみ保存）は
/// 「未保存かどうかの判定」と「実際に保存する処理」だけが異なり、
/// .inactive/.background/.active の3ケースの制御フロー自体は同一のため、
/// この共通実装を1箇所にまとめて挙動の乖離を防ぐ。
@MainActor
protocol DraftAutoSaving {
    /// 現在、未保存の編集（書きかけ）があるかどうか
    var hasDraftInProgress: Bool { get }
    /// 書きかけの内容を実際に保存する（保存対象が複数ある場合はすべて保存してから返ること）
    func performDraftSave() async
}

extension DraftAutoSaving {
    /// アプリのフォアグラウンド状態の変化に応じて、書きかけの保存とLive Activityを制御する
    func handleDraftScenePhase(_ phase: ScenePhase) {
        switch phase {
        case .inactive:
            // Live Activityの開始はフォアグラウンド中しかできないため、
            // バックグラウンド移行直前の inactive の時点で開始する
            if hasDraftInProgress {
                LiveActivityManager.shared.setDraft(true)
            }

        case .background:
            // 書きかけを失わないようにローカルへ自動保存し、保存完了後に書きかけフラグを解除する。
            // バックグラウンド移行直後にTaskがスケジュールされる前にOSへサスペンドされるのを防ぐため、
            // バックグラウンド実行時間を延長してから保存する
            if hasDraftInProgress {
                Task {
                    await SyncManager.withBackgroundTask {
                        await performDraftSave()
                    }
                    LiveActivityManager.shared.setDraft(false)
                }
            }

        case .active:
            // フォアグラウンド復帰したら書きかけのLive Activityを終了する
            LiveActivityManager.shared.setDraft(false)

        @unknown default:
            break
        }
    }
}
