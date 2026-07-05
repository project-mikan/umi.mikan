import Foundation
import Observation

/// アプリ起動時の初期読み込み状態。
///
/// スプラッシュ（起動アニメーション）は読み込み中インジケーターとして扱い、
/// 初期読み込みが完了したら即座に非表示にする。
@MainActor
@Observable
final class AppLaunchState {
    /// 初期読み込み中かどうか（trueの間だけスプラッシュを表示する）
    var isInitialLoading: Bool = true
}
