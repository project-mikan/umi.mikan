import Foundation

/// 日記詳細をハーフモーダル（シート）で表示するための項目。
/// 検索結果から開いた場合はハイライトするキーワードを保持する。
struct DiarySheetItem: Identifiable {
    let date: Diary_YMD
    /// 詳細画面でハイライトする検索キーワード（検索結果から開いた場合のみ）
    var highlightKeywords: [String] = []

    var id: String {
        LocalDiaryEntry.dateKey(date)
    }
}
