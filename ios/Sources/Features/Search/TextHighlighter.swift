import SwiftUI

/// 検索キーワードのハイライト処理を提供するユーティリティ
enum TextHighlighter {
    /// テキスト全体のキーワード出現箇所をハイライトした AttributedString を返す。
    /// 改行はそのまま保持されるため、日記詳細の全文表示に使える。
    static func highlight(text: String, keywords: [String]) -> AttributedString {
        applyHighlights(to: AttributedString(text), keywords: normalizedKeywords(keywords))
    }

    /// キーワード検索結果のスニペットをハイライト付きで生成する。
    ///
    /// フロントエンドと同様に、最初のマッチ位置がスニペットの先頭30文字付近に
    /// 来るよう150文字の窓で切り出し、全キーワードの出現箇所をハイライトする。
    static func snippet(content: String, keywords: [String]) -> AttributedString {
        // 改行を空白に置換して1行化し、連続スペースを正規化する
        let normalized = content
            .replacingOccurrences(of: "\r\n", with: " ")
            .replacingOccurrences(of: "\r", with: " ")
            .replacingOccurrences(of: "\n", with: " ")
            .replacingOccurrences(of: "  +", with: " ", options: .regularExpression)
            .trimmingCharacters(in: .whitespaces)

        let window = 150
        let prefixLength = 30
        let activeKeywords = normalizedKeywords(keywords)

        // 全キーワードから最初のマッチ位置を検索する（大文字小文字無視）
        var firstMatchIndex: String.Index?
        for kw in activeKeywords {
            if let range = normalized.range(of: kw, options: .caseInsensitive) {
                if firstMatchIndex == nil || range.lowerBound < firstMatchIndex! {
                    firstMatchIndex = range.lowerBound
                }
            }
        }

        // マッチ位置に応じてスニペットを切り出す
        var excerpt: String
        var prefix = ""
        var suffix = ""
        let matchOffset = firstMatchIndex.map { normalized.distance(from: normalized.startIndex, to: $0) } ?? -1

        if matchOffset == -1 || matchOffset < prefixLength {
            excerpt = String(normalized.prefix(window))
            if normalized.count > window { suffix = "..." }
        } else {
            let start = max(0, matchOffset - prefixLength)
            let startIndex = normalized.index(normalized.startIndex, offsetBy: start)
            let endOffset = min(normalized.count, start + window)
            let endIndex = normalized.index(normalized.startIndex, offsetBy: endOffset)
            excerpt = String(normalized[startIndex ..< endIndex])
            if start > 0 { prefix = "..." }
            if endOffset < normalized.count { suffix = "..." }
        }

        return applyHighlights(to: AttributedString(prefix + excerpt + suffix), keywords: activeKeywords)
    }

    /// 複数行テキストの中で最初にキーワードが出現する行番号を返す（マッチなしは nil）。
    /// ハイライト表示で最初のマッチ行まで自動スクロールするために使う。
    static func firstMatchLineIndex(lines: [String], keywords: [String]) -> Int? {
        let activeKeywords = normalizedKeywords(keywords)
        guard !activeKeywords.isEmpty else { return nil }
        for (index, line) in lines.enumerated() {
            for kw in activeKeywords where line.range(of: kw, options: .caseInsensitive) != nil {
                return index
            }
        }
        return nil
    }

    // MARK: - Private

    /// 前後の空白を除去し、空のキーワードを除外する
    private static func normalizedKeywords(_ keywords: [String]) -> [String] {
        keywords
            .map { $0.trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }
    }

    /// AttributedString 内のキーワード出現箇所に背景色を設定する（大文字小文字無視）
    private static func applyHighlights(to text: AttributedString, keywords: [String]) -> AttributedString {
        var attributed = text
        for kw in keywords {
            var searchStart = attributed.startIndex
            while let range = attributed[searchStart...].range(of: kw, options: .caseInsensitive) {
                attributed[range].backgroundColor = .yellow.opacity(0.5)
                searchStart = range.upperBound
            }
        }
        return attributed
    }
}
