package entity

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EntityEntry struct {
	g.UnimplementedEntityServiceServer
	DB database.DB
}

// CreateEntity エンティティを作成する
func (s *EntityEntry) CreateEntity(
	ctx context.Context,
	message *g.CreateEntityRequest,
) (*g.CreateEntityResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// エンティティ名が既存のエイリアスと重複していないかチェック
	checkAliasQuery := `
		SELECT COUNT(*) FROM entity_aliases ea
		INNER JOIN entities e ON ea.entity_id = e.id
		WHERE e.user_id = $1 AND ea.alias = $2
	`
	var aliasCount int
	if err := s.DB.(*sql.DB).QueryRowContext(ctx, checkAliasQuery, userID, message.Name).Scan(&aliasCount); err != nil {
		return nil, err
	}
	if aliasCount > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "name '%s' is already used as an alias", message.Name)
	}

	id := uuid.New()
	currentTime := time.Now().Unix()

	entity := &database.Entity{
		ID:         id,
		UserID:     userID,
		Name:       message.Name,
		CategoryID: int(message.Category),
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	}

	// Memoフィールドの設定
	if message.Memo != "" {
		entity.Memo = sql.NullString{
			String: message.Memo,
			Valid:  true,
		}
	}

	if err := entity.Insert(ctx, s.DB); err != nil {
		// PostgreSQLのユニーク制約違反エラーをチェック
		if pqErr, ok := err.(*pq.Error); ok {
			// エラーコード 23505 はユニーク制約違反
			if pqErr.Code == "23505" && strings.Contains(pqErr.Message, "entities_user_id_name_key") {
				return nil, status.Errorf(codes.AlreadyExists, "entity with name '%s' already exists", message.Name)
			}
		}
		return nil, err
	}

	// エイリアスは空配列で返す
	return &g.CreateEntityResponse{
		Entity: &g.Entity{
			Id:        entity.ID.String(),
			Name:      entity.Name,
			Category:  g.EntityCategory(entity.CategoryID),
			Memo:      entity.Memo.String,
			Aliases:   []*g.EntityAlias{},
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		},
	}, nil
}

// UpdateEntity エンティティを更新する
func (s *EntityEntry) UpdateEntity(
	ctx context.Context,
	message *g.UpdateEntityRequest,
) (*g.UpdateEntityResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	entityID, err := uuid.Parse(message.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity ID")
	}

	entity, err := database.EntityByID(ctx, s.DB, entityID)
	if err != nil {
		return nil, err
	}

	// エンティティの所有者確認
	if entity.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to update this entity")
	}

	// トランザクション内でエンティティを更新
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		entity.Name = message.Name
		entity.CategoryID = int(message.Category)
		entity.UpdatedAt = time.Now().Unix()

		// Memoフィールドの更新
		if message.Memo != "" {
			entity.Memo = sql.NullString{
				String: message.Memo,
				Valid:  true,
			}
		} else {
			entity.Memo = sql.NullString{Valid: false}
		}

		if err := entity.Update(ctx, tx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// エイリアスを取得
	aliases, err := database.EntityAliasesByEntityID(ctx, s.DB, entityID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	protoAliases := make([]*g.EntityAlias, 0, len(aliases))
	for _, alias := range aliases {
		protoAliases = append(protoAliases, &g.EntityAlias{
			Id:        alias.ID.String(),
			EntityId:  alias.EntityID.String(),
			Alias:     alias.Alias,
			CreatedAt: alias.CreatedAt,
			UpdatedAt: alias.UpdatedAt,
		})
	}

	return &g.UpdateEntityResponse{
		Entity: &g.Entity{
			Id:        entity.ID.String(),
			Name:      entity.Name,
			Category:  g.EntityCategory(entity.CategoryID),
			Memo:      entity.Memo.String,
			Aliases:   protoAliases,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		},
	}, nil
}

// DeleteEntity エンティティを削除する
func (s *EntityEntry) DeleteEntity(
	ctx context.Context,
	message *g.DeleteEntityRequest,
) (*g.DeleteEntityResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	entityID, err := uuid.Parse(message.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity ID")
	}

	entity, err := database.EntityByID(ctx, s.DB, entityID)
	if err != nil {
		return nil, err
	}

	// エンティティの所有者確認
	if entity.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to delete this entity")
	}

	// トランザクション内でエンティティを削除
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		return entity.Delete(ctx, tx)
	})
	if err != nil {
		return nil, err
	}

	return &g.DeleteEntityResponse{
		Success: true,
	}, nil
}

