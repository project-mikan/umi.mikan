import SwiftUI

/// 日記詳細をハーフモーダルで表示する共通シート。
/// 全画面遷移は廃止し、日記詳細は常にこのシート経由で表示する。
///
/// 左右スワイプで items 内の前後の日記へ移動できる。
/// 前後の並びは呼び出し元が決める（検索=検索結果順、ホーム=今日/昨日/一昨日、月ごと=その月の各日付）。
struct DiaryDetailSheet: View {
    let items: [DiarySheetItem]
    let authViewModel: AuthViewModel
    let syncManager: SyncManager

    /// 現在表示中の日記のインデックス
    @State private var index: Int
    /// スワイプ方向に応じた遷移アニメーションの起点
    @State private var pushEdge: Edge = .trailing

    // swiftlint:disable:next type_contents_order
    init(items: [DiarySheetItem], initialIndex: Int, authViewModel: AuthViewModel, syncManager: SyncManager) {
        self.items = items
        self.authViewModel = authViewModel
        self.syncManager = syncManager
        // 範囲外のインデックスが渡されても落ちないように丸める
        _index = State(initialValue: min(max(initialIndex, 0), max(items.count - 1, 0)))
    }

    var body: some View {
        NavigationStack {
            // NavigationStack 直下では .transition が反映されないため、
            // ZStack で包んでコンテナ側のアニメーションでスライド遷移させる
            ZStack {
                DiaryDetailView(
                    date: items[index].date,
                    authViewModel: authViewModel,
                    syncManager: syncManager,
                    highlightKeywords: items[index].highlightKeywords
                )
                // 日付が変わったらViewModelごと作り直してスライドアニメーションさせる
                .id(items[index].id)
                .transition(.push(from: pushEdge))
            }
            .animation(.easeInOut(duration: 0.25), value: index)
        }
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
        .simultaneousGesture(swipeGesture)
    }

    /// 左右スワイプで前後の日記へ移動するジェスチャー。
    /// 縦スクロールと誤反応しないよう、横方向が明確に大きい場合のみ反応する。
    private var swipeGesture: some Gesture {
        DragGesture(minimumDistance: 40)
            .onEnded { value in
                let horizontal = value.translation.width
                let vertical = value.translation.height
                guard abs(horizontal) > abs(vertical) * 1.5 else { return }
                if horizontal < 0 {
                    showNext()
                } else {
                    showPrevious()
                }
            }
    }

    /// 次の日記（配列の後ろ）へ移動する
    private func showNext() {
        guard index < items.count - 1 else { return }
        pushEdge = .trailing
        withAnimation(.easeInOut(duration: 0.25)) {
            index += 1
        }
    }

    /// 前の日記（配列の前）へ移動する
    private func showPrevious() {
        guard index > 0 else { return }
        pushEdge = .leading
        withAnimation(.easeInOut(duration: 0.25)) {
            index -= 1
        }
    }
}
