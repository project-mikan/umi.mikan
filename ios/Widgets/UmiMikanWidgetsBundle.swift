import SwiftUI
import WidgetKit

/// ウィジェット拡張のエントリーポイント
@main
struct UmiMikanWidgetsBundle: WidgetBundle {
    var body: some Widget {
        PendingDiaryLiveActivity()
    }
}
