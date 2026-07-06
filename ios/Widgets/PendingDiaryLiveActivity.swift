import ActivityKit
import SwiftUI
import WidgetKit

/// 未同期の日記の件数を知らせるLive Activity。
/// ロック画面とDynamic Islandに表示し、「今すぐ同期」ボタンで同期を実行できる。
struct PendingDiaryLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: DiaryActivityAttributes.self) { context in
            lockScreenView(state: context.state)
        } dynamicIsland: { context in
            DynamicIsland {
                DynamicIslandExpandedRegion(.leading) {
                    Image(systemName: "book.closed.fill")
                        .foregroundStyle(.orange)
                        .padding(.leading, 4)
                }
                DynamicIslandExpandedRegion(.trailing) {
                    Text("\(context.state.pendingCount)件")
                        .fontWeight(.semibold)
                        .foregroundStyle(.orange)
                        .padding(.trailing, 4)
                }
                DynamicIslandExpandedRegion(.bottom) {
                    HStack {
                        statusText(state: context.state)
                        Spacer()
                        syncButton(state: context.state)
                    }
                }
            } compactLeading: {
                Image(systemName: "book.closed.fill")
                    .foregroundStyle(.orange)
            } compactTrailing: {
                Text("\(context.state.pendingCount)")
                    .foregroundStyle(.orange)
            } minimal: {
                Image(systemName: "book.closed.fill")
                    .foregroundStyle(.orange)
            }
        }
    }

    /// ロック画面・通知エリア用のビュー
    private func lockScreenView(state: DiaryActivityAttributes.ContentState) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: "book.closed.fill")
                    .foregroundStyle(.orange)
                Text("未同期の日記が\(state.pendingCount)件あります")
                    .font(.subheadline)
                    .fontWeight(.semibold)
                Spacer()
            }
            HStack {
                statusText(state: state)
                Spacer()
                syncButton(state: state)
            }
        }
        .padding(14)
    }

    /// 同期状態の説明テキスト
    private func statusText(state: DiaryActivityAttributes.ContentState) -> some View {
        Text(state.isSyncing ? "同期中..." : "オンラインになると自動で同期されます")
            .font(.caption)
            .foregroundStyle(.secondary)
    }

    /// 「今すぐ同期」ボタン（LiveActivityIntentでアプリ本体の同期処理を起動する）
    private func syncButton(state: DiaryActivityAttributes.ContentState) -> some View {
        Button(intent: SyncPendingDiariesIntent()) {
            Label(state.isSyncing ? "同期中" : "今すぐ同期", systemImage: "arrow.triangle.2.circlepath")
                .font(.caption)
                .fontWeight(.medium)
        }
        .buttonStyle(.borderedProminent)
        .tint(.orange)
        .disabled(state.isSyncing)
    }
}
