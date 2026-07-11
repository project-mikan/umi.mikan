package diary

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
)

// GetDiaryEntriesByDateRange は指定ユーザーの指定日付範囲（開始日〜終了日、両端含む）の
// 日記エントリを日付昇順で返す。MCPサーバーなど、月単位ではなく日単位で範囲を指定したい
// 呼び出し元向けの共通ロジック。
func (s *DiaryEntry) GetDiaryEntriesByDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*database.Diary, error) {
	return database.DiariesByUserIDAndDateRangeDays(ctx, s.DB, userID.String(), from, to)
}
