package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
)

// dateLayout はMCPツールの入出力で使う日付フォーマット（YYYY-MM-DD）
const dateLayout = "2006-01-02"

// maxRangeDays は一度のリクエストで取得できる日数の上限（過大なクエリを防ぐため）
const maxRangeDays = 366

// GetDiaryEntriesByRangeInput は get_diary_entries_by_range ツールの入力
type GetDiaryEntriesByRangeInput struct {
	From string `json:"from" jsonschema:"取得開始日（YYYY-MM-DD形式、この日を含む）"`
	To   string `json:"to" jsonschema:"取得終了日（YYYY-MM-DD形式、この日を含む）。fromと同じかそれ以降の日付"`
}

// GetDiaryEntriesByRangeOutput は get_diary_entries_by_range ツールの出力
type GetDiaryEntriesByRangeOutput struct {
	Entries []DiaryEntryOutput `json:"entries" jsonschema:"範囲内の日記エントリ一覧（日付昇順）"`
}

// DiaryEntryOutput はMCPツールが返す日記エントリ1件分
type DiaryEntryOutput struct {
	ID        string `json:"id" jsonschema:"日記ID"`
	Date      string `json:"date" jsonschema:"日記の日付（YYYY-MM-DD形式）"`
	Content   string `json:"content" jsonschema:"日記本文"`
	CreatedAt int64  `json:"createdAt" jsonschema:"作成日時（UnixTime秒）"`
	UpdatedAt int64  `json:"updatedAt" jsonschema:"更新日時（UnixTime秒）"`
}

func getDiaryEntriesByRangeHandler(diaryService *diary.DiaryEntry) mcp.ToolHandlerFor[GetDiaryEntriesByRangeInput, GetDiaryEntriesByRangeOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input GetDiaryEntriesByRangeInput) (*mcp.CallToolResult, GetDiaryEntriesByRangeOutput, error) {
		userID, err := userIDFromContext(ctx)
		if err != nil {
			return nil, GetDiaryEntriesByRangeOutput{}, err
		}

		from, err := time.Parse(dateLayout, input.From)
		if err != nil {
			return nil, GetDiaryEntriesByRangeOutput{}, fmt.Errorf("invalid from date %q: must be YYYY-MM-DD format", input.From)
		}
		to, err := time.Parse(dateLayout, input.To)
		if err != nil {
			return nil, GetDiaryEntriesByRangeOutput{}, fmt.Errorf("invalid to date %q: must be YYYY-MM-DD format", input.To)
		}
		if to.Before(from) {
			return nil, GetDiaryEntriesByRangeOutput{}, fmt.Errorf("to (%s) must not be before from (%s)", input.To, input.From)
		}
		if to.Sub(from) > maxRangeDays*24*time.Hour {
			return nil, GetDiaryEntriesByRangeOutput{}, fmt.Errorf("range too large: at most %d days can be requested at once", maxRangeDays)
		}

		diaries, err := diaryService.GetDiaryEntriesByDateRange(ctx, userID, from, to)
		if err != nil {
			return nil, GetDiaryEntriesByRangeOutput{}, friendlyError(err)
		}

		return nil, GetDiaryEntriesByRangeOutput{Entries: toDiaryEntryOutputs(diaries)}, nil
	}
}
