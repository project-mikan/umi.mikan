package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestUpsertMonthlySummaryError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "monthly-summary-error@example.com", "User")

	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, "日記内容", "2020-04-15", now, now,
	); err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	t.Run("エラーレコードを新規作成できる", func(t *testing.T) {
		err := database.UpsertMonthlySummaryError(ctx, db, userID, 2020, 4, "PROHIBITED_CONTENT")
		if err != nil {
			t.Fatalf("UpsertMonthlySummaryError失敗: %v", err)
		}

		summary, err := database.DiarySummaryMonthByUserIDYearMonth(ctx, db, userID, 2020, 4)
		if err != nil {
			t.Fatalf("DiarySummaryMonthByUserIDYearMonth失敗: %v", err)
		}
		if !summary.ErrorReason.Valid {
			t.Error("error_reasonがNULLになっている")
		}
		if summary.ErrorReason.String != "PROHIBITED_CONTENT" {
			t.Errorf("期待 PROHIBITED_CONTENT, 実際 %s", summary.ErrorReason.String)
		}
		if summary.Summary != "" {
			t.Errorf("エラー時のsummaryは空文字であるべき, 実際 %s", summary.Summary)
		}
	})

	t.Run("エラーレコードが既存の場合は上書きする", func(t *testing.T) {
		err := database.UpsertMonthlySummaryError(ctx, db, userID, 2020, 4, "OTHER_ERROR")
		if err != nil {
			t.Fatalf("UpsertMonthlySummaryError失敗: %v", err)
		}

		summary, err := database.DiarySummaryMonthByUserIDYearMonth(ctx, db, userID, 2020, 4)
		if err != nil {
			t.Fatalf("DiarySummaryMonthByUserIDYearMonth失敗: %v", err)
		}
		if summary.ErrorReason.String != "OTHER_ERROR" {
			t.Errorf("期待 OTHER_ERROR, 実際 %s", summary.ErrorReason.String)
		}
	})

	t.Run("既存サマリーがある場合はsummaryを保持してerror_reasonのみ更新する", func(t *testing.T) {
		// 別の月(5月)に有効なサマリーを事前挿入
		summaryNow := time.Now().UnixMilli()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), userID, "5月の日記", "2020-05-10", summaryNow, summaryNow,
		); err != nil {
			t.Fatalf("日記の挿入に失敗: %v", err)
		}
		if _, err := db.ExecContext(ctx,
			`INSERT INTO diary_summary_months (id, user_id, year, month, summary, model_version, created_at, updated_at)
			 VALUES ($1, $2, 2020, 5, '5月のまとめ', 'gemini-2.5-flash-lite', $3, $4)`,
			uuid.New(), userID, summaryNow, summaryNow,
		); err != nil {
			t.Fatalf("サマリーの挿入に失敗: %v", err)
		}

		err := database.UpsertMonthlySummaryError(ctx, db, userID, 2020, 5, "PROHIBITED_CONTENT")
		if err != nil {
			t.Fatalf("UpsertMonthlySummaryError失敗: %v", err)
		}

		summary, err := database.DiarySummaryMonthByUserIDYearMonth(ctx, db, userID, 2020, 5)
		if err != nil {
			t.Fatalf("DiarySummaryMonthByUserIDYearMonth失敗: %v", err)
		}
		if summary.Summary != "5月のまとめ" {
			t.Errorf("既存サマリーが保持されるべき, 実際 %q", summary.Summary)
		}
		if summary.ErrorReason.String != "PROHIBITED_CONTENT" {
			t.Errorf("期待 PROHIBITED_CONTENT, 実際 %s", summary.ErrorReason.String)
		}
	})

	t.Run("エラーがあった月はスケジューラー対象外になる", func(t *testing.T) {
		months, err := database.MonthsNeedingMonthlySummary(ctx, db, userID.String())
		if err != nil {
			t.Fatalf("MonthsNeedingMonthlySummary失敗: %v", err)
		}
		for _, m := range months {
			if m.Year == 2020 && m.Month == 4 {
				t.Error("エラー済みの月がスケジューラー対象に含まれている")
			}
		}
	})
}
