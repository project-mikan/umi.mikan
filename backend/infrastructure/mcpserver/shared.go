package mcpserver

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"google.golang.org/grpc/status"
)

// userIDFromContext は認証ミドルウェアがコンテキストに注入したユーザーIDを取得する
func userIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("unauthenticated: %w", err)
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user id in token: %w", err)
	}
	return userID, nil
}

// friendlyError はサービス層が返すgRPC statusエラーを人が読みやすいメッセージに変換する
func friendlyError(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		return fmt.Errorf("%s", st.Message())
	}
	return err
}

// toDiaryEntryOutputs は database.Diary のスライスをMCPツール共通の出力形式に変換する。
// get_diary_entries_by_range と search_diary_entries_fulltext の両方で同じマッピングが
// 必要になるため、ここに集約して片方だけの修正漏れを防ぐ。
func toDiaryEntryOutputs(diaries []*database.Diary) []DiaryEntryOutput {
	entries := make([]DiaryEntryOutput, 0, len(diaries))
	for _, d := range diaries {
		entries = append(entries, DiaryEntryOutput{
			ID:        d.ID.String(),
			Date:      d.Date.Format(dateLayout),
			Content:   d.Content,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		})
	}
	return entries
}
