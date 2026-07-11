package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
)

// SearchDiaryEntriesFulltextInput は search_diary_entries_fulltext ツールの入力
type SearchDiaryEntriesFulltextInput struct {
	Keyword string `json:"keyword" jsonschema:"検索キーワード。登録済みの人物・エンティティ名の場合、関連する別名やエイリアスにも自動展開して検索される"`
}

// SearchDiaryEntriesFulltextOutput は search_diary_entries_fulltext ツールの出力
type SearchDiaryEntriesFulltextOutput struct {
	Entries          []DiaryEntryOutput `json:"entries" jsonschema:"キーワードにマッチした日記エントリ一覧"`
	ExpandedKeywords []string           `json:"expandedKeywords" jsonschema:"エンティティ展開によって追加検索されたキーワード一覧"`
}

func searchDiaryEntriesFulltextHandler(diaryService *diary.DiaryEntry) mcp.ToolHandlerFor[SearchDiaryEntriesFulltextInput, SearchDiaryEntriesFulltextOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SearchDiaryEntriesFulltextInput) (*mcp.CallToolResult, SearchDiaryEntriesFulltextOutput, error) {
		userID, err := userIDFromContext(ctx)
		if err != nil {
			return nil, SearchDiaryEntriesFulltextOutput{}, err
		}
		if input.Keyword == "" {
			return nil, SearchDiaryEntriesFulltextOutput{}, fmt.Errorf("keyword is required")
		}

		result, err := diaryService.SearchDiaryEntriesByUserID(ctx, userID, input.Keyword)
		if err != nil {
			return nil, SearchDiaryEntriesFulltextOutput{}, friendlyError(err)
		}

		return nil, SearchDiaryEntriesFulltextOutput{
			Entries:          toDiaryEntryOutputs(result.Entries),
			ExpandedKeywords: result.ExpandedKeywords,
		}, nil
	}
}

// SearchDiaryEntriesFuzzyInput は search_diary_entries_fuzzy ツールの入力
type SearchDiaryEntriesFuzzyInput struct {
	Query string `json:"query" jsonschema:"自然言語の検索クエリ（例: 「最近旅行に行った時の話」）"`
	// Limit はポインタ型にすることで「未指定」（nil、デフォルト10件を使う）と
	// 「明示的に0を指定」（0件、クライアントの意図通りに0件を返す）を区別する。
	// int の omitempty だとこの2つが同じ値（0）になり区別できなかった。
	Limit *int `json:"limit,omitempty" jsonschema:"返す件数の上限（省略時は10、最大50）"`
}

// SemanticSearchResultOutput はあいまい検索1件分の出力
type SemanticSearchResultOutput struct {
	DiaryID      string  `json:"diaryId" jsonschema:"日記ID"`
	Date         string  `json:"date" jsonschema:"日記の日付（YYYY-MM-DD形式）"`
	Snippet      string  `json:"snippet" jsonschema:"マッチした本文の抜粋"`
	Similarity   float32 `json:"similarity" jsonschema:"クエリとの類似度（0〜1）"`
	ChunkSummary string  `json:"chunkSummary" jsonschema:"マッチした箇所の要約"`
	ChunkCount   int     `json:"chunkCount" jsonschema:"日記内のチャンク総数"`
}

// SearchDiaryEntriesFuzzyOutput は search_diary_entries_fuzzy ツールの出力
type SearchDiaryEntriesFuzzyOutput struct {
	Results        []SemanticSearchResultOutput `json:"results" jsonschema:"クエリと意味的に類似した日記の一覧（類似度降順）"`
	EmbeddingModel string                       `json:"embeddingModel" jsonschema:"埋め込みベクトル生成に使用されたモデル"`
	ChunkModel     string                       `json:"chunkModel" jsonschema:"チャンク分割に使用されたモデル"`
}

// maxFuzzyLimit は search_diary_entries_fuzzy ツールで返せる件数の上限。
// embedding API への過大なリクエストを防ぐため、jsonschema の記述とコードを一致させる。
const maxFuzzyLimit = 50

func searchDiaryEntriesFuzzyHandler(diaryService *diary.DiaryEntry) mcp.ToolHandlerFor[SearchDiaryEntriesFuzzyInput, SearchDiaryEntriesFuzzyOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SearchDiaryEntriesFuzzyInput) (*mcp.CallToolResult, SearchDiaryEntriesFuzzyOutput, error) {
		userID, err := userIDFromContext(ctx)
		if err != nil {
			return nil, SearchDiaryEntriesFuzzyOutput{}, err
		}
		if input.Query == "" {
			return nil, SearchDiaryEntriesFuzzyOutput{}, fmt.Errorf("query is required")
		}

		// limit省略時（nil）はデフォルト10件、明示的な指定（0件を含む）はその値をそのまま使う。
		// ただしjsonschemaの「最大50」をコードでも強制する（0未満は0に正規化）。
		limit := 10
		if input.Limit != nil {
			limit = max(*input.Limit, 0)
			if limit > maxFuzzyLimit {
				limit = maxFuzzyLimit
			}
		}
		outcome, err := diaryService.SearchDiaryEntriesSemanticByUserID(ctx, userID, input.Query, limit)
		if err != nil {
			return nil, SearchDiaryEntriesFuzzyOutput{}, friendlyError(err)
		}

		results := make([]SemanticSearchResultOutput, 0, len(outcome.Results))
		for _, r := range outcome.Results {
			results = append(results, SemanticSearchResultOutput{
				DiaryID:      r.DiaryID.String(),
				Date:         r.Date.Format(dateLayout),
				Snippet:      r.Snippet,
				Similarity:   r.Similarity,
				ChunkSummary: r.ChunkSummary,
				ChunkCount:   r.ChunkCount,
			})
		}

		return nil, SearchDiaryEntriesFuzzyOutput{
			Results:        results,
			EmbeddingModel: outcome.EmbeddingModel,
			ChunkModel:     outcome.ChunkModel,
		}, nil
	}
}
