package diary

import (
	"testing"
	"time"

	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
)

func TestDiaryEntry_GetDiaryEntriesByDateRange(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// テスト用日記を作成
	dates := []struct {
		year, month, day uint32
		content          string
	}{
		{2024, 5, 1, "5月1日の日記"},
		{2024, 5, 5, "5月5日の日記"},
		{2024, 5, 10, "5月10日の日記"},
		{2024, 6, 1, "6月1日の日記"},
	}
	for _, d := range dates {
		req := &g.CreateDiaryEntryRequest{
			Content: d.content,
			Date:    &g.YMD{Year: d.year, Month: d.month, Day: d.day},
		}
		if _, err := diaryService.CreateDiaryEntry(ctx, req); err != nil {
			t.Fatalf("日記作成失敗: %v", err)
		}
	}

	t.Run("正常系: 範囲内の日記を昇順で取得する", func(t *testing.T) {
		from := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 5, 5, 0, 0, 0, 0, time.UTC)
		result, err := diaryService.GetDiaryEntriesByDateRange(ctx, userID, from, to)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("期待件数 2 に対して %d 件取得", len(result))
		}
		if result[0].Content != "5月1日の日記" || result[1].Content != "5月5日の日記" {
			t.Errorf("日付昇順で正しい内容が返っていない: %+v", result)
		}
	})

	t.Run("正常系: 範囲外の日記は含まれない", func(t *testing.T) {
		from := time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 5, 4, 0, 0, 0, 0, time.UTC)
		result, err := diaryService.GetDiaryEntriesByDateRange(ctx, userID, from, to)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("範囲外の日記は返らないことを期待したが: %d 件取得", len(result))
		}
	})
}
