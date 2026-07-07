import SwiftUI

/// 日記詳細をハーフモーダルで表示する共通シート。
/// 全画面遷移は廃止し、日記詳細は常にこのシート経由で表示する。
///
/// 左右スワイプで items 内の前後の日記へ移動できる。
/// 前後の並びは呼び出し元が決める（検索=検索結果順、ホーム=今日/昨日/一昨日、月ごと=その月の各日付）。
struct DiaryDetailSheet: View {
    /// スワイプ切り替えのアニメーション（軽快さを出すため短めのスプリング）
    private static let slideAnimation: Animation = .spring(response: 0.32, dampingFraction: 0.86)
    /// 遷移確定とみなす移動量のしきい値（画面幅に対する比率）
    private static let commitRatio: CGFloat = 0.28
    /// フリックとみなす速度のしきい値（pt/秒）。速く払えば移動量が小さくても遷移する
    private static let flickVelocity: CGFloat = 380

    let items: [DiarySheetItem]
    let authViewModel: AuthViewModel
    let syncManager: SyncManager

    /// 現在表示中の日記のインデックス
    @State private var index: Int
    /// ドラッグ中の横方向の移動量（指の動きに追従させて手応えを出す）
    @State private var dragOffset: CGFloat = 0
    /// 切り替え時のスライド方向（新しい日記が入ってくる側の端）
    @State private var slideInEdge: Edge = .trailing

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
            GeometryReader { proxy in
                DiaryDetailView(
                    date: items[index].date,
                    authViewModel: authViewModel,
                    syncManager: syncManager,
                    highlightKeywords: items[index].highlightKeywords
                )
                // 日付が変わったらViewModelごと作り直す
                .id(items[index].id)
                // ドラッグ量ぶんだけ横へずらして指に追従させる
                .offset(x: dragOffset)
                // 指を離した後の切り替えをスライドで見せる（新旧が反対方向へ滑る）
                .transition(.asymmetric(
                    insertion: .move(edge: slideInEdge),
                    removal: .move(edge: slideInEdge == .trailing ? .leading : .trailing)
                ))
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .contentShape(Rectangle())
                .simultaneousGesture(swipeGesture(width: proxy.size.width))
            }
        }
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
    }

    /// 左右スワイプで前後の日記へ移動するジェスチャー。
    /// ドラッグ中は指に追従させ、離した時に移動量か速度がしきい値を超えていれば遷移する。
    /// 縦スクロールと誤反応しないよう、横方向が明確に大きい場合のみ追従する。
    private func swipeGesture(width: CGFloat) -> some Gesture {
        DragGesture(minimumDistance: 12)
            .onChanged { value in
                // 縦方向が優勢なドラッグ（スクロール）には追従しない
                guard abs(value.translation.width) > abs(value.translation.height) else { return }
                // 端の日記でそれ以上めくれない方向はゴムのように抵抗をつける
                let raw = value.translation.width
                dragOffset = (raw > 0 && index == 0) || (raw < 0 && index == items.count - 1)
                    ? raw / 3
                    : raw
            }
            .onEnded { value in
                let horizontal = value.translation.width
                let velocity = value.predictedEndTranslation.width - horizontal
                guard abs(horizontal) > abs(value.translation.height) else {
                    resetOffset()
                    return
                }
                // 移動量がしきい値を超える、または勢いよく払った場合に遷移する
                let shouldCommit = abs(horizontal) > width * Self.commitRatio
                    || abs(velocity) > Self.flickVelocity
                if shouldCommit, horizontal < 0 {
                    showNext()
                } else if shouldCommit, horizontal > 0 {
                    showPrevious()
                } else {
                    resetOffset()
                }
            }
    }

    /// ドラッグ量を0へ戻す（遷移しなかった場合）
    private func resetOffset() {
        withAnimation(Self.slideAnimation) { dragOffset = 0 }
    }

    /// 次の日記（配列の後ろ）へ移動する。新しい日記は右から入ってくる
    private func showNext() {
        guard index < items.count - 1 else {
            resetOffset()
            return
        }
        slideInEdge = .trailing
        withAnimation(Self.slideAnimation) {
            index += 1
            dragOffset = 0
        }
    }

    /// 前の日記（配列の前）へ移動する。新しい日記は左から入ってくる
    private func showPrevious() {
        guard index > 0 else {
            resetOffset()
            return
        }
        slideInEdge = .leading
        withAnimation(Self.slideAnimation) {
            index -= 1
            dragOffset = 0
        }
    }
}
