package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

// insertTestAPIKey はテスト用のAPIキー行を挿入する。
// key_hashはグローバルにユニークな制約があるため、実行ごとに衝突しないランダム値を使う。
func insertTestAPIKey(t *testing.T, db *sql.DB, userID uuid.UUID) uuid.UUID {
	t.Helper()
	key := &database.UserAPIKey{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "テストキー",
		KeyHash:   "test-hash-" + uuid.New().String(),
		KeyPrefix: "umi_test1234",
		ExpiresAt: 1800000000, // 十分未来のUnix秒（テスト実行時点では期限切れにならない値）
		CreatedAt: 1700000000,
		UpdatedAt: 1700000000,
	}
	if err := key.Insert(context.Background(), db); err != nil {
		t.Fatalf("APIキーの挿入に失敗: %v", err)
	}
	return key.ID
}

func TestUpdateUserAPIKeyLastUsed(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "api-key-last-used@example.com", "APIKeyLastUsedUser")
	ctx := context.Background()

	t.Run("正常系: 最終使用日時が更新される", func(t *testing.T) {
		keyID := insertTestAPIKey(t, db, userID)
		if err := database.UpdateUserAPIKeyLastUsed(ctx, db, keyID, 1700001234); err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}

		updated, err := database.UserAPIKeyByID(ctx, db, keyID)
		if err != nil {
			t.Fatalf("APIキーの取得に失敗: %v", err)
		}
		if !updated.LastUsedAt.Valid || updated.LastUsedAt.Int64 != 1700001234 {
			t.Errorf("LastUsedAt: 期待 1700001234, 実際 %+v", updated.LastUsedAt)
		}
		if updated.UpdatedAt != 1700001234 {
			t.Errorf("UpdatedAt: 期待 1700001234, 実際 %d", updated.UpdatedAt)
		}
	})

	t.Run("正常系: 存在しないIDでもエラーにならない", func(t *testing.T) {
		if err := database.UpdateUserAPIKeyLastUsed(ctx, db, uuid.New(), 1700001234); err != nil {
			t.Errorf("存在しないIDでエラーが返った: %v", err)
		}
	})
}

func TestUpdateUserAPIKeyLastUsed_DBError(t *testing.T) {
	t.Run("異常系: DBがクローズされている場合はエラー", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		ctx := context.Background()

		// DBを閉じてクエリエラーを発生させる
		if err := db.Close(); err != nil {
			t.Fatalf("DB クローズに失敗: %v", err)
		}

		if err := database.UpdateUserAPIKeyLastUsed(ctx, db, uuid.New(), 1700001234); err == nil {
			t.Fatal("DBエラー時にエラーが返ることを期待したがnilが返った")
		}
	})
}
