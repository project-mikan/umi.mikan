package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestDiariesByUserIDAndDateRangeDays(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "diary-range-test@example.com", "DiaryRangeUser")
	ctx := context.Background()

	insertTestDiary(t, db, userID, "1日目の日記", "2024-05-01")
	insertTestDiary(t, db, userID, "5日目の日記", "2024-05-05")
	insertTestDiary(t, db, userID, "10日目の日記", "2024-05-10")
	insertTestDiary(t, db, userID, "翌月の日記", "2024-06-01")

	t.Run("正常系: 日単位の範囲内の日記を昇順で返す", func(t *testing.T) {
		from := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 5, 5, 0, 0, 0, 0, time.UTC)
		result, err := database.DiariesByUserIDAndDateRangeDays(ctx, db, userID.String(), from, to)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("期待件数 2 に対して %d 件取得", len(result))
		}
		if result[0].Content != "1日目の日記" || result[1].Content != "5日目の日記" {
			t.Errorf("日付昇順で正しい内容が返っていない: %+v", result)
		}
	})

	t.Run("正常系: 単一日を指定した場合その日のみ返す", func(t *testing.T) {
		from := time.Date(2024, 5, 10, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 5, 10, 0, 0, 0, 0, time.UTC)
		result, err := database.DiariesByUserIDAndDateRangeDays(ctx, db, userID.String(), from, to)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("期待件数 1 に対して %d 件取得", len(result))
		}
	})

	t.Run("正常系: 範囲外の日記は返さない", func(t *testing.T) {
		from := time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 5, 4, 0, 0, 0, 0, time.UTC)
		result, err := database.DiariesByUserIDAndDateRangeDays(ctx, db, userID.String(), from, to)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("範囲外の日記は返さないことを期待したが: %d 件取得", len(result))
		}
	})

	t.Run("異常系: 他ユーザーの日記はヒットしない", func(t *testing.T) {
		otherUserID := testutil.CreateTestUser(t, db, "other-range-test@example.com", "OtherRangeUser")
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		result, err := database.DiariesByUserIDAndDateRangeDays(ctx, db, otherUserID.String(), from, to)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("他ユーザーの日記が返らないことを期待したが: %d 件取得", len(result))
		}
	})
}

func TestDiariesByUserIDAndDateRangeDays_DBError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	if err := db.Close(); err != nil {
		t.Fatalf("DB クローズに失敗: %v", err)
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	_, err := database.DiariesByUserIDAndDateRangeDays(ctx, db, "00000000-0000-0000-0000-000000000001", from, to)
	if err == nil {
		t.Fatal("DBエラー時にエラーが返ることを期待したがnilが返った")
	}
}
