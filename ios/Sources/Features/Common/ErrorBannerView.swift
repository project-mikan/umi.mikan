import SwiftUI

/// エラーメッセージを画面下部に表示する共通バナー。
/// onDismiss をタップすると非表示になる。
struct ErrorBannerView: View {
    let message: String
    let onDismiss: () -> Void

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: "exclamationmark.circle.fill")
            Text(message)
                .font(.subheadline)
            Spacer()
            Button {
                onDismiss()
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
