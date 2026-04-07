package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// HourlyPubSubMetric は1時間ごとのPub/Sub処理件数を表す
type HourlyPubSubMetric struct {
	Hour                      time.Time
	DailySummariesProcessed   int32
	MonthlySummariesProcessed int32
	EmbeddingsProcessed       int32
	SemanticSearchesProcessed int32
}

// UserLLMAutoSettings はユーザーの自動処理設定を表す
type UserLLMAutoSettings struct {
	AutoSummaryDaily      bool
	AutoSummaryMonthly    bool
	AutoLatestTrend       bool
	SemanticSearchEnabled bool
}

// DeleteDiariesByUserID は指定ユーザーの全日記を削除する（トランザクション内で使用）
func DeleteDiariesByUserID(ctx context.Context, db DB, userID uuid.UUID) error {
	const sqlstr = `DELETE FROM diaries WHERE user_id = $1`
	if _, err := db.ExecContext(ctx, sqlstr, userID); err != nil {
		return fmt.Errorf("failed to delete diaries for user %s: %w", userID, err)
	}
	return nil
}

// DeleteUserLLMsByUserID は指定ユーザーのLLM設定を削除する（トランザクション内で使用）
func DeleteUserLLMsByUserID(ctx context.Context, db DB, userID uuid.UUID) error {
	const sqlstr = `DELETE FROM user_llms WHERE user_id = $1`
	if _, err := db.ExecContext(ctx, sqlstr, userID); err != nil {
		return fmt.Errorf("failed to delete user LLMs for user %s: %w", userID, err)
	}
	return nil
}

// DeleteUserPasswordAuthesByUserID は指定ユーザーのパスワード認証情報を削除する（トランザクション内で使用）
func DeleteUserPasswordAuthesByUserID(ctx context.Context, db DB, userID uuid.UUID) error {
	const sqlstr = `DELETE FROM user_password_authes WHERE user_id = $1`
	if _, err := db.ExecContext(ctx, sqlstr, userID); err != nil {
		return fmt.Errorf("failed to delete user password authes for user %s: %w", userID, err)
	}
	return nil
}

