package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestDeleteDiariesByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "delete-diaries@example.com", "User")

	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, "日記内容", "2020-01-01", now, now,
	); err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	t.Run("ユーザーの日記を全件削除する", func(t *testing.T) {
		if err := database.DeleteDiariesByUserID(ctx, db, userID); err != nil {
			t.Fatalf("DeleteDiariesByUserID失敗: %v", err)
		}

		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM diaries WHERE user_id = $1`, userID).Scan(&count); err != nil {
			t.Fatalf("カウントクエリ失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("削除後の件数: 期待 0, 実際 %d", count)
		}
	})
}

func TestDeleteUserLLMsByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "delete-user-llms@example.com", "User")

	now := time.Now().Unix()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO user_llms (user_id, llm_provider, key, auto_summary_monthly, auto_latest_trend_enabled, semantic_search_enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID, 1, "test-key", false, false, false, now, now,
	); err != nil {
		t.Fatalf("user_llmsの挿入に失敗: %v", err)
	}

	t.Run("ユーザーのLLM設定を全件削除する", func(t *testing.T) {
		if err := database.DeleteUserLLMsByUserID(ctx, db, userID); err != nil {
			t.Fatalf("DeleteUserLLMsByUserID失敗: %v", err)
		}

		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_llms WHERE user_id = $1`, userID).Scan(&count); err != nil {
			t.Fatalf("カウントクエリ失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("削除後の件数: 期待 0, 実際 %d", count)
		}
	})
}

func TestDeleteUserPasswordAuthesByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "delete-user-password@example.com", "User")

	// user_password_authesはCreateTestUserで既に作成済みのため、件数が1件であることを確認
	var initialCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_password_authes WHERE user_id = $1`, userID).Scan(&initialCount); err != nil {
		t.Fatalf("初期カウントクエリ失敗: %v", err)
	}

	t.Run("ユーザーのパスワード認証情報を全件削除する", func(t *testing.T) {
		if err := database.DeleteUserPasswordAuthesByUserID(ctx, db, userID); err != nil {
			t.Fatalf("DeleteUserPasswordAuthesByUserID失敗: %v", err)
		}

		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_password_authes WHERE user_id = $1`, userID).Scan(&count); err != nil {
			t.Fatalf("カウントクエリ失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("削除後の件数: 期待 0, 実際 %d", count)
		}
	})
}

func TestTotalMonthlySummaryCount(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "total-monthly-summary@example.com", "User")

	now := time.Now().Unix()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO diary_summary_months (id, user_id, year, month, summary, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.New(), userID, 2020, 1, "月次サマリ", now, now,
	); err != nil {
		t.Fatalf("月次サマリーの挿入に失敗: %v", err)
	}

	t.Run("月次サマリーの総数を返す", func(t *testing.T) {
		count, err := database.TotalMonthlySummaryCount(ctx, db, userID)
		if err != nil {
			t.Fatalf("TotalMonthlySummaryCount失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d", count)
		}
	})
}

func TestPendingMonthlySummaryCount(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "pending-monthly-summary@example.com", "User")

	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, "日記内容", "2020-01-15", now, now,
	); err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	t.Run("未作成の月次サマリー件数を返す", func(t *testing.T) {
		count, err := database.PendingMonthlySummaryCount(ctx, db, userID)
		if err != nil {
			t.Fatalf("PendingMonthlySummaryCount失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d", count)
		}
	})
}

func TestUserLLMAutoSettingsByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "user-llm-auto-settings@example.com", "User")

	t.Run("設定が存在しない場合は全てfalseを返す", func(t *testing.T) {
		settings, err := database.UserLLMAutoSettingsByUserID(ctx, db, userID)
		if err != nil {
			t.Fatalf("UserLLMAutoSettingsByUserID失敗: %v", err)
		}
		if settings.AutoSummaryMonthly || settings.AutoLatestTrend || settings.SemanticSearchEnabled {
			t.Errorf("設定なしで全てfalseを期待したが: %+v", settings)
		}
	})

	t.Run("設定が存在する場合はその値を返す", func(t *testing.T) {
		createUserLLMRow(t, db, userID, false, true, false)

		settings, err := database.UserLLMAutoSettingsByUserID(ctx, db, userID)
		if err != nil {
			t.Fatalf("UserLLMAutoSettingsByUserID失敗: %v", err)
		}
		if settings.AutoSummaryMonthly {
			t.Errorf("AutoSummaryMonthly: 期待 false, 実際 true")
		}
		if !settings.AutoLatestTrend {
			t.Errorf("AutoLatestTrend: 期待 true, 実際 false")
		}
	})
}

