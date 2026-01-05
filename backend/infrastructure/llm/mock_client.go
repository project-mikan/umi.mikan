package llm

import (
	"context"
	"fmt"
)

// MockLLMClient はテスト用のモッククライアント
type MockLLMClient struct {
	GenerateSummaryFunc       func(ctx context.Context, diaryContent string) (string, error)
	GenerateDailySummaryFunc  func(ctx context.Context, diaryContent string) (string, error)
	GenerateLatestTrendFunc   func(ctx context.Context, diaryContent string, yesterday string) (string, error)
	GenerateHighlightsFunc    func(ctx context.Context, diaryContent string) (string, error)
	CloseFunc                 func() error
}

// GenerateSummary はモックの月間サマリー生成
func (m *MockLLMClient) GenerateSummary(ctx context.Context, diaryContent string) (string, error) {
	if m.GenerateSummaryFunc != nil {
		return m.GenerateSummaryFunc(ctx, diaryContent)
	}
	return "", fmt.Errorf("GenerateSummaryFunc not implemented")
}

// GenerateDailySummary はモックの日次サマリー生成
func (m *MockLLMClient) GenerateDailySummary(ctx context.Context, diaryContent string) (string, error) {
	if m.GenerateDailySummaryFunc != nil {
		return m.GenerateDailySummaryFunc(ctx, diaryContent)
	}
	return "", fmt.Errorf("GenerateDailySummaryFunc not implemented")
}

// GenerateLatestTrend はモックのトレンド分析生成
func (m *MockLLMClient) GenerateLatestTrend(ctx context.Context, diaryContent string, yesterday string) (string, error) {
	if m.GenerateLatestTrendFunc != nil {
		return m.GenerateLatestTrendFunc(ctx, diaryContent, yesterday)
	}
	return "", fmt.Errorf("GenerateLatestTrendFunc not implemented")
}

// GenerateHighlights はモックのハイライト生成
func (m *MockLLMClient) GenerateHighlights(ctx context.Context, diaryContent string) (string, error) {
	if m.GenerateHighlightsFunc != nil {
		return m.GenerateHighlightsFunc(ctx, diaryContent)
	}
	return "", fmt.Errorf("GenerateHighlightsFunc not implemented")
}

// Close はモックのクローズ処理
func (m *MockLLMClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
