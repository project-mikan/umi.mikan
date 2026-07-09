import SwiftUI

/// ホーム画面の「おもいで」セクション（今日カードと昨日カードの間に配置する）。
/// 該当する日記が1件もない場合はセクションごと非表示にする。
struct MemorySectionView: View {
    let items: [MemoryItem]
    let onSelect: (MemoryItem) -> Void

    var body: some View {
        if !items.isEmpty {
            VStack(alignment: .leading, spacing: 10) {
                Text("おもいで")
                    .font(.subheadline)
                    .fontWeight(.semibold)
                    .foregroundStyle(Color.twHeading)

                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 12) {
                        ForEach(items) { item in
                            MemoryCardView(item: item) {
                                onSelect(item)
                            }
                        }
                    }
                    .padding(.horizontal, 2)
                }
            }
        }
    }
}

/// 「おもいで」1件分のカード
private struct MemoryCardView: View {
    let item: MemoryItem
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            VStack(alignment: .leading, spacing: 8) {
                Text("\(item.yearsAgo)年前")
                    .font(.caption)
                    .fontWeight(.semibold)
                    .foregroundStyle(Color.twBlue)
                Text(item.entry.content)
                    .font(.caption)
                    .foregroundStyle(Color.twBody)
                    .lineLimit(4)
                    .multilineTextAlignment(.leading)
                Spacer(minLength: 0)
            }
            .padding(12)
            .frame(width: 180, height: 120, alignment: .topLeading)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
    }
}
