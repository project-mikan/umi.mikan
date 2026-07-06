import Foundation
import Testing
@testable import umi_mikan

/// LocalDiaryStoreのテスト
@MainActor
struct LocalDiaryStoreTests {
    /// テスト用に一時ファイルへ保存するストアを生成する
    private func makeStore(fileURL: URL? = nil) -> LocalDiaryStore {
        let url = fileURL ?? FileManager.default.temporaryDirectory
            .appendingPathComponent("diary_store_test_\(UUID().uuidString).json")
        return LocalDiaryStore(fileURL: url)
    }

    /// テスト用の Diary_YMD を生成する
    private func ymd(_ year: Int, _ month: Int, _ day: Int) -> Diary_YMD {
        var date = Diary_YMD()
        date.year = UInt32(year)
        date.month = UInt32(month)
        date.day = UInt32(day)
        return date
    }

    /// テスト用のサーバーエントリを生成する
    private func serverEntry(id: String, date: Diary_YMD, content: String, updatedAt: Int64 = 100) -> Diary_DiaryEntry {
        var entry = Diary_DiaryEntry()
        entry.id = id
        entry.date = date
        entry.content = content
        entry.updatedAt = updatedAt
        return entry
    }

    @Test("正常系: ローカル編集を保存して取得できる")
    func saveAndLoadLocalEdit() {
        let store = makeStore()
        let date = ymd(2026, 7, 5)

        store.saveLocalEdit(date: date, content: "今日の日記")

        let entry = store.entry(for: date)
        #expect(entry?.content == "今日の日記")
        #expect(entry?.year == 2026)
        #expect(entry?.month == 7)
        #expect(entry?.day == 5)
    }

    @Test("正常系: ローカル編集はneedsSyncがtrueになり同期待ち一覧に含まれる")
    func localEditNeedsSync() {
        let store = makeStore()
        store.saveLocalEdit(date: ymd(2026, 7, 5), content: "オフライン編集")

        #expect(store.entry(for: ymd(2026, 7, 5))?.needsSync == true)
        #expect(store.pendingEntries().count == 1)
    }

    @Test("正常系: サーバーエントリの反映で同期済みエントリが上書きされる")
    func applyServerEntryOverwritesSynced() {
        let store = makeStore()
        let date = ymd(2026, 7, 5)
        store.applyServerEntry(serverEntry(id: "old-id", date: date, content: "古い内容"))

        store.applyServerEntry(serverEntry(id: "server-id", date: date, content: "サーバーの内容"))

        let entry = store.entry(for: date)
        #expect(entry?.content == "サーバーの内容")
        #expect(entry?.serverID == "server-id")
        #expect(entry?.needsSync == false)
    }

    @Test("正常系: 未同期のローカル編集はサーバーエントリで上書きされないがserverIDは補完される")
    func applyServerEntryKeepsLocalEdit() {
        let store = makeStore()
        let date = ymd(2026, 7, 5)
        store.saveLocalEdit(date: date, content: "ローカルの編集")

        store.applyServerEntry(serverEntry(id: "server-id", date: date, content: "サーバーの内容"))

        let entry = store.entry(for: date)
        #expect(entry?.content == "ローカルの編集")
        #expect(entry?.needsSync == true)
        // 同期時の更新に必要なserverIDだけ補完される
        #expect(entry?.serverID == "server-id")
    }

    @Test("正常系: completeSyncで同期完了となり同期待ちから外れる")
    func completeSyncClearsPending() {
        let store = makeStore()
        let date = ymd(2026, 7, 5)
        store.saveLocalEdit(date: date, content: "同期する内容")
        let key = LocalDiaryEntry.dateKey(date)

        store.completeSync(
            dateKey: key,
            pushedContent: "同期する内容",
            serverEntry: serverEntry(id: "server-id", date: date, content: "同期する内容")
        )

        let entry = store.entry(for: date)
        #expect(entry?.needsSync == false)
        #expect(entry?.serverID == "server-id")
        #expect(store.pendingEntries().isEmpty)
    }

    @Test("正常系: 同期中に編集された場合はneedsSyncが維持される")
    func completeSyncKeepsPendingWhenEditedDuringSync() {
        let store = makeStore()
        let date = ymd(2026, 7, 5)
        store.saveLocalEdit(date: date, content: "送信した内容")
        // 同期中にユーザーがさらに編集した状況を再現する
        store.saveLocalEdit(date: date, content: "同期中の新しい編集")
        let key = LocalDiaryEntry.dateKey(date)

        store.completeSync(
            dateKey: key,
            pushedContent: "送信した内容",
            serverEntry: serverEntry(id: "server-id", date: date, content: "送信した内容")
        )

        let entry = store.entry(for: date)
        #expect(entry?.content == "同期中の新しい編集")
        #expect(entry?.needsSync == true)
        #expect(entry?.serverID == "server-id")
    }

    @Test("正常系: 月単位でエントリを取得できる")
    func entriesByMonth() {
        let store = makeStore()
        store.saveLocalEdit(date: ymd(2026, 7, 1), content: "7月1日")
        store.saveLocalEdit(date: ymd(2026, 7, 15), content: "7月15日")
        store.saveLocalEdit(date: ymd(2026, 6, 30), content: "6月30日")

        let julyEntries = store.entries(year: 2026, month: 7)
        #expect(julyEntries.count == 2)
        #expect(store.entries(year: 2026, month: 6).count == 1)
    }

    @Test("正常系: clearで全エントリが削除される")
    func clearRemovesAll() {
        let store = makeStore()
        store.saveLocalEdit(date: ymd(2026, 7, 5), content: "削除される内容")

        store.clear()

        #expect(store.entry(for: ymd(2026, 7, 5)) == nil)
        #expect(store.pendingEntries().isEmpty)
    }

    @Test("正常系: 永続化したデータを別インスタンスで読み込める")
    func persistenceAcrossInstances() {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("diary_store_test_\(UUID().uuidString).json")
        let store = makeStore(fileURL: url)
        store.saveLocalEdit(date: ymd(2026, 7, 5), content: "永続化テスト")

        let reloaded = makeStore(fileURL: url)

        #expect(reloaded.entry(for: ymd(2026, 7, 5))?.content == "永続化テスト")
    }

    @Test("正常系: dateKeyがゼロ埋めされた形式で生成される")
    func dateKeyFormat() {
        #expect(LocalDiaryEntry.dateKey(year: 2026, month: 7, day: 5) == "2026-07-05")
        #expect(LocalDiaryEntry.dateKey(ymd(2026, 12, 31)) == "2026-12-31")
    }

    @Test("正常系: toProtoでDiary_DiaryEntryに変換できる")
    func toProtoConversion() {
        let entry = LocalDiaryEntry(
            serverID: "abc",
            year: 2026,
            month: 7,
            day: 5,
            content: "変換テスト",
            updatedAt: 1234,
            needsSync: false
        )

        let proto = entry.toProto()
        #expect(proto.id == "abc")
        #expect(proto.date.year == 2026)
        #expect(proto.date.month == 7)
        #expect(proto.date.day == 5)
        #expect(proto.content == "変換テスト")
        #expect(proto.updatedAt == 1234)
    }
}
