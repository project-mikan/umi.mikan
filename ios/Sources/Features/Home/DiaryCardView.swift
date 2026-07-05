import SwiftUI

/// 日記カード - 日付ラベル・テキストエリア・保存ボタンをまとめたコンポーネント
struct DiaryCardView: View {
    let title: String
    let date: Diary_YMD
    @Binding var content: String
    let isSaving: Bool
    let isSaved: Bool
    let onSave: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            cardHeader
            cardEditor
        }
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .glassEffect(.regular, in: .rect(cornerRadius: 16))
    }

    private var cardHeader: some View {
        HStack {
            // タイトル・日付をタップすると日記詳細へ遷移する
            NavigationLink(value: date) {
                HStack(spacing: 6) {
                    VStack(alignment: .leading, spacing: 2) {
                        Text(title)
                            .font(.headline)
                            .fontWeight(.semibold)
                            .foregroundStyle(Color.twHeading)
                        Text(dateString)
                            .font(.caption)
                            .foregroundStyle(Color.twSecondary)
                    }
                    Image(systemName: "chevron.right")
                        .font(.caption)
                        .foregroundStyle(Color.twSecondary)
                }
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            Spacer()
            saveButton
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
        .background(.ultraThinMaterial)
    }

    private var cardEditor: some View {
        VStack(alignment: .leading, spacing: 8) {
            TextEditor(text: $content)
                .frame(minHeight: 160)
                .scrollContentBackground(.hidden)
                .background(.clear)
                .font(.body)
                .foregroundStyle(Color.twBody)
                .padding(.horizontal, 4)

            HStack {
                Text("\(content.count)文字")
                    .font(.caption2)
                    .foregroundStyle(Color.twSecondary)
                if content.count >= 1000 {
                    Text("自動要約対象")
                        .font(.caption2)
                        .foregroundStyle(Color.twBlue)
                }
                Spacer()
            }
        }
        .padding(12)
    }

    private var saveButton: some View {
        Group {
            if isSaved {
                Button { onSave() } label: {
                    Label("保存済み", systemImage: "checkmark")
                        .font(.caption)
                        .fontWeight(.medium)
                        .frame(height: 32)
                        .padding(.horizontal, 12)
                }
                .buttonStyle(.glass)
            } else {
                Button { onSave() } label: {
                    Group {
                        if isSaving {
                            ProgressView()
                                .controlSize(.small)
                                .tint(.white)
                        } else {
                            Label("保存", systemImage: "square.and.arrow.down")
                                .font(.caption)
                                .fontWeight(.medium)
                        }
                    }
                    .frame(height: 32)
                    .padding(.horizontal, 12)
                }
                .buttonStyle(.glassProminent)
                .disabled(isSaving)
            }
        }
    }

    /// Diary_YMD を "YYYY/MM/DD" 形式の文字列に変換する
    private var dateString: String {
        String(format: "%d/%02d/%02d", date.year, date.month, date.day)
    }
}
