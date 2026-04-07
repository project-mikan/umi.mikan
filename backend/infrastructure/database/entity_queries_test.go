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

// insertEntity はテスト用エンティティをDBに直接挿入しIDを返す
func insertEntity(t *testing.T, db *sql.DB, ctx context.Context, userID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO entities (id, user_id, created_at, updated_at, category_id, name) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, userID, now, now, 0, name,
	); err != nil {
		t.Fatalf("エンティティの挿入に失敗: %v", err)
	}
	return id
}

// insertAlias はテスト用エイリアスをDBに直接挿入しIDを返す
func insertAlias(t *testing.T, db *sql.DB, ctx context.Context, entityID uuid.UUID, alias string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO entity_aliases (id, entity_id, created_at, updated_at, alias) VALUES ($1, $2, $3, $4, $5)`,
		id, entityID, now, now, alias,
	); err != nil {
		t.Fatalf("エイリアスの挿入に失敗: %v", err)
	}
	return id
}

func TestAliasesByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "aliases-by-user-id@example.com", "User")

	entityID := insertEntity(t, db, ctx, userID, "山田花子")
	insertAlias(t, db, ctx, entityID, "ハナ")
	insertAlias(t, db, ctx, entityID, "花ちゃん")

	t.Run("ユーザーの全エイリアスをentityIDをキーに返す", func(t *testing.T) {
		aliasMap, err := database.AliasesByUserID(ctx, db, userID)
		if err != nil {
			t.Fatalf("AliasesByUserID失敗: %v", err)
		}
		aliases := aliasMap[entityID.String()]
		if len(aliases) != 2 {
			t.Errorf("期待 2件, 実際 %d件", len(aliases))
		}
	})

	t.Run("他ユーザーのエイリアスは含まない", func(t *testing.T) {
		otherUserID := testutil.CreateTestUser(t, db, "aliases-other-user@example.com", "Other")
		aliasMap, err := database.AliasesByUserID(ctx, db, otherUserID)
		if err != nil {
			t.Fatalf("AliasesByUserID失敗: %v", err)
		}
		if len(aliasMap) != 0 {
			t.Errorf("他ユーザーのエイリアスが含まれている")
		}
	})
}

func TestCountAliasMatchingName(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "count-alias-matching-name@example.com", "User")

	entityID := insertEntity(t, db, ctx, userID, "山田花子")
	insertAlias(t, db, ctx, entityID, "ハナ")

	t.Run("エイリアスとして使用中の名前は1以上を返す", func(t *testing.T) {
		count, err := database.CountAliasMatchingName(ctx, db, userID, "ハナ")
		if err != nil {
			t.Fatalf("CountAliasMatchingName失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d", count)
		}
	})

	t.Run("使用されていない名前は0を返す", func(t *testing.T) {
		count, err := database.CountAliasMatchingName(ctx, db, userID, "存在しない")
		if err != nil {
			t.Fatalf("CountAliasMatchingName失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("期待 0, 実際 %d", count)
		}
	})
}

func TestCountEntityMatchingAlias(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "count-entity-matching-alias@example.com", "User")

	insertEntity(t, db, ctx, userID, "山田花子")

	t.Run("エンティティ名として使用中の名前は1以上を返す", func(t *testing.T) {
		count, err := database.CountEntityMatchingAlias(ctx, db, userID, "山田花子")
		if err != nil {
			t.Fatalf("CountEntityMatchingAlias失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d", count)
		}
	})

	t.Run("使用されていない名前は0を返す", func(t *testing.T) {
		count, err := database.CountEntityMatchingAlias(ctx, db, userID, "存在しない")
		if err != nil {
			t.Fatalf("CountEntityMatchingAlias失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("期待 0, 実際 %d", count)
		}
	})
}

func TestCountAliasDuplicate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "count-alias-duplicate@example.com", "User")

	entityID := insertEntity(t, db, ctx, userID, "山田花子")
	insertAlias(t, db, ctx, entityID, "ハナ")

	t.Run("既存エイリアスと重複する場合は1以上を返す", func(t *testing.T) {
		count, err := database.CountAliasDuplicate(ctx, db, userID, "ハナ")
		if err != nil {
			t.Fatalf("CountAliasDuplicate失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("期待 1, 実際 %d", count)
		}
	})

	t.Run("重複しない場合は0を返す", func(t *testing.T) {
		count, err := database.CountAliasDuplicate(ctx, db, userID, "存在しない")
		if err != nil {
			t.Fatalf("CountAliasDuplicate失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("期待 0, 実際 %d", count)
		}
	})
}

func TestCountAliasDuplicateExcluding(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "count-alias-duplicate-excluding@example.com", "User")

	entityID := insertEntity(t, db, ctx, userID, "山田花子")
	aliasID := insertAlias(t, db, ctx, entityID, "ハナ")
	insertAlias(t, db, ctx, entityID, "花ちゃん")

	t.Run("自分自身を除外した場合、自分のエイリアスは重複とみなさない", func(t *testing.T) {
		count, err := database.CountAliasDuplicateExcluding(ctx, db, userID, "ハナ", aliasID)
		if err != nil {
			t.Fatalf("CountAliasDuplicateExcluding失敗: %v", err)
		}
		if count != 0 {
			t.Errorf("自己除外で 0 を期待したが %d", count)
		}
	})

	t.Run("他のエイリアスとの重複は検出する", func(t *testing.T) {
		count, err := database.CountAliasDuplicateExcluding(ctx, db, userID, "花ちゃん", aliasID)
		if err != nil {
			t.Fatalf("CountAliasDuplicateExcluding失敗: %v", err)
		}
		if count != 1 {
			t.Errorf("他エイリアスとの重複で 1 を期待したが %d", count)
		}
	})
}

func TestSearchEntitiesByQuery(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	userID := testutil.CreateTestUser(t, db, "search-entities-by-query@example.com", "User")

	entityID := insertEntity(t, db, ctx, userID, "山田花子")
	insertAlias(t, db, ctx, entityID, "ハナ")
	insertEntity(t, db, ctx, userID, "田中太郎")

	t.Run("クエリ空で全件返す", func(t *testing.T) {
		entities, err := database.SearchEntitiesByQuery(ctx, db, userID, "")
		if err != nil {
			t.Fatalf("SearchEntitiesByQuery失敗: %v", err)
		}
		if len(entities) != 2 {
			t.Errorf("期待 2件, 実際 %d件", len(entities))
		}
	})

	t.Run("エンティティ名で部分一致検索する", func(t *testing.T) {
		entities, err := database.SearchEntitiesByQuery(ctx, db, userID, "山田")
		if err != nil {
			t.Fatalf("SearchEntitiesByQuery失敗: %v", err)
		}
		if len(entities) != 1 {
			t.Errorf("期待 1件, 実際 %d件", len(entities))
		}
		if entities[0].Name != "山田花子" {
			t.Errorf("期待 山田花子, 実際 %s", entities[0].Name)
		}
	})

	t.Run("エイリアスで部分一致検索する", func(t *testing.T) {
		entities, err := database.SearchEntitiesByQuery(ctx, db, userID, "ハナ")
		if err != nil {
			t.Fatalf("SearchEntitiesByQuery失敗: %v", err)
		}
		if len(entities) != 1 {
			t.Errorf("期待 1件, 実際 %d件", len(entities))
		}
	})

	t.Run("他ユーザーのエンティティはヒットしない", func(t *testing.T) {
		otherUserID := testutil.CreateTestUser(t, db, "search-entities-other@example.com", "Other")
		entities, err := database.SearchEntitiesByQuery(ctx, db, otherUserID, "")
		if err != nil {
			t.Fatalf("SearchEntitiesByQuery失敗: %v", err)
		}
		if len(entities) != 0 {
			t.Errorf("他ユーザーのエンティティが含まれている")
		}
	})

	t.Run("LIKEメタキャラクターがエスケープされる", func(t *testing.T) {
		// % を含む名前のエンティティを登録
		specialID := insertEntity(t, db, ctx, userID, "100%満足")
		defer func() {
			_, _ = db.ExecContext(ctx, "DELETE FROM entity_aliases WHERE entity_id = $1", specialID)
			_, _ = db.ExecContext(ctx, "DELETE FROM entities WHERE id = $1", specialID)
		}()

		// "%" で検索して "100%満足" がヒットし、"山田花子" などはヒットしないことを確認
		entities, err := database.SearchEntitiesByQuery(ctx, db, userID, "%")
		if err != nil {
			t.Fatalf("SearchEntitiesByQuery失敗: %v", err)
		}
		if len(entities) != 1 {
			t.Errorf("期待 1件 (%s), 実際 %d件", "100%満足", len(entities))
		} else if entities[0].Name != "100%満足" {
			t.Errorf("期待 100%%満足, 実際 %s", entities[0].Name)
		}
	})
}
