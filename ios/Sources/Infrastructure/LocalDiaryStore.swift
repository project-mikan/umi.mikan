import Foundation

/// 端末に保存する日記エントリ
struct LocalDiaryEntry: Codable, Equatable {
    /// サーバー側のID（未同期の新規作成時は空文字列）
    var serverID: String
    var year: Int
    var month: Int
    var day: Int
    var content: String
    /// 最終更新日時（Unix秒）
    var updatedAt: Int64
    /// サーバーへの同期が必要かどうか（オフライン編集時にtrue）
    var needsSync: Bool

    /// ストアのキーとなる日付文字列（"YYYY-MM-DD"）
    var dateKey: String {
        Self.dateKey(year: year, month: month, day: day)
    }

    /// 日付から dateKey を生成する
    static func dateKey(year: Int, month: Int, day: Int) -> String {
        String(format: "%04d-%02d-%02d", year, month, day)
    }

    /// Diary_YMD から dateKey を生成する
    static func dateKey(_ date: Diary_YMD) -> String {
        dateKey(year: Int(date.year), month: Int(date.month), day: Int(date.day))
    }

    /// 表示用の Diary_DiaryEntry に変換する
    func toProto() -> Diary_DiaryEntry {
        var entry = Diary_DiaryEntry()
        entry.id = serverID
        var ymd = Diary_YMD()
        ymd.year = UInt32(year)
        ymd.month = UInt32(month)
        ymd.day = UInt32(day)
        entry.date = ymd
        entry.content = content
        entry.updatedAt = updatedAt
        return entry
    }
}

/// 日記のローカル永続化ストア（オフライン対応）
///
/// 日記を端末内のJSONファイルに保存し、オフラインでも閲覧・編集できるようにする。
/// オフライン編集は needsSync フラグで管理し、SyncManager がオンライン復帰時にサーバーへ送信する。
@MainActor
final class LocalDiaryStore {
    static let shared = LocalDiaryStore()

    /// dateKey（"YYYY-MM-DD"）をキーとしたエントリのマップ
    private var entries: [String: LocalDiaryEntry] = [:]
    private let fileURL: URL

    /// テスト用に保存先ファイルを指定できるイニシャライザ
    init(fileURL: URL? = nil) {
        if let fileURL {
            self.fileURL = fileURL
        } else {
            let supportDir = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask)[0]
            try? FileManager.default.createDirectory(at: supportDir, withIntermediateDirectories: true)
            self.fileURL = supportDir.appendingPathComponent("diary_store.json")
        }
        load()
    }

    /// 指定日付のエントリを取得する
    func entry(for date: Diary_YMD) -> LocalDiaryEntry? {
        entries[LocalDiaryEntry.dateKey(date)]
    }

    /// 指定した年月のエントリ一覧を取得する
    func entries(year: Int, month: Int) -> [LocalDiaryEntry] {
        entries.values.filter { $0.year == year && $0.month == month }
    }

    /// 同期待ちのエントリ一覧を取得する
    func pendingEntries() -> [LocalDiaryEntry] {
        entries.values.filter(\.needsSync).sorted { $0.dateKey < $1.dateKey }
    }

    /// ローカル編集を保存する（needsSync=trueで同期待ちにする）
    func saveLocalEdit(date: Diary_YMD, content: String) {
        let key = LocalDiaryEntry.dateKey(date)
        var entry = entries[key] ?? LocalDiaryEntry(
            serverID: "",
            year: Int(date.year),
            month: Int(date.month),
            day: Int(date.day),
            content: "",
            updatedAt: 0,
            needsSync: false
        )
        entry.content = content
        entry.updatedAt = Int64(Date().timeIntervalSince1970)
        entry.needsSync = true
        entries[key] = entry
        persist()
    }

    /// サーバーから取得したエントリを反映する。
    /// ローカルに未同期の編集がある場合はローカルを優先して上書きしない。
    func applyServerEntry(_ entry: Diary_DiaryEntry) {
        guard entry.hasDate else { return }
        let key = LocalDiaryEntry.dateKey(entry.date)
        if let existing = entries[key], existing.needsSync {
            // 未同期のローカル編集を保持しつつserverIDだけ補完する（同期時の更新に必要）
            if existing.serverID.isEmpty, !entry.id.isEmpty {
                var updated = existing
                updated.serverID = entry.id
                entries[key] = updated
                persist()
            }
            return
        }
        entries[key] = LocalDiaryEntry(
            serverID: entry.id,
            year: Int(entry.date.year),
            month: Int(entry.date.month),
            day: Int(entry.date.day),
            content: entry.content,
            updatedAt: entry.updatedAt,
            needsSync: false
        )
        persist()
    }

    /// 同期完了を記録する。
    /// 同期中にユーザーがさらに編集した場合（内容が送信時と異なる場合）はneedsSyncを維持する。
    func completeSync(dateKey: String, pushedContent: String, serverEntry: Diary_DiaryEntry) {
        guard let existing = entries[dateKey] else { return }
        if existing.content != pushedContent {
            // 同期中に編集された：serverIDだけ反映して同期待ちを継続する
            var updated = existing
            updated.serverID = serverEntry.id
            entries[dateKey] = updated
            persist()
            return
        }
        entries[dateKey] = LocalDiaryEntry(
            serverID: serverEntry.id,
            year: existing.year,
            month: existing.month,
            day: existing.day,
            content: serverEntry.content,
            updatedAt: serverEntry.updatedAt,
            needsSync: false
        )
        persist()
    }

    /// 全エントリを削除する（ログアウト時に使用）
    func clear() {
        entries = [:]
        persist()
    }

    // MARK: - Private

    /// ファイルからエントリを読み込む
    private func load() {
        guard let data = try? Data(contentsOf: fileURL) else { return }
        entries = (try? JSONDecoder().decode([String: LocalDiaryEntry].self, from: data)) ?? [:]
    }

    /// エントリをファイルへ書き込む
    private func persist() {
        guard let data = try? JSONEncoder().encode(entries) else { return }
        try? data.write(to: fileURL, options: .atomic)
    }
}
