package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// UpdateUserAPIKeyLastUsed はAPIキーの最終使用日時を更新する。
// 認証のたびに呼ばれるため、対象行が存在しなくてもエラーにしない。
func UpdateUserAPIKeyLastUsed(ctx context.Context, db DB, id uuid.UUID, lastUsedAt int64) error {
	const sqlstr = `UPDATE user_api_keys SET last_used_at = $2, updated_at = $2 WHERE id = $1`
	if _, err := db.ExecContext(ctx, sqlstr, id, lastUsedAt); err != nil {
		return fmt.Errorf("failed to update api key last_used_at: %w", err)
	}
	return nil
}
