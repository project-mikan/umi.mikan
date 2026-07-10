package mcpserver

import (
	"testing"

	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
)

// testUUID はバリデーションのみを検証するテスト用にランダムなユーザーIDを返す
func testUUID(t *testing.T) uuid.UUID {
	t.Helper()
	return uuid.New()
}

// createDiaryReq はテスト用の日記作成リクエストを組み立てる
func createDiaryReq(year, month, day uint32, content string) *g.CreateDiaryEntryRequest {
	return &g.CreateDiaryEntryRequest{
		Content: content,
		Date:    &g.YMD{Year: year, Month: month, Day: day},
	}
}
