import CryptoKit
import Foundation
#if canImport(FoundationModels)
    import FoundationModels
#endif

/// オンデバイスLLM（Foundation Models）による日記要約のキャッシュエントリ
struct DiarySummaryCacheEntry: Codable, Equatable {
    /// 要約形式の現行バージョン。プロンプトや想定文字数など「本文が同じでも生成し直したい変更」をしたら上げる。
    /// 旧バージョンのエントリはキャッシュ不一致として扱われ、再生成される。
    static let currentVersion = 2

    /// 要約元本文のハッシュ値（本文が変わったら再生成させるための判定に使う）。
    /// ディスクへ永続化し次回起動時にも比較するため、プロセスごとに値が変わる Hasher ではなく
    /// SHA256 のような安定したハッシュを使う必要がある。
    var contentHash: String
    /// 生成された要約文（主な出来事2〜3個をつないだ1文、全角40〜55文字程度を想定）
    var summary: String
    /// このエントリを生成した時点の要約形式バージョン
    var version: Int
}

/// requestSummaries に渡す1件分の要約対象
struct DiarySummaryRequest {
    /// dateKey（"YYYY-MM-DD"）
    let key: String
    /// 要約元の日記本文
    let content: String
}

/// 月ごとの日記画面向けに、オンデバイスLLMで日記本文を「どんな日だったか」がパッと分かる1文
/// （主な出来事2〜3個をつないだ、表示枠の2〜3行に収まる要約）へ要約するストア。
///
/// サーバーには一切通信せず、Apple の Foundation Models framework でオンデバイス生成する。
/// 生成結果は日付キー＋本文ハッシュでキャッシュし、本文が変わらない限り再生成しない。
@MainActor
@Observable
final class DiarySummaryStore {
    /// init が nonisolated なので @MainActor 外からも生成可能。
    /// LocalDiaryStore と同じ理由で nonisolated static let にする。
    nonisolated static let shared = DiarySummaryStore()
    /// 要約文の最大文字数（2〜3行の折り返し表示枠に収まる目安）
    static let maxSummaryLength = 60

    /// dateKey（"YYYY-MM-DD"）をキーとした要約文のマップ。生成中のキーはこのマップに含まれない。
    private(set) var summaries: [String: String] = [:]
    /// 生成中の dateKey の集合（View 側で生成中かどうかを判定するために使う）
    private(set) var pendingKeys: Set<String> = []
    /// 直前に生成が完了した dateKey の集合。虹色フラッシュ演出を一度だけ再生するために使い、
    /// 演出開始後に View 側が consumeJustCompleted で消費する。
    private(set) var justCompletedKeys: Set<String> = []

    @ObservationIgnored private var cache: [String: DiarySummaryCacheEntry]
    private let fileURL: URL
    private let maxConcurrent = 3
    /// 同時生成数を制限するためのセマフォ的カウンタ
    @ObservationIgnored private var runningCount = 0
    /// 開始待ちのリクエストキュー（呼び出し側が渡した順序を保証するため、日付昇順で積まれる想定）
    @ObservationIgnored private var waitQueue: [() -> Void] = []

