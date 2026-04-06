package database_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

// createUserLLMRow はテスト用にuser_llmsレコードを直接挿入する
func createUserLLMRow(t *testing.T, db *sql.DB, userID uuid.UUID, autoSummaryDaily, autoSummaryMonthly, autoLatestTrend, semanticSearch bool) {
	t.Helper()
	now := time.Now().Unix()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO user_llms (user_id, llm_provider, key, auto_summary_daily, auto_summary_monthly, auto_latest_trend_enabled, semantic_search_enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		userID, 1, "test-key", autoSummaryDaily, autoSummaryMonthly, autoLatestTrend, semanticSearch, now, now,
	)
	if err != nil {
		t.Fatalf("user_llmsレコードの挿入に失敗: %v", err)
	}
}

func TestUserIDsWithAutoSummaryDaily(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	userID1 := testutil.CreateTestUser(t, db, "user-llms-daily-1@example.com", "User1")
	userID2 := testutil.CreateTestUser(t, db, "user-llms-daily-2@example.com", "User2")

	createUserLLMRow(t, db, userID1, true, false, false, false)
	createUserLLMRow(t, db, userID2, false, false, false, false)

	t.Run("auto_summary_daily=trueのユーザーのみ返す", func(t *testing.T) {
		ids, err := database.UserIDsWithAutoSummaryDaily(ctx, db)
		if err != nil {
			t.Fatalf("UserIDsWithAutoSummaryDaily失敗: %v", err)
		}
		found := false
		for _, id := range ids {
			if id == userID1.String() {
				found = true
			}
			if id == userID2.String() {
				t.Errorf("auto_summary_daily=falseのユーザーが含まれている: %s", id)
			}
		}
		if !found {
			t.Errorf("auto_summary_daily=trueのユーザーが含まれていない")
		}
	})
}

func TestUserIDsWithAutoSummaryMonthly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	userID1 := testutil.CreateTestUser(t, db, "user-llms-monthly-1@example.com", "User1")
	userID2 := testutil.CreateTestUser(t, db, "user-llms-monthly-2@example.com", "User2")

	createUserLLMRow(t, db, userID1, false, true, false, false)
	createUserLLMRow(t, db, userID2, false, false, false, false)

	t.Run("auto_summary_monthly=trueのユーザーのみ返す", func(t *testing.T) {
		ids, err := database.UserIDsWithAutoSummaryMonthly(ctx, db)
		if err != nil {
			t.Fatalf("UserIDsWithAutoSummaryMonthly失敗: %v", err)
		}
		found := false
		for _, id := range ids {
			if id == userID1.String() {
				found = true
			}
			if id == userID2.String() {
				t.Errorf("auto_summary_monthly=falseのユーザーが含まれている: %s", id)
			}
		}
		if !found {
			t.Errorf("auto_summary_monthly=trueのユーザーが含まれていない")
		}
	})
}

func TestUserIDsWithAutoLatestTrendEnabled(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	userID1 := testutil.CreateTestUser(t, db, "user-llms-trend-1@example.com", "User1")
	userID2 := testutil.CreateTestUser(t, db, "user-llms-trend-2@example.com", "User2")

	createUserLLMRow(t, db, userID1, false, false, true, false)
	createUserLLMRow(t, db, userID2, false, false, false, false)

	t.Run("auto_latest_trend_enabled=trueのユーザーのみ返す", func(t *testing.T) {
		ids, err := database.UserIDsWithAutoLatestTrendEnabled(ctx, db)
		if err != nil {
			t.Fatalf("UserIDsWithAutoLatestTrendEnabled失敗: %v", err)
		}
		found := false
		for _, id := range ids {
			if id == userID1.String() {
				found = true
			}
			if id == userID2.String() {
				t.Errorf("auto_latest_trend_enabled=falseのユーザーが含まれている: %s", id)
			}
		}
		if !found {
			t.Errorf("auto_latest_trend_enabled=trueのユーザーが含まれていない")
		}
	})
}

func TestUserIDsWithSemanticSearchEnabled(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	userID1 := testutil.CreateTestUser(t, db, "user-llms-semantic-1@example.com", "User1")
	userID2 := testutil.CreateTestUser(t, db, "user-llms-semantic-2@example.com", "User2")

	createUserLLMRow(t, db, userID1, false, false, false, true)
	createUserLLMRow(t, db, userID2, false, false, false, false)

	t.Run("semantic_search_enabled=trueのユーザーのみ返す", func(t *testing.T) {
		ids, err := database.UserIDsWithSemanticSearchEnabled(ctx, db)
		if err != nil {
			t.Fatalf("UserIDsWithSemanticSearchEnabled失敗: %v", err)
		}
		found := false
		for _, id := range ids {
			if id == userID1.String() {
				found = true
			}
			if id == userID2.String() {
				t.Errorf("semantic_search_enabled=falseのユーザーが含まれている: %s", id)
			}
		}
		if !found {
			t.Errorf("semantic_search_enabled=trueのユーザーが含まれていない")
		}
	})
}
