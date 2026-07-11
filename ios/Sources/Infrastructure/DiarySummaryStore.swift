import CryptoKit
import Foundation
#if canImport(FoundationModels)
    import FoundationModels
#endif

/// オンデバイスLLM（Foundation Models）による日記要約のキャッシュエントリ
struct DiarySummaryCacheEntry: Codable, Equatable {
    /// 要約元本文のハッシュ値（本文が変わったら再生成させるための判定に使う）。
    /// ディスクへ永続化し次回起動時にも比較するため、プロセスごとに値が変わる Hasher ではなく
    /// SHA256 のような安定したハッシュを使う必要がある。
    var contentHash: String
    /// 生成された要約の箇条書き（最大3件、各10文字以内を想定）
    var points: [String]
}

/// requestSummaries に渡す1件分の要約対象
struct DiarySummaryRequest {
    /// dateKey（"YYYY-MM-DD"）
    let key: String
    /// 要約元の日記本文
    let content: String
}

/// 月ごとの日記画面向けに、オンデバイスLLMで日記本文を箇条書き（重要なこと最大3点、各10文字以内）へ要約するストア。
///
/// サーバーには一切通信せず、Apple の Foundation Models framework でオンデバイス生成する。
/// 生成結果は日付キー＋本文ハッシュでキャッシュし、本文が変わらない限り再生成しない。
@MainActor
@Observable
final class DiarySummaryStore {
    /// init が nonisolated なので @MainActor 外からも生成可能。
    /// LocalDiaryStore と同じ理由で nonisolated static let にする。
    nonisolated static let shared = DiarySummaryStore()
    /// 箇条書き1件あたりの最大文字数（1行の表示枠に収まる目安）
    static let maxPointLength = 16

    /// dateKey（"YYYY-MM-DD"）をキーとした要約箇条書きのマップ。生成中のキーはこのマップに含まれない。
    private(set) var summaries: [String: [String]] = [:]
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

    /// モデルの生出力を箇条書き配列にパースする。
    /// 改行区切りを箇条書き1件とみなし、先頭の「・」「-」「*」等の記号・空白を取り除く。
    /// 最大3件、各 maxPointLength 文字を超える場合は切り詰める（表示が枠に収まらなくなるのを防ぐ）。
    static func parsePoints(_ raw: String) -> [String] {
        let bulletCharacters = CharacterSet(charactersIn: "・-*•‣◦ 　")
        return raw
            .split { $0.isNewline }
            .map { $0.trimmingCharacters(in: bulletCharacters) }
            .filter { !$0.isEmpty }
            .prefix(3)
            .map { $0.count > maxPointLength ? String($0.prefix(maxPointLength)) : $0 }
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
        if let cached = cache[key], cached.contentHash == hash {
            summaries[key] = cached.points
            return
        }

        // 既に同じ内容で生成中の場合は二重実行しない
        guard !pendingKeys.contains(key) else { return }
        pendingKeys.insert(key)

        enqueue { [weak self] in
            guard let self else { return }
            let points = await generate(content: content)
            pendingKeys.remove(key)
            guard let points, !points.isEmpty else { return }
            cache[key] = DiarySummaryCacheEntry(contentHash: hash, points: points)
            summaries[key] = points
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

    /// Foundation Models で本文から重要な点を最大3つ、各 maxPointLength 文字程度の箇条書きにする。失敗時は nil を返す。
    private func generate(content: String) async -> [String]? {
        #if canImport(FoundationModels)
            guard #available(iOS 26.0, *) else { return nil }
            do {
                let session = LanguageModelSession(
                    instructions: """
                    あなたは日記アプリの要約アシスタントです。与えられた日記本文から重要なことを最大3つ選び、\
                    それぞれ日本語12〜16文字程度の体言止めの短いフレーズにしてください。短すぎず、字数の目安まで内容を詰めてください。\
                    出力は1行につき1項目、改行区切りで最大3行のみとし、番号・記号・前置き・解説は一切付けないでください。
                    """
                )
                let response = try await session.respond(to: content)
                let points = Self.parsePoints(response.content)
                return points.isEmpty ? nil : points
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
