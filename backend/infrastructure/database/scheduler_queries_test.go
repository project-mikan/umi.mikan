package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestMonthsNeedingMonthlySummary(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "scheduler-monthly-summary@example.com", "User")

	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, "日記内容", "2020-03-15", now, now,
	); err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	t.Run("月次サマリ未生成の月を返す", func(t *testing.T) {
		months, err := database.MonthsNeedingMonthlySummary(ctx, db, userID.String())
		if err != nil {
			t.Fatalf("MonthsNeedingMonthlySummary失敗: %v", err)
		}
		if len(months) != 1 {
			t.Errorf("期待件数 1 に対して %d", len(months))
		}
		if months[0].Year != 2020 || months[0].Month != 3 {
			t.Errorf("期待 2020/3 に対して %d/%d", months[0].Year, months[0].Month)
		}
	})

	t.Run("月次サマリ生成済みの月は返さない", func(t *testing.T) {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO diary_summary_months (id, user_id, year, month, summary, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), userID, 2020, 3, "月次サマリ内容", now+1, now+1,
		); err != nil {
			t.Fatalf("月次サマリの挿入に失敗: %v", err)
		}

		months, err := database.MonthsNeedingMonthlySummary(ctx, db, userID.String())
		if err != nil {
			t.Fatalf("MonthsNeedingMonthlySummary失敗: %v", err)
		}
		for _, m := range months {
			if m.Year == 2020 && m.Month == 3 {
				t.Errorf("月次サマリ生成済みの月が含まれている: 2020/3")
			}
		}
	})
}

func TestDiaryCountInDateRange(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "scheduler-count-range@example.com", "User")

	now := time.Now().UnixMilli()
	for _, date := range []string{"2020-05-01", "2020-05-02", "2020-05-03"} {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), userID, "内容", date, now, now,
		); err != nil {
			t.Fatalf("日記の挿入に失敗: %v", err)
		}
	}

	from := time.Date(2020, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2020, 5, 3, 0, 0, 0, 0, time.UTC)

	t.Run("範囲内の日記件数を返す", func(t *testing.T) {
		count, err := database.DiaryCountInDateRange(ctx, db, userID.String(), from, to)
		if err != nil {
			t.Fatalf("DiaryCountInDateRange失敗: %v", err)
		}
		if count != 3 {
			t.Errorf("期待件数 3 に対して %d", count)
		}
	})

	t.Run("範囲外の日記は含まない", func(t *testing.T) {
		narrowFrom := time.Date(2020, 5, 2, 0, 0, 0, 0, time.UTC)
		narrowTo := time.Date(2020, 5, 2, 0, 0, 0, 0, time.UTC)
		count, err := database.DiaryCountInDateRange(ctx, db, userID.String(), narrowFrom, narrowTo)
		if err != nil {
			t.Fatalf("DiaryCountInDateRange失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待件数 1 に対して %d", count)
		}
	})
}

func TestDiaryCountInMonth(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "diary-count-in-month@example.com", "User")

	now := time.Now().UnixMilli()
	for _, date := range []string{"2020-01-10", "2020-01-20"} {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.New(), userID, "日記内容", date, now, now,
		); err != nil {
			t.Fatalf("日記の挿入に失敗: %v", err)
		}
	}

	t.Run("指定月の日記件数を返す", func(t *testing.T) {
		count, err := database.DiaryCountInMonth(ctx, db, userID.String(), 2020, 1)
		if err != nil {
			t.Fatalf("DiaryCountInMonth失敗: %v", err)
		}
		if count != 2 {
			t.Errorf("期待 2, 実際 %d", count)
		}
	})

	t.Run("日記が存在しない月は0を返す", func(t *testing.T) {
		count, err := database.DiaryCountInMonth(ctx, db, userID.String(), 2020, 2)
		if err != nil {
			t.Fatalf("DiaryCountInMonth失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("期待 0, 実際 %d", count)
		}
	})
}

func TestDiaryIDsNeedingEmbedding(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "scheduler-embedding@example.com", "User")

	now := time.Now().UnixMilli()
	targetDate := "2020-06-01"
	diaryID := uuid.New()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		diaryID, userID, "embedding対象の日記内容", targetDate, now, now,
	); err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	target := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)

	t.Run("embedding未生成の日記IDを返す", func(t *testing.T) {
		ids, err := database.DiaryIDsNeedingEmbedding(ctx, db, userID.String(), target)
		if err != nil {
			t.Fatalf("DiaryIDsNeedingEmbedding失敗: %v", err)
		}
		if len(ids) != 1 {
			t.Errorf("期待件数 1 に対して %d", len(ids))
		}
		if ids[0] != diaryID.String() {
			t.Errorf("期待ID %s に対して %s", diaryID, ids[0])
		}
	})

	t.Run("embedding生成済みで最新の日記は返さない", func(t *testing.T) {
		dummyEmbedding := make([]float32, 3072)
		chunks := []database.DiaryChunk{
			{Index: 0, Content: "内容", Summary: "概要", Embedding: dummyEmbedding, SplitModelVersion: "v1"},
		}
		if err := database.UpsertDiaryChunkEmbeddings(ctx, db, diaryID, userID, chunks, "gemini-embedding-001"); err != nil {
			t.Fatalf("UpsertDiaryChunkEmbeddings失敗: %v", err)
		}

		ids, err := database.DiaryIDsNeedingEmbedding(ctx, db, userID.String(), target)
		if err != nil {
			t.Fatalf("DiaryIDsNeedingEmbedding失敗: %v", err)
		}
		for _, id := range ids {
			if id == diaryID.String() {
				t.Errorf("embedding生成済みの日記IDが含まれている: %s", id)
			}
		}
	})
}
