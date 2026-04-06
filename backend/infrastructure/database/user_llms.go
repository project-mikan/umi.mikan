package database

import (
	"context"
	"fmt"
)

// UserIDsWithAutoSummaryDaily はauto_summary_dailyがtrueのユーザーIDの一覧を返す
func UserIDsWithAutoSummaryDaily(ctx context.Context, db DB) ([]string, error) {
	const sqlstr = `SELECT user_id FROM user_llms WHERE auto_summary_daily = true`
	return queryUserIDsByFlag(ctx, db, sqlstr)
}

// UserIDsWithAutoSummaryMonthly はauto_summary_monthlyがtrueのユーザーIDの一覧を返す
func UserIDsWithAutoSummaryMonthly(ctx context.Context, db DB) ([]string, error) {
	const sqlstr = `SELECT user_id FROM user_llms WHERE auto_summary_monthly = true`
	return queryUserIDsByFlag(ctx, db, sqlstr)
}

// UserIDsWithAutoLatestTrendEnabled はauto_latest_trend_enabledがtrueのユーザーIDの一覧を返す
func UserIDsWithAutoLatestTrendEnabled(ctx context.Context, db DB) ([]string, error) {
	const sqlstr = `SELECT user_id FROM user_llms WHERE auto_latest_trend_enabled = true`
	return queryUserIDsByFlag(ctx, db, sqlstr)
}

// UserIDsWithSemanticSearchEnabled はsemantic_search_enabledがtrueのユーザーIDの一覧を返す
func UserIDsWithSemanticSearchEnabled(ctx context.Context, db DB) ([]string, error) {
	const sqlstr = `SELECT user_id FROM user_llms WHERE semantic_search_enabled = true`
	return queryUserIDsByFlag(ctx, db, sqlstr)
}

// queryUserIDsByFlag はSQLクエリを実行してユーザーIDのスライスを返す内部ヘルパー
func queryUserIDsByFlag(ctx context.Context, db DB, sqlstr string) ([]string, error) {
	rows, err := db.QueryContext(ctx, sqlstr)
	if err != nil {
		return nil, fmt.Errorf("failed to query user IDs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return userIDs, nil
}
