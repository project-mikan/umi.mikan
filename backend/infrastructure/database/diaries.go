package database

import (
	"context"
	"fmt"
	"strings"
)

func DiariesByUserIDAndContent(ctx context.Context, db DB, userID string, content string) ([]*Diary, error) {
	// query
	const sqlstr = `SELECT ` +
		`id, user_id, content, date, created_at, updated_at ` +
		`FROM diaries ` +
		`WHERE user_id = $1 AND content LIKE $2 ORDER BY date DESC`
	rows, err := db.QueryContext(ctx, sqlstr, userID, "%"+content+"%")
	if err != nil {
		return nil, logerror(err)
	}
	defer func() { _ = rows.Close() }()

	// 結果をマップに格納
	diaries := make([]*Diary, 0)
	for rows.Next() {
		var diary Diary
		if err := rows.Scan(&diary.ID, &diary.UserID, &diary.Content, &diary.Date, &diary.CreatedAt, &diary.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		diaries = append(diaries, &diary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return diaries, nil
}

// DiariesByUserIDAndKeywords は複数キーワードのいずれかを本文に含む日記をORで検索する。
func DiariesByUserIDAndKeywords(ctx context.Context, db DB, userID string, keywords []string) ([]*Diary, error) {
	if len(keywords) == 0 {
		return DiariesByUserIDAndContent(ctx, db, userID, "")
	}

	// 複数キーワードのOR条件を動的に構築
	conditions := make([]string, 0, len(keywords))
	args := make([]any, 0, 1+len(keywords))
	args = append(args, userID)
	for i, kw := range keywords {
		conditions = append(conditions, fmt.Sprintf("content LIKE $%d", i+2))
		args = append(args, "%"+kw+"%")
	}

	sqlstr := `SELECT id, user_id, content, date, created_at, updated_at FROM diaries WHERE user_id = $1 AND (` +
		strings.Join(conditions, " OR ") +
		`) ORDER BY date DESC`

	rows, err := db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return nil, logerror(err)
	}
	defer func() { _ = rows.Close() }()

	diaries := make([]*Diary, 0)
	for rows.Next() {
		var diary Diary
		if err := rows.Scan(&diary.ID, &diary.UserID, &diary.Content, &diary.Date, &diary.CreatedAt, &diary.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		diaries = append(diaries, &diary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return diaries, nil
}
