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
	diaryService "github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/project-mikan/umi.mikan/backend/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiaryEntityIntegration 日記とエンティティの統合テスト
func TestDiaryEntityIntegration(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	entityService := &EntityEntry{DB: db}
	dService := &diaryService.DiaryEntry{DB: db}

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

	t.Run("日記作成時にエンティティが正しく紐付けられる", func(t *testing.T) {
		// 1. エンティティを作成
		createEntityReq := &g.CreateEntityRequest{
			Name:     "テスト人物",
			Category: g.EntityCategory_PEOPLE,
			Memo:     "テストメモ",
		}
		entityResp, err := entityService.CreateEntity(ctx, createEntityReq)
		require.NoError(t, err)
		require.NotNil(t, entityResp)
		entityID := entityResp.Entity.Id

		// 2. エンティティを含む日記を作成
		now := time.Now()
		createDiaryReq := &g.CreateDiaryEntryRequest{
			Content: "今日はテスト人物と会いました",
			Date: &g.YMD{
				Year:  uint32(now.Year()),
				Month: uint32(now.Month()),
				Day:   uint32(now.Day()),
			},
			DiaryEntities: []*g.DiaryEntityInput{
				{
					EntityId: entityID,
					Positions: []*g.Position{
						{Start: 3, End: 8}, // "テスト人物"の位置
					},
				},
			},
		}
		diaryResp, err := dService.CreateDiaryEntry(ctx, createDiaryReq)
		require.NoError(t, err)
		require.NotNil(t, diaryResp)

		// 3. diary_entitiesに正しく登録されているか確認
		diaryEntities, err := database.DiaryEntitiesByDiaryID(ctx, db, uuid.MustParse(diaryResp.Entry.Id))
		require.NoError(t, err)
		require.Len(t, diaryEntities, 1)
		assert.Equal(t, entityID, diaryEntities[0].EntityID.String())

		// 4. GetDiariesByEntityで取得できるか確認
		getDiariesReq := &g.GetDiariesByEntityRequest{
			EntityId: entityID,
		}
		diariesResp, err := entityService.GetDiariesByEntity(ctx, getDiariesReq)
		require.NoError(t, err)
		require.Len(t, diariesResp.Diaries, 1)
		assert.Equal(t, diaryResp.Entry.Id, diariesResp.Diaries[0].Id)
		assert.Equal(t, "今日はテスト人物と会いました", diariesResp.Diaries[0].Content)
		assert.Len(t, diariesResp.Diaries[0].Positions, 1)
		assert.Equal(t, uint32(3), diariesResp.Diaries[0].Positions[0].Start)
		assert.Equal(t, uint32(8), diariesResp.Diaries[0].Positions[0].End)
	})

	t.Run("日記更新時にエンティティの紐付けが更新される", func(t *testing.T) {
		// 1. エンティティを作成
		createEntityReq := &g.CreateEntityRequest{
			Name:     "更新テスト人物",
			Category: g.EntityCategory_PEOPLE,
		}
		entityResp, err := entityService.CreateEntity(ctx, createEntityReq)
		require.NoError(t, err)
		entityID := entityResp.Entity.Id

		// 2. エンティティを含む日記を作成
		now := time.Now().Add(24 * time.Hour)
		createDiaryReq := &g.CreateDiaryEntryRequest{
			Content: "更新テスト人物と会った",
			Date: &g.YMD{
				Year:  uint32(now.Year()),
				Month: uint32(now.Month()),
				Day:   uint32(now.Day()),
			},
			DiaryEntities: []*g.DiaryEntityInput{
				{
					EntityId: entityID,
					Positions: []*g.Position{
						{Start: 0, End: 8},
					},
				},
			},
		}
		diaryResp, err := dService.CreateDiaryEntry(ctx, createDiaryReq)
		require.NoError(t, err)

		// 3. 日記を更新してエンティティを削除
		updateDiaryReq := &g.UpdateDiaryEntryRequest{
			Id:      diaryResp.Entry.Id,
			Content: "今日は誰とも会わなかった",
			Date: &g.YMD{
				Year:  uint32(now.Year()),
				Month: uint32(now.Month()),
				Day:   uint32(now.Day()),
			},
			DiaryEntities: []*g.DiaryEntityInput{}, // エンティティなし
		}
		_, err = dService.UpdateDiaryEntry(ctx, updateDiaryReq)
		require.NoError(t, err)

		// 4. diary_entitiesから削除されているか確認
		diaryEntities, err := database.DiaryEntitiesByDiaryID(ctx, db, uuid.MustParse(diaryResp.Entry.Id))
		require.NoError(t, err)
		assert.Len(t, diaryEntities, 0)

		// 5. GetDiariesByEntityで取得されないことを確認
		getDiariesReq := &g.GetDiariesByEntityRequest{
			EntityId: entityID,
		}
		diariesResp, err := entityService.GetDiariesByEntity(ctx, getDiariesReq)
		require.NoError(t, err)
		assert.Len(t, diariesResp.Diaries, 0)
	})

	t.Run("エンティティ削除時にdiary_entitiesもカスケード削除される", func(t *testing.T) {
		// 1. エンティティを作成
		createEntityReq := &g.CreateEntityRequest{
			Name:     "削除テスト人物",
			Category: g.EntityCategory_PEOPLE,
		}
		entityResp, err := entityService.CreateEntity(ctx, createEntityReq)
		require.NoError(t, err)
		entityID := entityResp.Entity.Id

		// 2. エンティティを含む日記を作成
		now := time.Now().Add(48 * time.Hour)
		createDiaryReq := &g.CreateDiaryEntryRequest{
			Content: "削除テスト人物と遊んだ",
			Date: &g.YMD{
				Year:  uint32(now.Year()),
				Month: uint32(now.Month()),
				Day:   uint32(now.Day()),
			},
			DiaryEntities: []*g.DiaryEntityInput{
				{
					EntityId: entityID,
					Positions: []*g.Position{
						{Start: 0, End: 8},
					},
				},
			},
		}
		diaryResp, err := dService.CreateDiaryEntry(ctx, createDiaryReq)
		require.NoError(t, err)
		diaryID := uuid.MustParse(diaryResp.Entry.Id)

		// 3. エンティティを削除
		deleteEntityReq := &g.DeleteEntityRequest{
			Id: entityID,
		}
		_, err = entityService.DeleteEntity(ctx, deleteEntityReq)
		require.NoError(t, err)

		// 4. diary_entitiesから削除されているか確認
		diaryEntities, err := database.DiaryEntitiesByDiaryID(ctx, db, diaryID)
		require.NoError(t, err)
		assert.Len(t, diaryEntities, 0)

		// 5. 日記は残っているか確認
		diary, err := database.DiaryByID(ctx, db, diaryID)
		require.NoError(t, err)
		assert.Equal(t, "削除テスト人物と遊んだ", diary.Content)
	})

	t.Run("複数エンティティを含む日記の作成と取得", func(t *testing.T) {
		// 1. 複数のエンティティを作成
		entity1Resp, err := entityService.CreateEntity(ctx, &g.CreateEntityRequest{
			Name:     "人物A",
			Category: g.EntityCategory_PEOPLE,
		})
		require.NoError(t, err)
		entity1ID := entity1Resp.Entity.Id

		entity2Resp, err := entityService.CreateEntity(ctx, &g.CreateEntityRequest{
			Name:     "人物B",
			Category: g.EntityCategory_PEOPLE,
		})
		require.NoError(t, err)
		entity2ID := entity2Resp.Entity.Id

		// 2. 複数エンティティを含む日記を作成
		now := time.Now().Add(72 * time.Hour)
		createDiaryReq := &g.CreateDiaryEntryRequest{
			Content: "人物Aと人物Bと会った",
			Date: &g.YMD{
				Year:  uint32(now.Year()),
				Month: uint32(now.Month()),
				Day:   uint32(now.Day()),
			},
			DiaryEntities: []*g.DiaryEntityInput{
				{
					EntityId: entity1ID,
					Positions: []*g.Position{
						{Start: 0, End: 3},
					},
				},
				{
					EntityId: entity2ID,
					Positions: []*g.Position{
						{Start: 4, End: 7},
					},
				},
			},
		}
		diaryResp, err := dService.CreateDiaryEntry(ctx, createDiaryReq)
		require.NoError(t, err)

		// 3. 各エンティティからdiary取得を確認
		diaries1Resp, err := entityService.GetDiariesByEntity(ctx, &g.GetDiariesByEntityRequest{
			EntityId: entity1ID,
		})
		require.NoError(t, err)
		require.Len(t, diaries1Resp.Diaries, 1)
		assert.Equal(t, diaryResp.Entry.Id, diaries1Resp.Diaries[0].Id)

		diaries2Resp, err := entityService.GetDiariesByEntity(ctx, &g.GetDiariesByEntityRequest{
			EntityId: entity2ID,
		})
		require.NoError(t, err)
		require.Len(t, diaries2Resp.Diaries, 1)
		assert.Equal(t, diaryResp.Entry.Id, diaries2Resp.Diaries[0].Id)
	})
}

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
