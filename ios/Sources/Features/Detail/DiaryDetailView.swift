import SwiftUI

/// 日付ごとの日記詳細・編集画面
struct DiaryDetailView: View {
    @State private var viewModel: DiaryDetailViewModel
    /// 「この日記の概要」カーテンの開閉状態（デフォルトは閉じる）
    @State private var isTimelineExpanded = false
    /// ハイライト表示から編集モードへ切り替えたかどうか
    @State private var isEditing = false

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
        .toolbar { saveToolbarButton }
        .task {
            await viewModel.fetch()
        }
        // 保存完了時に成功の触覚フィードバックを鳴らす
        .sensoryFeedback(.success, trigger: viewModel.isSaved) { _, newValue in newValue }
        .overlay(alignment: .bottom) {
            if let error = viewModel.errorMessage {
                errorBanner(message: error)
            }
        }
    }

    /// 検索キーワードのハイライト表示（読み取り専用）を出すかどうか
    private var showsHighlight: Bool {
        !highlightKeywords.isEmpty && !isEditing && !viewModel.content.isEmpty
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

            Text(TextHighlighter.highlight(text: viewModel.content, keywords: highlightKeywords))
                .font(.body)
                .foregroundStyle(Color.twBody)
                .textSelection(.enabled)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal, 4)
        }
        .padding(12)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .glassEffect(.regular, in: .rect(cornerRadius: 16))
    }

    @ToolbarContentBuilder private var saveToolbarButton: some ToolbarContent {
        ToolbarItem(placement: .topBarTrailing) {
            Button {
                Task { await viewModel.save() }
            } label: {
                if viewModel.isSaving {
                    ProgressView()
                        .controlSize(.small)
                } else if viewModel.isSaved {
                    Label("保存済み", systemImage: "checkmark")
                } else {
                    Label("保存", systemImage: "square.and.arrow.down")
                }
            }
            .buttonStyle(.glassProminent)
            .disabled(viewModel.isSaving || viewModel.isLoading)
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

    /// Unix秒を日時文字列に変換する
    private func formatTimestamp(_ timestamp: Int64) -> String {
        let date = Date(timeIntervalSince1970: TimeInterval(timestamp))
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ja_JP")
        formatter.dateFormat = "yyyy/MM/dd HH:mm"
        return formatter.string(from: date)
    }

    private func errorBanner(message: String) -> some View {
        HStack(spacing: 8) {
            Image(systemName: "exclamationmark.circle.fill")
            Text(message)
                .font(.subheadline)
            Spacer()
            Button {
                viewModel.errorMessage = nil
            } label: {
                Image(systemName: "xmark")
                    .font(.caption)
            }
        }
        .padding(16)
        .background(.red.opacity(0.15))
        .foregroundStyle(Color.twRed)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .glassEffect(.regular.tint(.red), in: .rect(cornerRadius: 12))
        .padding(16)
    }
}
