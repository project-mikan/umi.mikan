import SwiftUI
import Testing
@testable import umi_mikan

struct ReadableWidthTests {
    struct MaxWidthCase {
        let name: String
        let sizeClass: UserInterfaceSizeClass?
        let expected: CGFloat?
    }

    @Test(
        "maxWidth(for:): regular幅のときだけ本文幅に上限が付く",
        arguments: [
            // iPad全画面・iPhone Pro Max横向きなど、1行が長くなりすぎる幅なので上限を付ける
            MaxWidthCase(name: "正常系: regular幅を渡すと上限幅を返す", sizeClass: .regular, expected: ReadableWidth.maxWidth),
            // iPhone縦向き・iPadのSplit View狭幅など、従来どおり全幅で表示する
            MaxWidthCase(name: "正常系: compact幅を渡すと上限なし(nil)を返す", sizeClass: .compact, expected: nil),
            // Preview等でサイズクラスが未確定の場合は、既存のiPhone表示を崩さないよう上限を付けない
            MaxWidthCase(name: "正常系: サイズクラス未確定(nil)を渡すと上限なし(nil)を返す", sizeClass: nil, expected: nil)
        ]
    )
    func maxWidth(testCase: MaxWidthCase) {
        #expect(ReadableWidth.maxWidth(for: testCase.sizeClass) == testCase.expected, "\(testCase.name)")
    }
}
