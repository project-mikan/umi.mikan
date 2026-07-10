package database

import (
	"context"
	"time"
)

// DiariesByUserIDAndDateRangeDays は指定ユーザーの指定日付範囲（開始日〜終了日、両端含む）の日記をdate昇順で返す。
// MCPサーバーなど、年月単位ではなく日単位で範囲を指定したい呼び出し元向け。
func DiariesByUserIDAndDateRangeDays(ctx context.Context, db DB, userID string, fromDate, toDate time.Time) ([]*Diary, error) {
	return diariesByUserIDAndDateRange(ctx, db, userID, fromDate, toDate)
}
