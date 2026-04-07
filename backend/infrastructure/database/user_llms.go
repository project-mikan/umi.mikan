package database

import (
	"context"
)

// UserIDsWithAutoSummaryDaily はauto_summary_dailyがtrueのユーザーIDの一覧を返す
func UserIDsWithAutoSummaryDaily(ctx context.Context, db DB) ([]string, error) {
	const sqlstr = `SELECT user_id FROM user_llms WHERE auto_summary_daily = true`
	return queryStringSlice(ctx, db, sqlstr)
}

// UserIDsWithAutoSummaryMonthly はauto_summary_monthlyがtrueのユーザーIDの一覧を返す
func UserIDsWithAutoSummaryMonthly(ctx context.Context, db DB) ([]string, error) {
	const sqlstr = `SELECT user_id FROM user_llms WHERE auto_summary_monthly = true`
	return queryStringSlice(ctx, db, sqlstr)
}

// UserIDsWithAutoLatestTrendEnabled はauto_latest_trend_enabledがtrueのユーザーIDの一覧を返す
func UserIDsWithAutoLatestTrendEnabled(ctx context.Context, db DB) ([]string, error) {
	const sqlstr = `SELECT user_id FROM user_llms WHERE auto_latest_trend_enabled = true`
	return queryStringSlice(ctx, db, sqlstr)
}

// UserIDsWithSemanticSearchEnabled はsemantic_search_enabledがtrueのユーザーIDの一覧を返す
func UserIDsWithSemanticSearchEnabled(ctx context.Context, db DB) ([]string, error) {
	const sqlstr = `SELECT user_id FROM user_llms WHERE semantic_search_enabled = true`
	return queryStringSlice(ctx, db, sqlstr)
}
