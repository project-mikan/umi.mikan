package llm

import "context"

// LLMClient はLLM APIクライアントのインターフェース
type LLMClient interface {
	// GenerateSummary は月間サマリーを生成
	GenerateSummary(ctx context.Context, diaryContent string) (string, error)
	// GenerateDailySummary は日次サマリーを生成
	GenerateDailySummary(ctx context.Context, diaryContent string) (string, error)
	// GenerateLatestTrend はトレンド分析を生成
	GenerateLatestTrend(ctx context.Context, diaryContent string, yesterday string) (string, error)
	// GenerateHighlights はハイライトを生成
	GenerateHighlights(ctx context.Context, diaryContent string) (string, error)
	// Close はクライアントをクローズ
	Close() error
}
