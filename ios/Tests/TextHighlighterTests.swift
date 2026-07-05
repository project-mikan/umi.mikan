import Foundation
import SwiftUI
import Testing
@testable import umi_mikan

/// TextHighlighterのテスト
struct TextHighlighterTests {
    /// highlight のテーブル駆動テスト用ケース
    struct HighlightCase: Sendable {
        let name: String
        let text: String
        let keywords: [String]
        let expected: [String]
    }

    /// AttributedString から背景色が設定された部分文字列の一覧を取り出す
    private func highlightedSubstrings(_ text: AttributedString) -> [String] {
        text.runs.compactMap { run in
            run.backgroundColor != nil ? String(text[run.range].characters) : nil
        }
    }

    // MARK: - highlight

    @Test(
        "highlight: キーワード出現箇所のみハイライトされる",
        arguments: [
            HighlightCase(
                name: "正常系: キーワードが1回出現するとその箇所がハイライトされる",
                text: "今日は海へ行った",
                keywords: ["海"],
                expected: ["海"]
            ),
            HighlightCase(
                name: "正常系: キーワードが複数回出現すると全てハイライトされる",
                text: "海が好き。海は広い。",
                keywords: ["海"],
                expected: ["海", "海"]
            ),
            HighlightCase(
                name: "正常系: 複数キーワードがそれぞれハイライトされる",
                text: "みかんと海の写真",
                keywords: ["海", "みかん"],
                expected: ["みかん", "海"]
            ),
            HighlightCase(
                name: "正常系: 英字キーワードは大文字小文字を無視してハイライトされる",
                text: "Swift と swift の話",
                keywords: ["swift"],
                expected: ["Swift", "swift"]
            ),
            HighlightCase(
                name: "正常系: キーワードが出現しない場合はハイライトされない",
                text: "今日は山へ行った",
                keywords: ["海"],
                expected: []
            ),
            HighlightCase(
                name: "正常系: 空・空白のみのキーワードは無視される",
                text: "今日は海へ行った",
                keywords: ["", "  "],
                expected: []
            ),
            HighlightCase(
                name: "正常系: 改行を含むテキストでもハイライトされる",
                text: "1行目\n海に行った\n3行目",
                keywords: ["海"],
                expected: ["海"]
            )
        ]
    )
    func highlight(testCase: HighlightCase) {
        let result = TextHighlighter.highlight(text: testCase.text, keywords: testCase.keywords)
        #expect(highlightedSubstrings(result) == testCase.expected, "\(testCase.name)")
        // ハイライトしてもテキスト自体は変化しない
        #expect(String(result.characters) == testCase.text, "\(testCase.name): テキストが変化しない")
    }

    // MARK: - snippet

    @Test("正常系: マッチが先頭付近にある場合は先頭から150文字のスニペットになる")
    func snippetMatchNearHead() {
        let content = "海に行った。" + String(repeating: "あ", count: 200)
        let result = TextHighlighter.snippet(content: content, keywords: ["海"])
        let text = String(result.characters)
        #expect(text.hasPrefix("海に行った。"))
        #expect(text.hasSuffix("..."))
        #expect(highlightedSubstrings(result) == ["海"])
    }

    @Test("正常系: マッチが後方にある場合はマッチ位置の30文字前から切り出され前後に...が付く")
    func snippetMatchInMiddle() {
        let content = String(repeating: "あ", count: 100) + "海に行った" + String(repeating: "い", count: 200)
        let result = TextHighlighter.snippet(content: content, keywords: ["海"])
        let text = String(result.characters)
        #expect(text.hasPrefix("..."))
        #expect(text.hasSuffix("..."))
        #expect(text.contains("海に行った"))
        #expect(highlightedSubstrings(result) == ["海"])
    }

    @Test("正常系: マッチしない場合は先頭150文字のスニペットになりハイライトされない")
    func snippetNoMatch() {
        let content = String(repeating: "あ", count: 300)
        let result = TextHighlighter.snippet(content: content, keywords: ["海"])
        let text = String(result.characters)
        #expect(text == String(repeating: "あ", count: 150) + "...")
        #expect(highlightedSubstrings(result).isEmpty)
    }

    @Test("正常系: 改行は空白に正規化されて1行のスニペットになる")
    func snippetNormalizesNewlines() {
        let content = "1行目\r\n2行目\rに海がある\n3行目"
        let result = TextHighlighter.snippet(content: content, keywords: ["海"])
        let text = String(result.characters)
        #expect(!text.contains("\n"))
        #expect(!text.contains("\r"))
        #expect(highlightedSubstrings(result) == ["海"])
    }
}
