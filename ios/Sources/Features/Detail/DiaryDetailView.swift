import SwiftUI

/// 日付ごとの日記詳細・編集画面
struct DiaryDetailView: View {
    /// Unix秒を日時文字列に変換する（DateFormatter は再生成コストが高いため static にキャッシュする）
    private static let timestampFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ja_JP")
        formatter.timeZone = TimeZone(identifier: "Asia/Tokyo")
        formatter.dateFormat = "yyyy/MM/dd HH:mm"
        return formatter
    }()

    @State private var viewModel: DiaryDetailViewModel
    /// 「この日記の概要」カーテンの開閉状態（デフォルトは閉じる）
    @State private var isTimelineExpanded = false
    /// ハイライト表示から編集モードへ切り替えたかどうか
    @State private var isEditing = false
    /// 本文エディタのフォーカス状態（キーボードツールバーの閉じるボタン用）
    @FocusState private var isEditorFocused: Bool

    /// 検索結果から開いた場合にハイライトするキーワード
    private let highlightKeywords: [String]

    // swiftlint:disable:next type_contents_order
    init(date: Diary_YMD, authViewModel: AuthViewModel, syncManager: SyncManager, highlightKeywords: [String] = []) {
        self.highlightKeywords = highlightKeywords
        _viewModel = State(
            initialValue: DiaryDetailViewModel(date: date, authViewModel: authViewModel, syncManager: syncManager)
        )
    }

    var body: some View {
        ScrollViewReader { proxy in
            scrollContent
                .toolbar { keyboardToolbar }
                .task {
                    await viewModel.fetch()
                    await scrollToFirstHighlight(proxy)
                }
                // カーソル（フォーカス）が外れたら未保存の変更を自動保存する
                .onChange(of: isEditorFocused) { _, focused in
                    if !focused, viewModel.hasUnsavedChanges {
                        Task { await viewModel.save() }
                    }
                }
        }
    }

    /// 本文のスクロールビュー（ツールバー・fetch以外の共通修飾込み）
    private var scrollContent: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                if viewModel.isLoading {
                    loadingView
                } else {
                    if let status = viewModel.embeddingStatus {
                        chunkTimelineCurtain(status)
                    }
                    if showsHighlight {
                        highlightCard
                    } else {
                        editorCard
                    }
                }
            }
            .padding(16)
        }
        .navigationTitle(dateTitle)
        .navigationBarTitleDisplayMode(.inline)
        // 保存完了時に成功の触覚フィードバックを鳴らす
        .sensoryFeedback(.success, trigger: viewModel.isSaved) { _, newValue in newValue }
        .overlay(alignment: .bottom) {
            if let error = viewModel.errorMessage {
                ErrorBannerView(message: error) { viewModel.errorMessage = nil }
            }
        }
    }

    /// 検索キーワードのハイライト表示（読み取り専用）を出すかどうか
    private var showsHighlight: Bool {
        !highlightKeywords.isEmpty && !isEditing && !viewModel.content.isEmpty
    }

    /// ハイライト表示用に本文を行単位に分割したもの（自動スクロールの行IDに使う）
    private var highlightLines: [String] {
        viewModel.content.components(separatedBy: "\n")
    }

    private var dateTitle: String {
        String(format: "%d年%d月%d日", viewModel.date.year, viewModel.date.month, viewModel.date.day)
    }

    private var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .controlSize(.large)
            Text("読み込み中...")
                .font(.subheadline)
                .foregroundStyle(Color.twSecondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 60)
    }

    private var editorCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            TextEditor(text: $viewModel.content)
                .focused($isEditorFocused)
                .frame(minHeight: 300)
                .scrollContentBackground(.hidden)
                .background(.clear)
                .font(.body)
                .foregroundStyle(Color.twBody)
                .padding(.horizontal, 4)

            HStack {
                Text("\(viewModel.content.count)文字")
                    .font(.caption2)
                    .foregroundStyle(Color.twSecondary)
                if viewModel.content.count >= 1000 {
                    Text("自動要約対象")
                        .font(.caption2)
                        .foregroundStyle(Color.twBlue)
                }
                Spacer()
            }
        }
        .padding(12)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .glassEffect(.regular, in: .rect(cornerRadius: 16))
    }

    /// 検索キーワードをハイライトした読み取り専用の本文カード。
    /// 「編集」を押すと通常のエディタへ切り替わる。
    private var highlightCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Label("検索キーワードをハイライト表示中", systemImage: "highlighter")
                    .font(.caption)
                    .foregroundStyle(Color.twSecondary)
                Spacer()
                Button {
                    isEditing = true
                } label: {
                    Label("編集", systemImage: "pencil")
                        .font(.caption)
                        .fontWeight(.medium)
                        .frame(height: 28)
                        .padding(.horizontal, 8)
                }
                .buttonStyle(.glass)
            }

            // 自動スクロールできるよう行単位でレンダリングし、各行にIDを振る
            VStack(alignment: .leading, spacing: 2) {
                ForEach(Array(highlightLines.enumerated()), id: \.offset) { lineIndex, line in
                    Text(TextHighlighter.highlight(text: line.isEmpty ? " " : line, keywords: highlightKeywords))
                        .font(.body)
                        .foregroundStyle(Color.twBody)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .id("highlight-line-\(lineIndex)")
                }
            }
            .textSelection(.enabled)
            .padding(.horizontal, 4)
        }
        .padding(12)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .glassEffect(.regular, in: .rect(cornerRadius: 16))
    }

    /// キーボードの上に表示するツールバー（保存・キーボードを閉じる）
    @ToolbarContentBuilder private var keyboardToolbar: some ToolbarContent {
        ToolbarItemGroup(placement: .keyboard) {
            Spacer()
            // 閉じるボタンと見分けやすいよう、チェックマーク＋青の塗りつぶしボタンにする
            Button {
                Task { await viewModel.save() }
            } label: {
                Label(
                    viewModel.isSaved ? "保存済み" : "保存",
                    systemImage: viewModel.isSaved ? "checkmark" : "checkmark.circle.fill"
                )
                .labelStyle(.titleAndIcon)
                .fontWeight(.semibold)
            }
            .buttonStyle(.borderedProminent)
            .tint(Color.twBlue)
            .controlSize(.small)
            .disabled(viewModel.isSaving)
            Button {
                isEditorFocused = false
            } label: {
                Image(systemName: "keyboard.chevron.compact.down")
            }
        }
    }

    /// カーテンのヘッダー（タップで開閉）
    private var curtainHeader: some View {
        Button {
            withAnimation(.easeInOut(duration: 0.25)) {
                isTimelineExpanded.toggle()
            }
        } label: {
            HStack {
                Text("この日記の概要")
                    .font(.subheadline)
                    .fontWeight(.semibold)
                    .foregroundStyle(Color.twIndigo)
                Spacer()
                Image(systemName: "chevron.down")
                    .font(.caption)
                    .foregroundStyle(Color.twIndigo)
                    .rotationEffect(.degrees(isTimelineExpanded ? 180 : 0))
            }
            .padding(14)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    /// この日記の概要（チャンク一覧）カーテン。
    /// ヘッダーをタップすると上から下へカーテンのように開閉する。
    /// 最初にキーワードがマッチした行まで自動スクロールする。
    /// レイアウト確定を待つため少し遅らせてから実行する。
    private func scrollToFirstHighlight(_ proxy: ScrollViewProxy) async {
        guard showsHighlight else { return }
        guard let line = TextHighlighter.firstMatchLineIndex(lines: highlightLines, keywords: highlightKeywords) else {
            return
        }
        try? await Task.sleep(for: .milliseconds(300))
        withAnimation(.easeInOut(duration: 0.4)) {
            proxy.scrollTo("highlight-line-\(line)", anchor: .center)
        }
    }

    private func chunkTimelineCurtain(_ status: Diary_GetDiaryEmbeddingStatusResponse) -> some View {
        VStack(alignment: .leading, spacing: 0) {
            curtainHeader
            if isTimelineExpanded {
                curtainBody(status)
                    .transition(.move(edge: .top).combined(with: .opacity))
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular.tint(.indigo.opacity(0.15)), in: .rect(cornerRadius: 14))
    }

    /// カーテン本体（チャンクごとの概要をタイムライン形式で表示する）
    private func curtainBody(_ status: Diary_GetDiaryEmbeddingStatusResponse) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            VStack(alignment: .leading, spacing: 10) {
                ForEach(Array(status.chunkSummaries.enumerated()), id: \.offset) { _, summary in
                    HStack(alignment: .top, spacing: 10) {
                        Circle()
                            .strokeBorder(Color.twIndigo, lineWidth: 2)
                            .frame(width: 10, height: 10)
                            .padding(.top, 4)
                        Text(summary.isEmpty ? "（概要なし）" : summary)
                            .font(.subheadline)
                            .foregroundStyle(Color.twBody)
                    }
                }
            }

            Text("最終更新日時: \(formatTimestamp(status.updatedAt))")
                .font(.caption2)
                .foregroundStyle(Color.twSecondary)
        }
        .padding(.horizontal, 14)
        .padding(.bottom, 14)
    }

    private func formatTimestamp(_ timestamp: Int64) -> String {
        let date = Date(timeIntervalSince1970: TimeInterval(timestamp))
        return Self.timestampFormatter.string(from: date)
    }
}
