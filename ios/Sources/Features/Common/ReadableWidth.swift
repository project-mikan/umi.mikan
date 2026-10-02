import SwiftUI

/// iPad等の広い画面で本文の1行が長くなりすぎないよう、コンテンツ幅の上限を決める
enum ReadableWidth {
    /// regular幅のときのコンテンツ最大幅（pt）
    static let maxWidth: CGFloat = 720

    /// サイズクラスに応じたコンテンツ最大幅を返す。
    /// regular幅のときだけ上限を付け、compact幅・未確定（nil）のときは上限なし（nil）とする。
    /// iPadのSplit View等で狭くなった場合も compact になるため、端末種別ではなくサイズクラスで判定する。
    static func maxWidth(for sizeClass: UserInterfaceSizeClass?) -> CGFloat? {
        sizeClass == .regular ? maxWidth : nil
    }
}

/// regular幅のときにコンテンツを最大幅で中央寄せする修飾子
private struct ReadableWidthModifier: ViewModifier {
    @Environment(\.horizontalSizeClass)
    private var horizontalSizeClass

    func body(content: Content) -> some View {
        // サイズクラスで修飾子自体を付け外しするとView同一性が変わり、
        // Split Viewでのリサイズ時に入力中のフォーカス等が失われるため、常に同じ構造で幅だけ変える
        content
            .frame(maxWidth: ReadableWidth.maxWidth(for: horizontalSizeClass))
            .frame(maxWidth: .infinity)
    }
}

extension View {
    /// regular幅（iPad等）のときにコンテンツ幅へ上限を付けて中央寄せする
    func readableContentWidth() -> some View {
        modifier(ReadableWidthModifier())
    }
}
