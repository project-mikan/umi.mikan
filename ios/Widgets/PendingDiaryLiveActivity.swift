import ActivityKit
import SwiftUI
import WidgetKit

/// 未同期・書きかけの日記を知らせるLive Activity。
/// ロック画面とDynamic Islandに表示し、未同期がある場合は「今すぐ同期」ボタンで同期を実行できる。
/// 書きかけがある場合はタップしてアプリに戻り続きを書けることを知らせる。
struct PendingDiaryLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: DiaryActivityAttributes.self) { context in
            lockScreenView(state: context.state)
        } dynamicIsland: { context in
            dynamicIsland(state: context.state)
        }
    }

    /// Dynamic Island のレイアウト
    private func dynamicIsland(state: DiaryActivityAttributes.ContentState) -> DynamicIsland {
        DynamicIsland {
            DynamicIslandExpandedRegion(.leading) {
                Image(systemName: iconName(state: state))
                    .foregroundStyle(.orange)
                    .padding(.leading, 4)
            }
            DynamicIslandExpandedRegion(.trailing) {
                Text(trailingText(state: state))
                    .fontWeight(.semibold)
                    .foregroundStyle(.orange)
                    .padding(.trailing, 4)
            }
            DynamicIslandExpandedRegion(.bottom) {
                HStack {
                    statusText(state: state)
                    Spacer()
                    if state.pendingCount > 0 {
                        syncButton(state: state)
                    }
                }
            }
        } compactLeading: {
            Image(systemName: iconName(state: state))
                .foregroundStyle(.orange)
        } compactTrailing: {
            if state.pendingCount > 0 {
                Text("\(state.pendingCount)")
                    .foregroundStyle(.orange)
            } else {
                Image(systemName: "ellipsis")
                    .foregroundStyle(.orange)
            }
        } minimal: {
            Image(systemName: iconName(state: state))
                .foregroundStyle(.orange)
        }
    }

    /// ロック画面・通知エリア用のビュー
    private func lockScreenView(state: DiaryActivityAttributes.ContentState) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(spacing: 6) {
                Image(systemName: iconName(state: state))
                    .foregroundStyle(.orange)
                Text(titleText(state: state))
                    .font(.subheadline)
                    .fontWeight(.semibold)
                Spacer()
            }
            HStack {
                statusText(state: state)
                Spacer()
                if state.pendingCount > 0 {
                    syncButton(state: state)
                }
            }
        }
        .padding(14)
    }

    /// 状態に応じたアイコン名（書きかけを優先して表示する）
    private func iconName(state: DiaryActivityAttributes.ContentState) -> String {
        state.hasDraft ? "pencil.line" : "book.closed.fill"
    }

    /// ロック画面のタイトル文言
    private func titleText(state: DiaryActivityAttributes.ContentState) -> String {
        if state.hasDraft {
            return state.pendingCount > 0
                ? "書きかけの日記があります（未同期\(state.pendingCount)件）"
                : "書きかけの日記があります"
        }
        return "未同期の日記が\(state.pendingCount)件あります"
    }

    /// Dynamic Island展開時の右側の文言
    private func trailingText(state: DiaryActivityAttributes.ContentState) -> String {
        state.pendingCount > 0 ? "\(state.pendingCount)件" : "書きかけ"
    }

    /// 状態の説明テキスト
    private func statusText(state: DiaryActivityAttributes.ContentState) -> some View {
        Text(statusMessage(state: state))
            .font(.caption)
            .foregroundStyle(.secondary)
    }

    /// 状態の説明文言（同期中 > 書きかけ > 未同期 の優先順）
    private func statusMessage(state: DiaryActivityAttributes.ContentState) -> String {
        if state.isSyncing {
            return "同期中..."
        }
        if state.hasDraft {
            return "タップしてアプリに戻り続きを書けます"
        }
        return "オンラインになると自動で同期されます"
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