// HourlyPubSubMetrics は過去24時間の1時間ごとのPub/Sub処理件数を返す
func HourlyPubSubMetrics(ctx context.Context, db DB, userID uuid.UUID) ([]*HourlyPubSubMetric, error) {
	const sqlstr = `
		WITH hours AS (
			SELECT generate_series(
				date_trunc('hour', NOW() - INTERVAL '23 hours'),
				date_trunc('hour', NOW()),
				INTERVAL '1 hour'
			) AS hour
		),
		daily_summaries AS (
			SELECT
				date_trunc('hour', to_timestamp(created_at)) as hour,
				COUNT(*) as created_count
			FROM diary_summary_days
			WHERE user_id = $1
			AND created_at >= EXTRACT(EPOCH FROM NOW() - INTERVAL '24 hours')
			GROUP BY date_trunc('hour', to_timestamp(created_at))
		),
		monthly_summaries AS (
			SELECT
				date_trunc('hour', to_timestamp(created_at)) as hour,
				COUNT(*) as created_count
			FROM diary_summary_months
			WHERE user_id = $1
			AND created_at >= EXTRACT(EPOCH FROM NOW() - INTERVAL '24 hours')
			GROUP BY date_trunc('hour', to_timestamp(created_at))
		),
		diary_embeddings AS (
			SELECT
				date_trunc('hour', created_at) as hour,
				COUNT(*) as created_count
			FROM diary_embeddings
			WHERE user_id = $1
			AND created_at >= NOW() - INTERVAL '24 hours'
			GROUP BY date_trunc('hour', created_at)
		),
		semantic_searches AS (
			SELECT
				date_trunc('hour', created_at) as hour,
				COUNT(*) as created_count
			FROM semantic_search_logs
			WHERE user_id = $1
			AND created_at >= NOW() - INTERVAL '24 hours'
			GROUP BY date_trunc('hour', created_at)
		)
		SELECT
			h.hour,
			COALESCE(ds.created_count, 0) as daily_summaries_processed,
			COALESCE(ms.created_count, 0) as monthly_summaries_processed,
			COALESCE(de.created_count, 0) as diary_embeddings_processed,
			COALESCE(ss.created_count, 0) as semantic_searches_processed
		FROM hours h
		LEFT JOIN daily_summaries ds ON h.hour = ds.hour
		LEFT JOIN monthly_summaries ms ON h.hour = ms.hour
		LEFT JOIN diary_embeddings de ON h.hour = de.hour
		LEFT JOIN semantic_searches ss ON h.hour = ss.hour
		ORDER BY h.hour
	`
	rows, err := db.QueryContext(ctx, sqlstr, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query hourly pub/sub metrics: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var metrics []*HourlyPubSubMetric
	for rows.Next() {
		var m HourlyPubSubMetric
		if err := rows.Scan(&m.Hour, &m.DailySummariesProcessed, &m.MonthlySummariesProcessed, &m.EmbeddingsProcessed, &m.SemanticSearchesProcessed); err != nil {
			return nil, fmt.Errorf("failed to scan hourly metrics: %w", err)
		}
		metrics = append(metrics, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}
	return metrics, nil
}

// TotalDailySummaryCount は指定ユーザーの日次サマリー総数を返す
func TotalDailySummaryCount(ctx context.Context, db DB, userID uuid.UUID) (int32, error) {
	const sqlstr = `SELECT COUNT(*) FROM diary_summary_days WHERE user_id = $1`
	count, err := queryCount(ctx, db, sqlstr, userID)
	return int32(count), err
}

// TotalMonthlySummaryCount は指定ユーザーの月次サマリー総数を返す
func TotalMonthlySummaryCount(ctx context.Context, db DB, userID uuid.UUID) (int32, error) {
	const sqlstr = `SELECT COUNT(*) FROM diary_summary_months WHERE user_id = $1`
	count, err := queryCount(ctx, db, sqlstr, userID)
	return int32(count), err
}

// PendingDailySummaryCount は指定ユーザーの未作成日次サマリー数を返す（今日を除く）
func PendingDailySummaryCount(ctx context.Context, db DB, userID uuid.UUID) (int32, error) {
	const sqlstr = `
		SELECT COUNT(*)
		FROM diaries d
		LEFT JOIN diary_summary_days dsd ON d.user_id = dsd.user_id AND d.date = dsd.date
		WHERE d.user_id = $1
		  AND d.date < CURRENT_DATE
		  AND (dsd.id IS NULL OR dsd.updated_at < d.updated_at)
	`
	count, err := queryCount(ctx, db, sqlstr, userID)
	return int32(count), err
}

// PendingMonthlySummaryCount は指定ユーザーの未作成月次サマリー数を返す（今月を除く）
func PendingMonthlySummaryCount(ctx context.Context, db DB, userID uuid.UUID) (int32, error) {
	const sqlstr = `
		WITH monthly_diary_stats AS (
			SELECT
				EXTRACT(YEAR FROM d.date) as year,
				EXTRACT(MONTH FROM d.date) as month,
				MAX(d.updated_at) as latest_diary_updated_at
			FROM diaries d
			WHERE d.user_id = $1
			GROUP BY EXTRACT(YEAR FROM d.date), EXTRACT(MONTH FROM d.date)
		)
		SELECT COUNT(*)
		FROM monthly_diary_stats mds
		LEFT JOIN diary_summary_months dsm ON dsm.user_id = $1
			AND dsm.year = mds.year
			AND dsm.month = mds.month
		WHERE (mds.year < EXTRACT(YEAR FROM CURRENT_DATE)
			OR (mds.year = EXTRACT(YEAR FROM CURRENT_DATE) AND mds.month < EXTRACT(MONTH FROM CURRENT_DATE)))
		AND (dsm.updated_at IS NULL OR dsm.updated_at < mds.latest_diary_updated_at)
	`
	count, err := queryCount(ctx, db, sqlstr, userID)
	return int32(count), err
}

// UserLLMAutoSettingsByUserID は指定ユーザーのLLM自動処理設定を返す
// 設定が存在しない場合は全てfalseの設定を返す
func UserLLMAutoSettingsByUserID(ctx context.Context, db DB, userID uuid.UUID) (*UserLLMAutoSettings, error) {
	ul, err := UserLlmByUserIDLlmProvider(ctx, db, userID, 1)
	if err != nil {
		// 設定が存在しない場合は全てfalse
		return &UserLLMAutoSettings{}, nil
	}
	return &UserLLMAutoSettings{
		AutoSummaryDaily:      ul.AutoSummaryDaily,
		AutoSummaryMonthly:    ul.AutoSummaryMonthly,
		AutoLatestTrend:       ul.AutoLatestTrendEnabled,
		SemanticSearchEnabled: ul.SemanticSearchEnabled,
	}, nil
}

// TotalEmbeddingCount は指定ユーザーのembedding総数を返す
func TotalEmbeddingCount(ctx context.Context, db DB, userID uuid.UUID) (int32, error) {
	const sqlstr = `SELECT COUNT(*) FROM diary_embeddings WHERE user_id = $1`
	count, err := queryCount(ctx, db, sqlstr, userID)
	return int32(count), err
}

// PendingEmbeddingCount は指定ユーザーのembedding未生成日記数を返す
func PendingEmbeddingCount(ctx context.Context, db DB, userID uuid.UUID) (int32, error) {
	const sqlstr = `
		SELECT COUNT(*)
		FROM diaries d
		WHERE d.user_id = $1
		  AND NOT EXISTS (
		    SELECT 1 FROM diary_embeddings e WHERE e.diary_id = d.id
		  )
	`
	count, err := queryCount(ctx, db, sqlstr, userID)
	return int32(count), err
}
