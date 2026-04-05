package diary

import (
	"os"
	"testing"

	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDiaryEntry_TriggerLatestTrend_ProductionEnv(t *testing.T) {
	if err := os.Setenv("ENV", "production"); err != nil {
		t.Fatalf("Setenv失敗: %v", err)
	}
	defer func() { _ = os.Unsetenv("ENV") }()

	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	_, err := svc.TriggerLatestTrend(ctx, &g.TriggerLatestTrendRequest{})
	if err == nil {
		t.Fatal("production環境でエラーが返らなかった")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("gRPCステータスエラーが返らなかった: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("コード: got %v, want %v", st.Code(), codes.PermissionDenied)
	}
}

func TestDiaryEntry_TriggerLatestTrend_NoLLMKey(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	_, err := svc.TriggerLatestTrend(ctx, &g.TriggerLatestTrendRequest{})
	if err == nil {
		t.Fatal("LLMキーが存在しないのにエラーが返らなかった")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("gRPCステータスエラーが返らなかった: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("コード: got %v, want %v", st.Code(), codes.NotFound)
	}
}
