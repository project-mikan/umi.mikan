package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// AliasesByUserID はユーザーの全エイリアスを取得してentityID文字列をキーとするマップで返す（N+1クエリ回避）
func AliasesByUserID(ctx context.Context, db DB, userID uuid.UUID) (map[string][]*EntityAlias, error) {
	const sqlstr = `
		SELECT ea.id, ea.entity_id, ea.created_at, ea.updated_at, ea.alias
		FROM entity_aliases ea
		INNER JOIN entities e ON ea.entity_id = e.id
		WHERE e.user_id = $1
		ORDER BY ea.entity_id, ea.created_at
	`
	rows, err := db.QueryContext(ctx, sqlstr, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query aliases by user ID: %w", err)
	}
	defer func() { _ = rows.Close() }()

	aliasMap := make(map[string][]*EntityAlias)
	for rows.Next() {
		var alias EntityAlias
		if err := rows.Scan(&alias.ID, &alias.EntityID, &alias.CreatedAt, &alias.UpdatedAt, &alias.Alias); err != nil {
			return nil, fmt.Errorf("failed to scan alias: %w", err)
		}
		entityIDStr := alias.EntityID.String()
		aliasMap[entityIDStr] = append(aliasMap[entityIDStr], &alias)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return aliasMap, nil
}

// CountAliasMatchingName はユーザーのエイリアス中に指定した名前が存在するか件数を返す
// エンティティ名がエイリアスと重複していないか確認するために使用する
func CountAliasMatchingName(ctx context.Context, db DB, userID uuid.UUID, name string) (int, error) {
	const sqlstr = `
		SELECT COUNT(*) FROM entity_aliases ea
		INNER JOIN entities e ON ea.entity_id = e.id
		WHERE e.user_id = $1 AND ea.alias = $2
	`
	return queryCount(ctx, db, sqlstr, userID, name)
}

// CountEntityMatchingAlias はユーザーのエンティティ名に指定したエイリアスが存在するか件数を返す
// エイリアスがエンティティ名と重複していないか確認するために使用する
func CountEntityMatchingAlias(ctx context.Context, db DB, userID uuid.UUID, alias string) (int, error) {
	const sqlstr = `
		SELECT COUNT(*) FROM entities
		WHERE user_id = $1 AND name = $2
	`
	return queryCount(ctx, db, sqlstr, userID, alias)
}

// CountAliasDuplicate はユーザーの全エイリアス中に指定したエイリアスが存在するか件数を返す
func CountAliasDuplicate(ctx context.Context, db DB, userID uuid.UUID, alias string) (int, error) {
	const sqlstr = `
		SELECT COUNT(*) FROM entity_aliases ea
		INNER JOIN entities e ON ea.entity_id = e.id
		WHERE e.user_id = $1 AND ea.alias = $2
	`
	return queryCount(ctx, db, sqlstr, userID, alias)
}

// CountAliasDuplicateExcluding はユーザーの全エイリアス中に指定したエイリアスが存在するか件数を返す（自分自身を除く）
// エイリアス更新時に自分以外との重複チェックに使用する
func CountAliasDuplicateExcluding(ctx context.Context, db DB, userID uuid.UUID, alias string, excludeAliasID uuid.UUID) (int, error) {
	const sqlstr = `
		SELECT COUNT(*) FROM entity_aliases ea
		INNER JOIN entities e ON ea.entity_id = e.id
		WHERE e.user_id = $1 AND ea.alias = $2 AND ea.id != $3
	`
	return queryCount(ctx, db, sqlstr, userID, alias, excludeAliasID)
}

// SearchEntitiesByQuery はユーザーのエンティティをクエリ文字列で検索する
// クエリが空の場合は全件返す
func SearchEntitiesByQuery(ctx context.Context, db DB, userID uuid.UUID, query string) ([]*Entity, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if query == "" {
		const sqlstr = `
			SELECT DISTINCT e.id, e.user_id, e.created_at, e.updated_at, e.category_id, e.name, e.memo
			FROM entities e
			WHERE e.user_id = $1
			ORDER BY e.name
		`
		rows, err = db.QueryContext(ctx, sqlstr, userID)
	} else {
		const sqlstr = `
			SELECT DISTINCT e.id, e.user_id, e.created_at, e.updated_at, e.category_id, e.name, e.memo
			FROM entities e
			LEFT JOIN entity_aliases ea ON e.id = ea.entity_id
			WHERE e.user_id = $1
			AND (e.name ILIKE $2 OR ea.alias ILIKE $2)
			ORDER BY e.name
		`
		rows, err = db.QueryContext(ctx, sqlstr, userID, "%"+query+"%")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entities []*Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.UserID, &e.CreatedAt, &e.UpdatedAt, &e.CategoryID, &e.Name, &e.Memo); err != nil {
			return nil, fmt.Errorf("failed to scan entity: %w", err)
		}
		entities = append(entities, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return entities, nil
}
