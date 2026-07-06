import ActivityKit
import Foundation

/// 未同期の日記を知らせるLive Activityの属性。
/// アプリ本体とウィジェット拡張の両方でコンパイルされる共有コード。
struct DiaryActivityAttributes: ActivityAttributes {
    /// 動的に更新される状態
    struct ContentState: Codable, Hashable {
        /// 未同期の日記の件数
        var pendingCount: Int
        /// 同期処理を実行中かどうか
        var isSyncing: Bool
    }
}
