package database

import (
	"context"
	"fmt"
	"time"
)

// DiariesByUserIDAndDateRange は指定ユーザーの指定期間（開始年月〜終了年月）の全日記を返す。
// 開始月の1日から終了月の末日までを対象とし、date昇順で返す。
// 大量データ対応のため1回のSQLで取得する。
func DiariesByUserIDAndDateRange(ctx context.Context, db DB, userID string, fromYear, fromMonth, toYear, toMonth int) ([]*Diary, error) {
	// 開始日（月初）と終了日（月末）を計算する
	fromDate := time.Date(fromYear, time.Month(fromMonth), 1, 0, 0, 0, 0, time.UTC)
	// 翌月の1日から1日引いて月末を取得する
	toDate := time.Date(toYear, time.Month(toMonth)+1, 0, 0, 0, 0, 0, time.UTC)

	return diariesByUserIDAndDateRange(ctx, db, userID, fromDate, toDate)
}

// diariesByUserIDAndDateRange は指定ユーザーの指定日付範囲（両端含む）の全日記をdate昇順で返す。
func diariesByUserIDAndDateRange(ctx context.Context, db DB, userID string, fromDate, toDate time.Time) ([]*Diary, error) {
	const sqlstr = `
		SELECT id, user_id, content, date, created_at, updated_at
		FROM diaries
		WHERE user_id = $1
		  AND date >= $2
		  AND date <= $3
		ORDER BY date ASC
	`

	rows, err := db.QueryContext(ctx, sqlstr, userID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query diaries by date range: %w", err)
	}
	defer func() { _ = rows.Close() }()

	diaries := make([]*Diary, 0)
	for rows.Next() {
		var d Diary
		if err := rows.Scan(&d.ID, &d.UserID, &d.Content, &d.Date, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan diary row: %w", err)
		}
		diaries = append(diaries, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return diaries, nil
}
