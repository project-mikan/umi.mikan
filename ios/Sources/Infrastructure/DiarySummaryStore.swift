import Foundation
#if canImport(FoundationModels)
    import FoundationModels
#endif

/// オンデバイスLLM（Foundation Models）による日記要約のキャッシュエントリ
struct DiarySummaryCacheEntry: Codable, Equatable {
    /// 要約元本文のハッシュ値（本文が変わったら再生成させるための判定に使う）
    var contentHash: Int
    /// 生成された要約テキスト
    var summary: String
}

/// 月ごとの日記画面向けに、オンデバイスLLMで日記本文を1〜2文へ要約するストア。
///
/// サーバーには一切通信せず、Apple の Foundation Models framework でオンデバイス生成する。
/// 生成結果は日付キー＋本文ハッシュでキャッシュし、本文が変わらない限り再生成しない。
@MainActor
@Observable
final class DiarySummaryStore {
    /// init が nonisolated なので @MainActor 外からも生成可能。
    /// LocalDiaryStore と同じ理由で nonisolated static let にする。
    nonisolated static let shared = DiarySummaryStore()

    /// dateKey（"YYYY-MM-DD"）をキーとした要約のマップ。生成中のキーはこのマップに含まれない。
    private(set) var summaries: [String: String] = [:]
    /// 生成中の dateKey の集合（View 側で生成中かどうかを判定するために使う）
    private(set) var pendingKeys: Set<String> = []

    @ObservationIgnored private var cache: [String: DiarySummaryCacheEntry]
    private let fileURL: URL
    private let maxConcurrent = 3
    /// 同時生成数を制限するためのセマフォ的カウンタ
    @ObservationIgnored private var runningCount = 0
    /// 同時実行数を超えた際に待機させるキュー
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

    /// 本文の内容ハッシュを計算する（キャッシュの有効性判定に使う）
    static func contentHash(_ content: String) -> Int {
        var hasher = Hasher()
        hasher.combine(content)
        return hasher.finalize()
    }

    /// 指定日の要約をリクエストする。
    /// キャッシュ済み（本文ハッシュ一致）なら summaries に即座に反映する。
    /// 未キャッシュ・本文変更時は非同期生成をキューイングする（同時実行数は maxConcurrent で制限）。
    func requestSummary(key: String, content: String) {
        guard isAvailable, !content.isEmpty else { return }

        let hash = Self.contentHash(content)
        if let cached = cache[key], cached.contentHash == hash {
            summaries[key] = cached.summary
            return
        }

        // 既に同じ内容で生成中の場合は二重実行しない
        guard !pendingKeys.contains(key) else { return }
        pendingKeys.insert(key)

        enqueue { [weak self] in
            guard let self else { return }
            let result = await generate(content: content)
            pendingKeys.remove(key)
            guard let result else { return }
            cache[key] = DiarySummaryCacheEntry(contentHash: hash, summary: result)
            summaries[key] = result
            persist()
        }
    }

    // MARK: - Private

    /// 同時実行数を制限しつつ非同期タスクを実行する
    private func enqueue(_ task: @escaping () async -> Void) {
        let work: () -> Void = { [weak self] in
            let runner = Task { @MainActor in
                await task()
                self?.runNext()
            }
            _ = runner
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

    /// Foundation Models で本文を1〜2文に要約する。失敗時は nil を返す。
    private func generate(content: String) async -> String? {
        #if canImport(FoundationModels)
            guard #available(iOS 26.0, *) else { return nil }
            do {
                let session = LanguageModelSession(
                    instructions: "あなたは日記アプリの要約アシスタントです。与えられた日記本文を、日本語で1〜2文の簡潔な要約にしてください。要約以外の文章（前置きや解説）は出力しないでください。"
                )
                let response = try await session.respond(to: content)
                let trimmed = response.content.trimmingCharacters(in: .whitespacesAndNewlines)
                return trimmed.isEmpty ? nil : trimmed
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
