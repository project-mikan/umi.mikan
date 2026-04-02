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

func insertTestDiary(t *testing.T, db *sql.DB, userID uuid.UUID, content string, date string) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UnixMilli()
	_, err := db.ExecContext(ctx, `INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, content, date, now, now,
	)
	if err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}
}

func TestDiariesByUserIDAndKeywords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "diaries-keywords-test@example.com", "DiariesKeywordsUser")
	ctx := context.Background()

	// テスト用日記を挿入
	insertTestDiary(t, db, userID, "今日は田中太郎と映画を観た", "2024-09-01")
	insertTestDiary(t, db, userID, "タナカが来てくれた", "2024-09-02")
	insertTestDiary(t, db, userID, "今日は読書をした", "2024-09-03")

	t.Run("空スライスは全件返す", func(t *testing.T) {
		result, err := database.DiariesByUserIDAndKeywords(ctx, db, userID.String(), []string{})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("期待件数 3 に対して %d 件取得", len(result))
		}
	})

	t.Run("1キーワードで一致する日記を返す", func(t *testing.T) {
		result, err := database.DiariesByUserIDAndKeywords(ctx, db, userID.String(), []string{"田中太郎"})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("期待件数 1 に対して %d 件取得", len(result))
		}
	})

	t.Run("複数キーワードでOR検索し一致する全日記を返す", func(t *testing.T) {
		result, err := database.DiariesByUserIDAndKeywords(ctx, db, userID.String(), []string{"田中太郎", "タナカ"})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("期待件数 2 に対して %d 件取得", len(result))
		}
	})

	t.Run("一致しないキーワードは空スライスを返す", func(t *testing.T) {
		result, err := database.DiariesByUserIDAndKeywords(ctx, db, userID.String(), []string{"存在しないキーワード"})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("一致なしで空スライスを期待したが: %d 件取得", len(result))
		}
	})

	t.Run("他ユーザーの日記はヒットしない", func(t *testing.T) {
		otherUserID := testutil.CreateTestUser(t, db, "other-diaries-test@example.com", "OtherUser")
		result, err := database.DiariesByUserIDAndKeywords(ctx, db, otherUserID.String(), []string{"田中太郎"})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("他ユーザーの日記が返らないことを期待したが: %d 件取得", len(result))
		}
	})
}