    /// 現在の端末・OS設定でオンデバイス要約が利用可能かどうか
    var isAvailable: Bool {
        #if canImport(FoundationModels)
            if #available(iOS 26.0, *) {
                return SystemLanguageModel.default.availability == .available
            }
            return false
        #else
            return false
        #endif
    }

    /// テスト用に保存先ファイルを指定できるイニシャライザ
    nonisolated init(fileURL: URL? = nil) {
        let resolvedURL = _resolveDiarySummaryStoreFileURL(fileURL)
        self.fileURL = resolvedURL
        cache = _loadDiarySummaryCache(from: resolvedURL)
    }

    /// 本文の内容ハッシュを計算する（キャッシュの有効性判定に使う）。
    /// Swift の Hasher はプロセスごとにシードがランダム化されるため、アプリ再起動をまたいで
    /// ディスクキャッシュと比較する用途には使えない（比較が常に不一致になりキャッシュが無効化される）。
    /// そのため SHA256 のようなプロセスに依存しない安定したハッシュを使う。
    static func contentHash(_ content: String) -> String {
        let digest = SHA256.hash(data: Data(content.utf8))
        return digest.map { String(format: "%02x", $0) }.joined()
    }

    /// モデルの生出力を1行の要約文にパースする。
    /// 複数行返ってきた場合は先頭の非空行のみを採用し、前後の空白・箇条書き記号を取り除く。
    /// maxSummaryLength 文字を超える場合は切り詰める（2〜3行の表示枠に収まらなくなるのを防ぐ）。
    static func parseSummary(_ raw: String) -> String {
        let bulletCharacters = CharacterSet(charactersIn: "・-*•‣◦ 　")
        let lines = raw.split { $0.isNewline }
        let trimmedLines = lines.map { $0.trimmingCharacters(in: bulletCharacters) }
        guard let firstLine = trimmedLines.first(where: { !$0.isEmpty }) else {
            return ""
        }
        return firstLine.count > maxSummaryLength ? String(firstLine.prefix(maxSummaryLength)) : firstLine
    }

    /// 複数日分の要約をまとめてリクエストする。
    /// 呼び出し側が渡した配列の順序どおりにキューへ積むため、日付昇順の配列を渡せば
    /// 「月の上（1日）から順に」生成が開始される（同時実行数 maxConcurrent の範囲で並列実行はされる）。
    /// キャッシュ済み・生成中のキーは自動的にスキップされる。
    func requestSummaries(_ requests: [DiarySummaryRequest]) {
        for request in requests {
            requestSummary(key: request.key, content: request.content)
        }
    }

    /// 指定日の要約をリクエストする。
    /// キャッシュ済み（本文ハッシュ一致）なら summaries に即座に反映する。
    /// 未キャッシュ・本文変更時は非同期生成をキューイングする（同時実行数は maxConcurrent で制限）。
    func requestSummary(key: String, content: String) {
        guard isAvailable, !content.isEmpty else { return }

        let hash = Self.contentHash(content)
        // 本文ハッシュに加えて要約形式バージョンも一致した場合のみキャッシュを使う
        // （プロンプト変更などで形式が変わった旧要約を表示し続けないようにするため）
        if let cached = cache[key], cached.contentHash == hash, cached.version == DiarySummaryCacheEntry.currentVersion {
            summaries[key] = cached.summary
            return
        }

        // 既に同じ内容で生成中の場合は二重実行しない
        guard !pendingKeys.contains(key) else { return }
        pendingKeys.insert(key)

        enqueue { [weak self] in
            guard let self else { return }
            let summary = await generate(content: content)
            pendingKeys.remove(key)
            guard let summary, !summary.isEmpty else { return }
            cache[key] = DiarySummaryCacheEntry(
                contentHash: hash,
                summary: summary,
                version: DiarySummaryCacheEntry.currentVersion
            )
            summaries[key] = summary
            justCompletedKeys.insert(key)
            persist()
        }
    }

    /// 虹色フラッシュ演出を再生し終えた View から呼ばれ、演出の再トリガーを防ぐ
    func consumeJustCompleted(key: String) {
        justCompletedKeys.remove(key)
    }

    // MARK: - Private

    /// 同時実行数を制限しつつ非同期タスクを実行する。
    /// waitQueue は FIFO で消費するため、呼び出し順（= requestSummaries に渡した順）どおりに開始される。
    /// task 内で早期 return しても runningCount が狂わないよう、defer で必ず runNext を呼ぶ。
    private func enqueue(_ task: @escaping () async -> Void) {
        let work: () -> Void = { [weak self] in
            Task { @MainActor in
                defer { self?.runNext() }
                await task()
            }
        }
        if runningCount < maxConcurrent {
            runningCount += 1
            work()
        } else {
            waitQueue.append(work)
        }
    }

    /// 実行中タスクが1件終わったら、待機中の次のタスクを開始する
    private func runNext() {
        if waitQueue.isEmpty {
            runningCount = max(0, runningCount - 1)
            return
        }
        let next = waitQueue.removeFirst()
        next()
    }

    /// Foundation Models で本文からその日の主な出来事を複数拾い、「どんな日だったか」がパッと分かる
    /// 全角40〜55文字程度の1文（表示枠の2〜3行に収まる長さ）に要約する。失敗時は nil を返す。
    private func generate(content: String) async -> String? {
        #if canImport(FoundationModels)
            guard #available(iOS 26.0, *) else { return nil }
            do {
                let session = LanguageModelSession(
                    instructions: """
                    あなたは日記アプリの要約アシスタントです。与えられた日記本文から主な出来事を2〜3個拾い、\
                    「〜して、〜で、〜だった」のように自然につないで、どんな日だったかがひと目で分かる日本語40〜55文字程度の1文にまとめてください。\
                    出来事が1つしかない場合はその内容を字数の目安まで具体的にまとめてください。\
                    箇条書きや改行、番号、記号は使わず、前置きや解説も付けず、要約の1文だけを出力してください。
                    """
                )
                let response = try await session.respond(to: content)
                let summary = Self.parseSummary(response.content)
                return summary.isEmpty ? nil : summary
            } catch {
                return nil
            }
        #else
            return nil
        #endif
    }

    /// キャッシュをファイルへ書き込む
    private func persist() {
        guard let data = try? JSONEncoder().encode(cache) else { return }
        try? data.write(to: fileURL, options: .atomic)
    }
}

// MARK: - File-scope helpers

/// DiarySummaryStore の保存先 URL を解決する
private nonisolated func _resolveDiarySummaryStoreFileURL(_ override: URL?) -> URL {
    if let override {
        return override
    }
    let supportDir = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask)[0]
    try? FileManager.default.createDirectory(at: supportDir, withIntermediateDirectories: true)
    return supportDir.appendingPathComponent("diary_summary_cache.json")
}

/// DiarySummaryStore のキャッシュをファイルから読み込む
private nonisolated func _loadDiarySummaryCache(from url: URL) -> [String: DiarySummaryCacheEntry] {
    let data = try? Data(contentsOf: url)
    return data.flatMap { try? JSONDecoder().decode([String: DiarySummaryCacheEntry].self, from: $0) } ?? [:]
}