// GetEntity エンティティを取得する
func (s *EntityEntry) GetEntity(
	ctx context.Context,
	message *g.GetEntityRequest,
) (*g.GetEntityResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	entityID, err := uuid.Parse(message.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity ID")
	}

	entity, err := database.EntityByID(ctx, s.DB, entityID)
	if err != nil {
		return nil, err
	}

	// エンティティの所有者確認
	if entity.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to view this entity")
	}

	// エイリアスを取得
	aliases, err := database.EntityAliasesByEntityID(ctx, s.DB, entityID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	protoAliases := make([]*g.EntityAlias, 0, len(aliases))
	for _, alias := range aliases {
		protoAliases = append(protoAliases, &g.EntityAlias{
			Id:        alias.ID.String(),
			EntityId:  alias.EntityID.String(),
			Alias:     alias.Alias,
			CreatedAt: alias.CreatedAt,
			UpdatedAt: alias.UpdatedAt,
		})
	}

	return &g.GetEntityResponse{
		Entity: &g.Entity{
			Id:        entity.ID.String(),
			Name:      entity.Name,
			Category:  g.EntityCategory(entity.CategoryID),
			Memo:      entity.Memo.String,
			Aliases:   protoAliases,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		},
	}, nil
}

// ListEntities エンティティ一覧を取得する
func (s *EntityEntry) ListEntities(
	ctx context.Context,
	message *g.ListEntitiesRequest,
) (*g.ListEntitiesResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// ユーザーのエンティティを取得
	entities, err := database.EntitiesByUserID(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	protoEntities := make([]*g.Entity, 0)
	for _, entity := range entities {
		// カテゴリフィルタ
		// all_categories が true の場合は全て表示、false の場合は category でフィルタ
		if !message.AllCategories && g.EntityCategory(entity.CategoryID) != message.Category {
			continue
		}

		// エイリアスを取得
		aliases, err := database.EntityAliasesByEntityID(ctx, s.DB, entity.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		protoAliases := make([]*g.EntityAlias, 0, len(aliases))
		for _, alias := range aliases {
			protoAliases = append(protoAliases, &g.EntityAlias{
				Id:        alias.ID.String(),
				EntityId:  alias.EntityID.String(),
				Alias:     alias.Alias,
				CreatedAt: alias.CreatedAt,
				UpdatedAt: alias.UpdatedAt,
			})
		}

		protoEntities = append(protoEntities, &g.Entity{
			Id:        entity.ID.String(),
			Name:      entity.Name,
			Category:  g.EntityCategory(entity.CategoryID),
			Memo:      entity.Memo.String,
			Aliases:   protoAliases,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		})
	}

	return &g.ListEntitiesResponse{
		Entities: protoEntities,
	}, nil
}

// CreateEntityAlias エイリアスを追加する
func (s *EntityEntry) CreateEntityAlias(
	ctx context.Context,
	message *g.CreateEntityAliasRequest,
) (*g.CreateEntityAliasResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	entityID, err := uuid.Parse(message.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity ID")
	}

	// エンティティの所有者確認
	entity, err := database.EntityByID(ctx, s.DB, entityID)
	if err != nil {
		return nil, err
	}
	if entity.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to add alias to this entity")
	}

	// エイリアスが既存のエンティティ名と重複していないかチェック
	checkEntityQuery := `
		SELECT COUNT(*) FROM entities
		WHERE user_id = $1 AND name = $2
	`
	var entityCount int
	if err := s.DB.(*sql.DB).QueryRowContext(ctx, checkEntityQuery, userID, message.Alias).Scan(&entityCount); err != nil {
		return nil, err
	}
	if entityCount > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "alias '%s' is already used as an entity name", message.Alias)
	}

	// エイリアスが既存の他のエイリアスと重複していないかチェック（同じユーザー内）
	checkOtherAliasQuery := `
		SELECT COUNT(*) FROM entity_aliases ea
		INNER JOIN entities e ON ea.entity_id = e.id
		WHERE e.user_id = $1 AND ea.alias = $2
	`
	var otherAliasCount int
	if err := s.DB.(*sql.DB).QueryRowContext(ctx, checkOtherAliasQuery, userID, message.Alias).Scan(&otherAliasCount); err != nil {
		return nil, err
	}
	if otherAliasCount > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "alias '%s' is already used", message.Alias)
	}

	id := uuid.New()
	currentTime := time.Now().Unix()

	alias := &database.EntityAlias{
		ID:        id,
		EntityID:  entityID,
		Alias:     message.Alias,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	if err := alias.Insert(ctx, s.DB); err != nil {
		// PostgreSQLのユニーク制約違反エラーをチェック
		if pqErr, ok := err.(*pq.Error); ok {
			// エラーコード 23505 はユニーク制約違反
			if pqErr.Code == "23505" {
				return nil, status.Errorf(codes.AlreadyExists, "alias '%s' already exists for this entity", message.Alias)
			}
		}
		return nil, err
	}

	return &g.CreateEntityAliasResponse{
		Alias: &g.EntityAlias{
			Id:        alias.ID.String(),
			EntityId:  alias.EntityID.String(),
			Alias:     alias.Alias,
			CreatedAt: alias.CreatedAt,
			UpdatedAt: alias.UpdatedAt,
		},
	}, nil
}