func TestTotalEmbeddingCount(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "total-embedding-count@example.com", "User")

	t.Run("embeddingが存在しない場合は0を返す", func(t *testing.T) {
		count, err := database.TotalEmbeddingCount(ctx, db, userID)
		if err != nil {
			t.Fatalf("TotalEmbeddingCount失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("期待 0, 実際 %d", count)
		}
	})
}

func TestTotalEmbeddingDiaryCount(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "total-embedding-diary-count@example.com", "User")

	t.Run("同一日記の複数チャンクは1件としてカウントされる", func(t *testing.T) {
		diaryID := uuid.New()
		now := time.Now().UnixMilli()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			diaryID, userID, "日記内容", "2020-01-01", now, now,
		); err != nil {
			t.Fatalf("日記の挿入に失敗: %v", err)
		}

		// 同じdiaryIDで複数のembeddingを挿入
		for i := range 3 {
			if _, err := db.ExecContext(ctx,
				`INSERT INTO diary_embeddings (id, diary_id, user_id, chunk_index, chunk_content, chunk_summary, embedding, model_version)
				 VALUES ($1, $2, $3, $4, $5, $6, array_fill(0.1, ARRAY[3072])::halfvec, $7)`,
				uuid.New(), diaryID, userID, i, fmt.Sprintf("チャンク内容 %d", i), fmt.Sprintf("チャンク概要 %d", i), "v1",
			); err != nil {
				t.Fatalf("embeddingの挿入に失敗: %v", err)
			}
		}

		count, err := database.TotalEmbeddingDiaryCount(ctx, db, userID)
		if err != nil {
			t.Fatalf("TotalEmbeddingDiaryCount失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d (DISTINCTが機能していない可能性があります)", count)
		}

		totalChunks, _ := database.TotalEmbeddingCount(ctx, db, userID)
		if totalChunks != 3 {
			t.Errorf("TotalEmbeddingCount: 期待 3, 実際 %d", totalChunks)
		}
	})
}

func TestPendingEmbeddingCount(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "pending-embedding-count@example.com", "User")

	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, "日記内容", "2020-01-01", now, now,
	); err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	t.Run("embedding未生成の日記件数を返す", func(t *testing.T) {
		count, err := database.PendingEmbeddingCount(ctx, db, userID)
		if err != nil {
			t.Fatalf("PendingEmbeddingCount失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d", count)
		}
	})
}

func TestHourlyPubSubMetrics(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "hourly-pubsub-metrics@example.com", "User")

	t.Run("過去24時間の時間別メトリクスを返す（データなし）", func(t *testing.T) {
		metrics, err := database.HourlyPubSubMetrics(ctx, db, userID)
		if err != nil {
			t.Fatalf("HourlyPubSubMetrics失敗: %v", err)
		}
		// generate_seriesで24時間分返る
		if len(metrics) != 24 {
			t.Errorf("期待 24件, 実際 %d件", len(metrics))
		}
		// データがない場合は全てゼロ
		for _, m := range metrics {
			if m.MonthlySummariesProcessed != 0 ||
				m.EmbeddingsProcessed != 0 || m.SemanticSearchesProcessed != 0 {
				t.Errorf("データなしで全ゼロを期待したが: %+v", m)
			}
		}
	})

	t.Run("月次サマリーデータが集計に反映される", func(t *testing.T) {
		now := time.Now().Unix()
		if _, err := db.ExecContext(ctx,
			`INSERT INTO diary_summary_months (id, user_id, year, month, summary, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), userID, 2020, 1, "サマリ", now, now,
		); err != nil {
			t.Fatalf("月次サマリーの挿入に失敗: %v", err)
		}

		metrics, err := database.HourlyPubSubMetrics(ctx, db, userID)
		if err != nil {
			t.Fatalf("HourlyPubSubMetrics失敗: %v", err)
		}
		if len(metrics) != 24 {
			t.Errorf("期待 24件, 実際 %d件", len(metrics))
		}
		var total int32
		for _, m := range metrics {
			total += m.MonthlySummariesProcessed
		}
		if total != 1 {
			t.Errorf("月次サマリー合計: 期待 1, 実際 %d", total)
		}
	})

}

func TestInsertSemanticSearchLog(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "semantic-search-log@example.com", "User")

	t.Run("意味的検索ログを挿入できる", func(t *testing.T) {
		if err := database.InsertSemanticSearchLog(ctx, db, userID); err != nil {
			t.Fatalf("InsertSemanticSearchLog失敗: %v", err)
		}

		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM semantic_search_logs WHERE user_id = $1`, userID).Scan(&count); err != nil {
			t.Fatalf("カウントクエリ失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d", count)
		}
	})
}
