package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestRelatedKeywordsByUserIDAndKeyword(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "entity-search-test@example.com", "EntitySearchUser")
	ctx := context.Background()
	now := time.Now().UnixMilli()

	// エンティティと複数エイリアスをDBに直接挿入
	entityID := uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO entities (id, user_id, created_at, updated_at, category_id, name) VALUES ($1, $2, $3, $4, $5, $6)`,
		entityID, userID, now, now, 0, "山田花子",
	)
	if err != nil {
		t.Fatalf("エンティティの挿入に失敗: %v", err)
	}
	for _, alias := range []string{"ハナ", "花ちゃん"} {
		_, err = db.ExecContext(ctx, `INSERT INTO entity_aliases (id, entity_id, created_at, updated_at, alias) VALUES ($1, $2, $3, $4, $5)`,
			uuid.New(), entityID, now, now, alias,
		)
		if err != nil {
			t.Fatalf("エイリアスの挿入に失敗: %v", err)
		}
	}

	t.Run("空白キーワードはnilを返す", func(t *testing.T) {
		result, err := database.RelatedKeywordsByUserIDAndKeyword(ctx, db, userID.String(), "")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if result != nil {
			t.Errorf("空白キーワードはnilを期待したが: %v", result)
		}
	})

	t.Run("エンティティ名に完全一致するキーワードでエイリアスを全て返す", func(t *testing.T) {
		result, err := database.RelatedKeywordsByUserIDAndKeyword(ctx, db, userID.String(), "山田花子")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		// エイリアス「ハナ」「花ちゃん」の2件が返ること
		if len(result) != 2 {
			t.Errorf("期待件数 2 に対して %d 件取得: %v", len(result), result)
		}
	})

	t.Run("エイリアスに完全一致するキーワードでエンティティ名と他エイリアスを返す", func(t *testing.T) {
		result, err := database.RelatedKeywordsByUserIDAndKeyword(ctx, db, userID.String(), "ハナ")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		// 「山田花子」と「花ちゃん」の2件が返ること（「ハナ」自身は除外）
		if len(result) != 2 {
			t.Errorf("期待件数 2 に対して %d 件取得: %v", len(result), result)
		}
	})

	t.Run("一致しないキーワードは空スライスを返す", func(t *testing.T) {
		result, err := database.RelatedKeywordsByUserIDAndKeyword(ctx, db, userID.String(), "存在しないキーワード")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("一致なしで空スライスを期待したが: %v", result)
		}
	})

	t.Run("他ユーザーのエンティティはヒットしない", func(t *testing.T) {
		otherUserID := testutil.CreateTestUser(t, db, "other-entity-test@example.com", "OtherUser")
		result, err := database.RelatedKeywordsByUserIDAndKeyword(ctx, db, otherUserID.String(), "山田花子")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("他ユーザーのエンティティが返らないことを期待したが: %v", result)
		}
	})
}
