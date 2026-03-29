package entity

import (
	"context"
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
)

// TestValidationIntegration バリデーションの統合テスト
func TestValidationIntegration(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	entityService := &EntityEntry{DB: db}

	// テスト用ユーザーを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	t.Run("空のエンティティ名は拒否される", func(t *testing.T) {
		req := &g.CreateEntityRequest{
			Name:     "",
			Category: g.EntityCategory_PEOPLE,
		}
		_, err := entityService.CreateEntity(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entity name cannot be empty")
	})

	t.Run("255文字を超えるエンティティ名は拒否される", func(t *testing.T) {
		longName := string(make([]byte, 256))
		for i := range longName {
			longName = longName[:i] + "あ"
		}
		req := &g.CreateEntityRequest{
			Name:     longName[:256],
			Category: g.EntityCategory_PEOPLE,
		}
		_, err := entityService.CreateEntity(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be less than 255 characters")
	})

	t.Run("制御文字を含むエンティティ名は拒否される", func(t *testing.T) {
		req := &g.CreateEntityRequest{
			Name:     "テスト\x00人物",
			Category: g.EntityCategory_PEOPLE,
		}
		_, err := entityService.CreateEntity(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid control characters")
	})

	t.Run("前後に空白を含むエンティティ名は拒否される", func(t *testing.T) {
		req := &g.CreateEntityRequest{
			Name:     " テスト人物 ",
			Category: g.EntityCategory_PEOPLE,
		}
		_, err := entityService.CreateEntity(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "leading or trailing whitespace")
	})

	t.Run("空のエイリアスは拒否される", func(t *testing.T) {
		// まずエンティティを作成
		entityResp, err := entityService.CreateEntity(ctx, &g.CreateEntityRequest{
			Name:     "テストエンティティ",
			Category: g.EntityCategory_PEOPLE,
		})
		require.NoError(t, err)

		req := &g.CreateEntityAliasRequest{
			EntityId: entityResp.Entity.Id,
			Alias:    "",
		}
		_, err = entityService.CreateEntityAlias(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "alias cannot be empty")
	})

	t.Run("255文字を超えるエイリアスは拒否される", func(t *testing.T) {
		// まずエンティティを作成
		entityResp, err := entityService.CreateEntity(ctx, &g.CreateEntityRequest{
			Name:     "テストエンティティ2",
			Category: g.EntityCategory_PEOPLE,
		})
		require.NoError(t, err)

		longAlias := string(make([]byte, 256))
		for i := range longAlias {
			longAlias = longAlias[:i] + "あ"
		}
		req := &g.CreateEntityAliasRequest{
			EntityId: entityResp.Entity.Id,
			Alias:    longAlias[:256],
		}
		_, err = entityService.CreateEntityAlias(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be less than 255 characters")
	})
}
