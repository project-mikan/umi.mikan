package mcpserver

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
