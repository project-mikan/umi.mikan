package database_test

import (
	"context"
	"testing"

	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestDiariesByUserIDAndDateRange(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "diary-export-test@example.com", "DiaryExportUser")
	ctx := context.Background()

	// テスト用日記を挿入（期間をまたぐ複数月）
	insertTestDiary(t, db, userID, "2024年1月の日記", "2024-01-15")
	insertTestDiary(t, db, userID, "2024年3月の日記", "2024-03-10")
	insertTestDiary(t, db, userID, "2024年6月の日記", "2024-06-20")
	insertTestDiary(t, db, userID, "2024年12月の日記", "2024-12-31")
	insertTestDiary(t, db, userID, "2025年1月の日記", "2025-01-01")

	t.Run("正常系: 期間内の全日記を昇順で返す", func(t *testing.T) {
		// 2024年3月〜2024年6月
		result, err := database.DiariesByUserIDAndDateRange(ctx, db, userID.String(), 2024, 3, 2024, 6)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("期待件数 2 に対して %d 件取得", len(result))
		}
		// 日付昇順であることを確認
		if len(result) == 2 && result[0].Date.After(result[1].Date) {
			t.Errorf("日付昇順になっていない: %v > %v", result[0].Date, result[1].Date)
		}
	})

	t.Run("正常系: 月初と月末を含む境界値が正しく取得される", func(t *testing.T) {
		// 2024年1月〜2024年1月（1ヶ月）
		result, err := database.DiariesByUserIDAndDateRange(ctx, db, userID.String(), 2024, 1, 2024, 1)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("期待件数 1 に対して %d 件取得", len(result))
		}
	})

	t.Run("正常系: 期間外の日記は返さない", func(t *testing.T) {
		// 2024年7月〜2024年11月（日記なし）
		result, err := database.DiariesByUserIDAndDateRange(ctx, db, userID.String(), 2024, 7, 2024, 11)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("期間外の日記は返さないことを期待したが: %d 件取得", len(result))
		}
	})

	t.Run("正常系: 跨年期間を正しく取得する", func(t *testing.T) {
		// 2024年12月〜2025年1月
		result, err := database.DiariesByUserIDAndDateRange(ctx, db, userID.String(), 2024, 12, 2025, 1)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("期待件数 2 に対して %d 件取得", len(result))
		}
	})

	t.Run("異常系: 他ユーザーの日記はヒットしない", func(t *testing.T) {
		otherUserID := testutil.CreateTestUser(t, db, "other-export-test@example.com", "OtherExportUser")
		result, err := database.DiariesByUserIDAndDateRange(ctx, db, otherUserID.String(), 2024, 1, 2025, 12)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("他ユーザーの日記が返らないことを期待したが: %d 件取得", len(result))
		}
	})
}
