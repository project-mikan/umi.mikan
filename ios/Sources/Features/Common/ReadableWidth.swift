import SwiftUI

/// iPad等の広い画面で1行が長くなりすぎないよう本文幅に上限を付ける
enum ReadableWidth {
    static let maxWidth: CGFloat = 720

    /// 端末種別ではなくサイズクラスで判定する（iPadのSplit View狭幅はcompactになるため）
    static func maxWidth(for sizeClass: UserInterfaceSizeClass?) -> CGFloat? {
        sizeClass == .regular ? maxWidth : nil
    }
}

private struct ReadableWidthModifier: ViewModifier {
    @Environment(\.horizontalSizeClass)
    private var horizontalSizeClass

    func body(content: Content) -> some View {
        // 修飾子を付け外しするとView同一性が変わりリサイズ時にフォーカスが失われるため、常に同じ構造にする
        content
            .frame(maxWidth: ReadableWidth.maxWidth(for: horizontalSizeClass))
            .frame(maxWidth: .infinity)
    }
}

extension View {
    func readableContentWidth() -> some View {
        modifier(ReadableWidthModifier())
    }
}
