package mcpserver

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserIDFromContext(t *testing.T) {
	t.Run("正常系: コンテキストに有効なユーザーIDが設定されている場合はUUIDを返す", func(t *testing.T) {
		userID := uuid.New()
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

		got, err := userIDFromContext(ctx)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if got != userID {
			t.Errorf("ユーザーID: 期待 %v, 実際 %v", userID, got)
		}
	})

	t.Run("異常系: コンテキストにユーザーIDが設定されていない場合はエラーを返す", func(t *testing.T) {
		_, err := userIDFromContext(context.Background())
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: コンテキストのユーザーIDがUUID形式でない場合はエラーを返す", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, "not-a-uuid")
		_, err := userIDFromContext(ctx)
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
	})
}

func TestFriendlyError(t *testing.T) {
	t.Run("正常系: nilを渡した場合はnilを返す", func(t *testing.T) {
		if err := friendlyError(nil); err != nil {
			t.Errorf("nilを期待したが %v が返った", err)
		}
	})

	t.Run("正常系: gRPC statusエラーはメッセージのみを抽出する", func(t *testing.T) {
		err := status.Errorf(codes.NotFound, "Gemini API key not found")
		got := friendlyError(err)
		if got.Error() != "Gemini API key not found" {
			t.Errorf("メッセージ: 期待 %q, 実際 %q", "Gemini API key not found", got.Error())
		}
	})

	t.Run("正常系: 通常のエラーはそのままのメッセージを返す", func(t *testing.T) {
		err := errors.New("plain error")
		got := friendlyError(err)
		if got.Error() != "plain error" {
			t.Errorf("メッセージ: 期待 %q, 実際 %q", "plain error", got.Error())
		}
	})
}
