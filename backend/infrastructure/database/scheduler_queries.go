package database

import (
	"context"
	"fmt"
	"time"
)

// YearMonth は年月を表す
type YearMonth struct {
	Year  int
	Month int
}

// DiaryDatesNeedingDailySummary は指定ユーザーの日次サマリが未生成または古い日付を返す
// 文字数1000以上の日記のみ対象（今日は除く）
func DiaryDatesNeedingDailySummary(ctx context.Context, db DB, userID string) ([]time.Time, error) {
	const sqlstr = `
		SELECT d.date
		FROM diaries d
		LEFT JOIN diary_summary_days dsd ON d.user_id = dsd.user_id AND d.date = dsd.date
		WHERE d.user_id = $1
		  AND d.date < CURRENT_DATE
		  AND LENGTH(d.content) >= 1000
		  AND (dsd.id IS NULL OR dsd.updated_at < d.updated_at)
		ORDER BY d.date
	`
	rows, err := db.QueryContext(ctx, sqlstr, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query diary dates needing daily summary: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var dates []time.Time
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, fmt.Errorf("failed to scan date: %w", err)
		}
		dates = append(dates, date)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return dates, nil
}

// MonthsNeedingMonthlySummary は指定ユーザーの月次サマリが未生成または古い年月を返す
// 1件以上の日記がある月のみ対象（今月は除く）
func MonthsNeedingMonthlySummary(ctx context.Context, db DB, userID string) ([]YearMonth, error) {
	const sqlstr = `
		WITH monthly_diary_stats AS (
			SELECT
				EXTRACT(YEAR FROM d.date) as year,
				EXTRACT(MONTH FROM d.date) as month,
				MAX(d.updated_at) as latest_diary_updated_at,
				COUNT(*) as diary_count
			FROM diaries d
			WHERE d.user_id = $1
			GROUP BY EXTRACT(YEAR FROM d.date), EXTRACT(MONTH FROM d.date)
			HAVING COUNT(*) >= 1
		)
		SELECT mds.year, mds.month
		FROM monthly_diary_stats mds
		LEFT JOIN diary_summary_months dsm ON dsm.user_id = $1
			AND dsm.year = mds.year
			AND dsm.month = mds.month
		WHERE (mds.year < EXTRACT(YEAR FROM CURRENT_DATE)
			OR (mds.year = EXTRACT(YEAR FROM CURRENT_DATE) AND mds.month < EXTRACT(MONTH FROM CURRENT_DATE)))
		AND (dsm.updated_at IS NULL OR dsm.updated_at < mds.latest_diary_updated_at)
		ORDER BY mds.year, mds.month
	`
	rows, err := db.QueryContext(ctx, sqlstr, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query months needing monthly summary: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var months []YearMonth
	for rows.Next() {
		var year, month int
		if err := rows.Scan(&year, &month); err != nil {
			return nil, fmt.Errorf("failed to scan year/month: %w", err)
		}
		months = append(months, YearMonth{Year: year, Month: month})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return months, nil
}

// DiaryCountInMonth は指定ユーザーの指定年月の日記件数を返す
func DiaryCountInMonth(ctx context.Context, db DB, userID string, year, month int) (int, error) {
	const sqlstr = `
		SELECT COUNT(*) FROM diaries
		WHERE user_id = $1
		AND EXTRACT(YEAR FROM date) = $2
		AND EXTRACT(MONTH FROM date) = $3
	`
	return queryCount(ctx, db, sqlstr, userID, year, month)
}

// DiaryCountInDateRange は指定ユーザーの指定期間内の日記件数を返す
func DiaryCountInDateRange(ctx context.Context, db DB, userID string, from, to time.Time) (int, error) {
	const sqlstr = `SELECT COUNT(*) FROM diaries WHERE user_id = $1 AND date >= $2 AND date <= $3`
	return queryCount(ctx, db, sqlstr, userID, from, to)
}

// DiaryIDsNeedingEmbedding は指定ユーザーの指定日付でembeddingが未生成または古い日記IDを返す
// diaries.updated_atはBIGINT（ミリ秒）、diary_embeddings.updated_atはTIMESTAMPのため型変換して比較する
func DiaryIDsNeedingEmbedding(ctx context.Context, db DB, userID string, targetDate time.Time) ([]string, error) {
	const sqlstr = `
		SELECT d.id
		FROM diaries d
		WHERE d.user_id = $1
		  AND d.date = $2
		  AND (
		    NOT EXISTS (SELECT 1 FROM diary_embeddings de WHERE de.diary_id = d.id)
		    OR (SELECT MAX(de.updated_at) FROM diary_embeddings de WHERE de.diary_id = d.id) < to_timestamp(d.updated_at / 1000.0)
		  )
	`
	return queryStringSlice(ctx, db, sqlstr, userID, targetDate)
}
