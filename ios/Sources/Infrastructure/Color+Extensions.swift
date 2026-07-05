import SwiftUI

/// フロントエンド（Tailwind CSS）と共通の色定義。
///
/// Tailwind CSS のデフォルトカラーパレットに合わせた値を定義し、
/// ライトモード・ダークモードで適切な色が使われるようにする。
extension Color {
    // MARK: - Tailwind カラーパレット

    /// Tailwind gray-900 / dark: gray-100 — 見出し・強調テキスト色
    static let twHeading = adaptive(light: Color(hex: "111827"), dark: Color(hex: "F3F4F6"))

    /// Tailwind gray-700 / dark: gray-300 — 本文テキスト色
    static let twBody = adaptive(light: Color(hex: "374151"), dark: Color(hex: "D1D5DB"))

    /// Tailwind gray-500 / dark: gray-400 — サブタイトルや補足テキスト色
    static let twSecondary = adaptive(light: Color(hex: "6B7280"), dark: Color(hex: "9CA3AF"))

    /// Tailwind blue-600 / dark: blue-400 — リンク・プライマリアクション色
    static let twBlue = adaptive(light: Color(hex: "2563EB"), dark: Color(hex: "60A5FA"))

    /// Tailwind red-600 / dark: red-400 — エラー・危険アクション色
    static let twRed = adaptive(light: Color(hex: "DC2626"), dark: Color(hex: "F87171"))

    /// Tailwind green-600 / dark: green-400 — 成功・保存完了表示色
    static let twGreen = adaptive(light: Color(hex: "16A34A"), dark: Color(hex: "4ADE80"))

    /// Tailwind indigo-500 / dark: indigo-400 — フォーカスリング・フォームアクセント色
    static let twIndigo = adaptive(light: Color(hex: "6366F1"), dark: Color(hex: "818CF8"))

    // MARK: - Initializer

    /// HEX文字列から Color を生成する。
    /// - Parameter hex: 6桁の16進数カラーコード（"#" なし）
    init(hex: String) {
        let scanner = Scanner(string: hex)
        var value: UInt64 = 0
        scanner.scanHexInt64(&value)
        let red = Double((value >> 16) & 0xFF) / 255.0
        let green = Double((value >> 8) & 0xFF) / 255.0
        let blue = Double(value & 0xFF) / 255.0
        self.init(red: red, green: green, blue: blue)
    }

    // MARK: - Private Helpers

    /// ライト・ダーク両対応の色を生成する
    private static func adaptive(light: Color, dark: Color) -> Color {
        Color(UIColor { traits in
            traits.userInterfaceStyle == .dark ? UIColor(dark) : UIColor(light)
        })
    }
}