// DeleteEntityAlias エイリアスを削除する
func (s *EntityEntry) DeleteEntityAlias(
	ctx context.Context,
	message *g.DeleteEntityAliasRequest,
) (*g.DeleteEntityAliasResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	aliasID, err := uuid.Parse(message.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid alias ID")
	}

	alias, err := database.EntityAliasByID(ctx, s.DB, aliasID)
	if err != nil {
		return nil, err
	}

	// エンティティの所有者確認
	entity, err := database.EntityByID(ctx, s.DB, alias.EntityID)
	if err != nil {
		return nil, err
	}
	if entity.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to delete this alias")
	}

	// トランザクション内でエイリアスを削除
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		return alias.Delete(ctx, tx)
	})
	if err != nil {
		return nil, err
	}

	return &g.DeleteEntityAliasResponse{
		Success: true,
	}, nil
}

// SearchEntities エンティティを検索する（ユーザーの入力に対する候補表示）
func (s *EntityEntry) SearchEntities(
	ctx context.Context,
	message *g.SearchEntitiesRequest,
) (*g.SearchEntitiesResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// エンティティ名とエイリアスから検索
	// LIKE検索を使用して部分一致検索を行う
	query := `
		SELECT DISTINCT e.id, e.user_id, e.created_at, e.updated_at, e.category_id, e.name, e.memo
		FROM entities e
		LEFT JOIN entity_aliases ea ON e.id = ea.entity_id
		WHERE e.user_id = $1
		AND (e.name ILIKE $2 OR ea.alias ILIKE $2)
		ORDER BY e.name
	`

	rows, err := s.DB.(*sql.DB).QueryContext(ctx, query, userID, "%"+message.Query+"%")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	protoEntities := make([]*g.Entity, 0)
	processedIDs := make(map[string]bool) // 重複を避けるためのマップ

	for rows.Next() {
		var entity database.Entity
		if err := rows.Scan(&entity.ID, &entity.UserID, &entity.CreatedAt, &entity.UpdatedAt, &entity.CategoryID, &entity.Name, &entity.Memo); err != nil {
			return nil, err
		}

		// 重複チェック
		if processedIDs[entity.ID.String()] {
			continue
		}
		processedIDs[entity.ID.String()] = true

		// エイリアスを取得
		aliases, err := database.EntityAliasesByEntityID(ctx, s.DB, entity.ID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		protoAliases := make([]*g.EntityAlias, 0, len(aliases))
		for _, alias := range aliases {
			protoAliases = append(protoAliases, &g.EntityAlias{
				Id:        alias.ID.String(),
				EntityId:  alias.EntityID.String(),
				Alias:     alias.Alias,
				CreatedAt: alias.CreatedAt,
				UpdatedAt: alias.UpdatedAt,
			})
		}

		protoEntities = append(protoEntities, &g.Entity{
			Id:        entity.ID.String(),
			Name:      entity.Name,
			Category:  g.EntityCategory(entity.CategoryID),
			Memo:      entity.Memo.String,
			Aliases:   protoAliases,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &g.SearchEntitiesResponse{
		Entities: protoEntities,
	}, nil
}

// GetDiariesByEntity エンティティに紐づく日記を取得する
func (s *EntityEntry) GetDiariesByEntity(
	ctx context.Context,
	message *g.GetDiariesByEntityRequest,
) (*g.GetDiariesByEntityResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	entityID, err := uuid.Parse(message.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity ID")
	}

	// エンティティの所有者確認
	entity, err := database.EntityByID(ctx, s.DB, entityID)
	if err != nil {
		return nil, err
	}
	if entity.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to view this entity")
	}

	// エンティティに紐づく日記を取得
	diaryEntities, err := database.DiaryEntitiesByEntityID(ctx, s.DB, entityID)
	if err != nil {
		return nil, err
	}

	protoDiaries := make([]*g.DiaryWithEntity, 0, len(diaryEntities))
	for _, de := range diaryEntities {
		// 日記を取得
		diary, err := database.DiaryByID(ctx, s.DB, de.DiaryID)
		if err != nil {
			continue // 日記が削除されている場合はスキップ
		}

		// positionsをJSONBから[]Positionに変換
		var positions []struct {
			Start uint32 `json:"start"`
			End   uint32 `json:"end"`
		}
		if err := json.Unmarshal(de.Positions, &positions); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse positions")
		}

		protoPositions := make([]*g.Position, 0, len(positions))
		for _, pos := range positions {
			protoPositions = append(protoPositions, &g.Position{
				Start: pos.Start,
				End:   pos.End,
			})
		}

		protoDiaries = append(protoDiaries, &g.DiaryWithEntity{
			Id:        diary.ID.String(),
			Content:   diary.Content,
			Date:      diary.Date.Format("2006-01-02"),
			Positions: protoPositions,
			CreatedAt: diary.CreatedAt,
			UpdatedAt: diary.UpdatedAt,
		})
	}

	return &g.GetDiariesByEntityResponse{
		Diaries: protoDiaries,
	}, nil
}
