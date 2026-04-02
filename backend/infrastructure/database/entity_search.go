package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// RelatedKeywordsByUserIDAndKeyword はキーワードに完全一致（大文字小文字無視）する
// エンティティ名またはエイリアスを持つエンティティを探し、
// そのエンティティに紐づく全てのキーワード（名前＋エイリアス）を返す。
// キーワード自身は結果に含まれない。
func RelatedKeywordsByUserIDAndKeyword(ctx context.Context, db DB, userID string, keyword string) ([]string, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, nil
	}

	// キーワードに一致するエンティティの名前と全エイリアスを取得
	const sqlstr = `
		SELECT e.name, ea.alias
		FROM entities e
		LEFT JOIN entity_aliases ea ON e.id = ea.entity_id
		WHERE e.user_id = $1
		AND (
			LOWER(e.name) = LOWER($2)
			OR EXISTS (
				SELECT 1 FROM entity_aliases ea2
				WHERE ea2.entity_id = e.id AND LOWER(ea2.alias) = LOWER($2)
			)
		)
	`
	rows, err := db.QueryContext(ctx, sqlstr, userID, keyword)
	if err != nil {
		return nil, logerror(err)
	}
	defer func() { _ = rows.Close() }()

	// エンティティ名・エイリアスを収集（元のキーワードは除外）
	seen := make(map[string]struct{})
	lowerKeyword := strings.ToLower(keyword)
	result := make([]string, 0)

	for rows.Next() {
		var name string
		var alias sql.NullString
		if err := rows.Scan(&name, &alias); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		// 元のキーワードと異なる場合のみ追加
		if strings.ToLower(name) != lowerKeyword {
			if _, exists := seen[strings.ToLower(name)]; !exists {
				seen[strings.ToLower(name)] = struct{}{}
				result = append(result, name)
			}
		}
		if alias.Valid && strings.ToLower(alias.String) != lowerKeyword {
			if _, exists := seen[strings.ToLower(alias.String)]; !exists {
				seen[strings.ToLower(alias.String)] = struct{}{}
				result = append(result, alias.String)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return result, nil
}
