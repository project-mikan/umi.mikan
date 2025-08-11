package database

import (
	"context"
	"fmt"
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

func CountDiariesByUserID(ctx context.Context, db DB, userID string) (int, error) {
	const sqlstr = `SELECT COUNT(*) FROM diaries WHERE user_id = $1`
	var count int
	err := db.QueryRowContext(ctx, sqlstr, userID).Scan(&count)
	if err != nil {
		return 0, logerror(err)
	}
	return count, nil
}
