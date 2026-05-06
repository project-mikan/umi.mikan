package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UpsertMonthlySummaryError はLLM生成の永続的なエラーをdiary_summary_monthsテーブルに保存する
// error_reasonが設定された月はスケジューラーに再キューイングされない（日記更新がない限り）
func UpsertMonthlySummaryError(ctx context.Context, db DB, userID uuid.UUID, year, month int, errorReason string) error {
	const sqlstr = `
		INSERT INTO diary_summary_months (id, user_id, year, month, summary, error_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '', $5, $6, $7)
		ON CONFLICT (user_id, year, month) DO UPDATE SET
			summary = '',
			error_reason = EXCLUDED.error_reason,
			updated_at = EXCLUDED.updated_at
	`
	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx, sqlstr, uuid.New(), userID, year, month, errorReason, now, now); err != nil {
		return fmt.Errorf("failed to upsert monthly summary error: %w", err)
	}
	return nil
}
