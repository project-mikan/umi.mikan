package entity

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateEntity(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	service := &EntityEntry{DB: db}

	// テスト用ユーザーを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	// コンテキストにユーザーIDを設定
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	t.Run("正常にエンティティを作成できる", func(t *testing.T) {
		req := &g.CreateEntityRequest{
			Name:     "テストエンティティ",
			Category: g.EntityCategory_PEOPLE,
			Memo:     "テストメモ",
		}

		resp, err := service.CreateEntity(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Entity.Id)
		assert.Equal(t, "テストエンティティ", resp.Entity.Name)
		assert.Equal(t, g.EntityCategory_PEOPLE, resp.Entity.Category)
		assert.Equal(t, "テストメモ", resp.Entity.Memo)
	})

	t.Run("同じ名前のエンティティは作成できない", func(t *testing.T) {
		// 最初のエンティティを作成
		req1 := &g.CreateEntityRequest{
			Name:     "重複テスト",
			Category: g.EntityCategory_PEOPLE,
		}
		_, err := service.CreateEntity(ctx, req1)
		require.NoError(t, err)

		// 同じ名前で作成を試みる
		req2 := &g.CreateEntityRequest{
			Name:     "重複テスト",
			Category: g.EntityCategory_PEOPLE,
		}
		_, err = service.CreateEntity(ctx, req2)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("既存のエイリアスと同じ名前のエンティティは作成できない", func(t *testing.T) {
		// エンティティを作成
		entity := &database.Entity{
			ID:         uuid.New(),
			UserID:     userID,
			Name:       "元のエンティティ",
			CategoryID: 1,
			CreatedAt:  time.Now().Unix(),
			UpdatedAt:  time.Now().Unix(),
		}
		require.NoError(t, entity.Insert(context.Background(), db))

		// エイリアスを作成
		alias := &database.EntityAlias{
			ID:        uuid.New(),
			EntityID:  entity.ID,
			Alias:     "既存エイリアス",
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		require.NoError(t, alias.Insert(context.Background(), db))

		// エイリアスと同じ名前でエンティティ作成を試みる
		req := &g.CreateEntityRequest{
			Name:     "既存エイリアス",
			Category: g.EntityCategory_PEOPLE,
		}
		_, err := service.CreateEntity(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})
}

func TestCreateEntityAlias(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	service := &EntityEntry{DB: db}

	// テスト用ユーザーとエンティティを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	entity := &database.Entity{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       "テストエンティティ",
		CategoryID: 1,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}
	require.NoError(t, entity.Insert(context.Background(), db))

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	t.Run("正常にエイリアスを追加できる", func(t *testing.T) {
		req := &g.CreateEntityAliasRequest{
			EntityId: entity.ID.String(),
			Alias:    "新しいエイリアス",
		}

		resp, err := service.CreateEntityAlias(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Alias.Id)
		assert.Equal(t, entity.ID.String(), resp.Alias.EntityId)
		assert.Equal(t, "新しいエイリアス", resp.Alias.Alias)
	})

	t.Run("同じエイリアスは重複登録できない", func(t *testing.T) {
		// 最初のエイリアスを作成
		req1 := &g.CreateEntityAliasRequest{
			EntityId: entity.ID.String(),
			Alias:    "重複エイリアス",
		}
		_, err := service.CreateEntityAlias(ctx, req1)
		require.NoError(t, err)

		// 同じエイリアスで登録を試みる
		req2 := &g.CreateEntityAliasRequest{
			EntityId: entity.ID.String(),
			Alias:    "重複エイリアス",
		}
		_, err = service.CreateEntityAlias(ctx, req2)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("既存のエンティティ名と同じエイリアスは作成できない", func(t *testing.T) {
		// 別のエンティティを作成
		entity2 := &database.Entity{
			ID:         uuid.New(),
			UserID:     userID,
			Name:       "別のエンティティ",
			CategoryID: 1,
			CreatedAt:  time.Now().Unix(),
			UpdatedAt:  time.Now().Unix(),
		}
		require.NoError(t, entity2.Insert(context.Background(), db))

		// 既存のエンティティ名と同じエイリアスを作成しようとする
		req := &g.CreateEntityAliasRequest{
			EntityId: entity.ID.String(),
			Alias:    "別のエンティティ",
		}
		_, err := service.CreateEntityAlias(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})
}

func TestDeleteEntity(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	service := &EntityEntry{DB: db}

	// テスト用ユーザーとエンティティを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	entity := &database.Entity{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       "削除テスト",
		CategoryID: 1,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}
	require.NoError(t, entity.Insert(context.Background(), db))

	// エイリアスも作成
	alias := &database.EntityAlias{
		ID:        uuid.New(),
		EntityID:  entity.ID,
		Alias:     "削除テストエイリアス",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, alias.Insert(context.Background(), db))

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	t.Run("エンティティ削除時にエイリアスもカスケード削除される", func(t *testing.T) {
		// エンティティを削除
		req := &g.DeleteEntityRequest{
			Id: entity.ID.String(),
		}
		resp, err := service.DeleteEntity(ctx, req)
		require.NoError(t, err)
		assert.True(t, resp.Success)

		// エンティティが削除されていることを確認
		_, err = database.EntityByID(context.Background(), db, entity.ID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// エイリアスも削除されていることを確認
		_, err = database.EntityAliasByID(context.Background(), db, alias.ID)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})

	t.Run("他のユーザーのエンティティは削除できない", func(t *testing.T) {
		// 別のユーザーとエンティティを作成
		otherUserID := uuid.New()
		otherUser := &database.User{
			ID:        otherUserID,
			Email:     fmt.Sprintf("other-%s@example.com", otherUserID.String()),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		require.NoError(t, otherUser.Insert(context.Background(), db))

		otherEntity := &database.Entity{
			ID:         uuid.New(),
			UserID:     otherUserID,
			Name:       "他人のエンティティ",
			CategoryID: 1,
			CreatedAt:  time.Now().Unix(),
			UpdatedAt:  time.Now().Unix(),
		}
		require.NoError(t, otherEntity.Insert(context.Background(), db))

		// 元のユーザーのコンテキストで他人のエンティティ削除を試みる
		req := &g.DeleteEntityRequest{
			Id: otherEntity.ID.String(),
		}
		_, err := service.DeleteEntity(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})
}

func TestListEntities(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	service := &EntityEntry{DB: db}

	// テスト用ユーザーを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	// 複数のエンティティとエイリアスを作成
	entity1 := &database.Entity{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       "人物1",
		CategoryID: 1,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}
	require.NoError(t, entity1.Insert(context.Background(), db))

	alias1 := &database.EntityAlias{
		ID:        uuid.New(),
		EntityID:  entity1.ID,
		Alias:     "人物1のエイリアス",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, alias1.Insert(context.Background(), db))

	entity2 := &database.Entity{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       "未分類エンティティ",
		CategoryID: 0,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}
	require.NoError(t, entity2.Insert(context.Background(), db))

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	t.Run("全エンティティを取得できる", func(t *testing.T) {
		req := &g.ListEntitiesRequest{
			AllCategories: true,
		}

		resp, err := service.ListEntities(ctx, req)
		require.NoError(t, err)
		assert.Len(t, resp.Entities, 2)
	})

	t.Run("カテゴリでフィルタリングできる", func(t *testing.T) {
		req := &g.ListEntitiesRequest{
			Category:      g.EntityCategory_PEOPLE,
			AllCategories: false,
		}

		resp, err := service.ListEntities(ctx, req)
		require.NoError(t, err)
		assert.Len(t, resp.Entities, 1)
		assert.Equal(t, "人物1", resp.Entities[0].Name)
	})

	t.Run("エイリアスも含めて取得できる", func(t *testing.T) {
		req := &g.ListEntitiesRequest{
			AllCategories: true,
		}

		resp, err := service.ListEntities(ctx, req)
		require.NoError(t, err)

		// entity1にはエイリアスがある
		for _, ent := range resp.Entities {
			if ent.Name == "人物1" {
				assert.Len(t, ent.Aliases, 1)
				assert.Equal(t, "人物1のエイリアス", ent.Aliases[0].Alias)
			}
		}
	})
}

func TestGetAllAliasesByUserID(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	service := &EntityEntry{DB: db}

	// テスト用ユーザーとエンティティを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	entity1 := &database.Entity{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       "エンティティ1",
		CategoryID: 1,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}
	require.NoError(t, entity1.Insert(context.Background(), db))

	entity2 := &database.Entity{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       "エンティティ2",
		CategoryID: 1,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}
	require.NoError(t, entity2.Insert(context.Background(), db))

	// エイリアスを作成
	alias1 := &database.EntityAlias{
		ID:        uuid.New(),
		EntityID:  entity1.ID,
		Alias:     "エイリアス1",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, alias1.Insert(context.Background(), db))

	alias2 := &database.EntityAlias{
		ID:        uuid.New(),
		EntityID:  entity1.ID,
		Alias:     "エイリアス2",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, alias2.Insert(context.Background(), db))

	alias3 := &database.EntityAlias{
		ID:        uuid.New(),
		EntityID:  entity2.ID,
		Alias:     "エイリアス3",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, alias3.Insert(context.Background(), db))

	t.Run("ユーザーの全エイリアスを一括取得できる", func(t *testing.T) {
		aliasMap, err := service.getAllAliasesByUserID(context.Background(), userID)
		require.NoError(t, err)

		// エンティティ1には2つのエイリアス
		assert.Len(t, aliasMap[entity1.ID.String()], 2)

		// エンティティ2には1つのエイリアス
		assert.Len(t, aliasMap[entity2.ID.String()], 1)
	})

	t.Run("エイリアスがない場合も正常に動作する", func(t *testing.T) {
		// エイリアスがないユーザー
		newUserID := uuid.New()
		newUser := &database.User{
			ID:        newUserID,
			Email:     fmt.Sprintf("newuser-%s@example.com", newUserID.String()),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		require.NoError(t, newUser.Insert(context.Background(), db))

		aliasMap, err := service.getAllAliasesByUserID(context.Background(), newUserID)
		require.NoError(t, err)
		assert.Empty(t, aliasMap)
	})
}
