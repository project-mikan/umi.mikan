package llm

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
	}, nil
}

func (g *GeminiClient) Close() error {
	return g.client.Close()
}

func (g *GeminiClient) GenerateSummary(ctx context.Context, diaryContent string) (string, error) {
	model := g.client.GenerativeModel("gemini-1.5-flash")

	prompt := fmt.Sprintf(`以下の日記の内容を読んで、月間サマリーを生成してください。
サマリーは以下の要件を満たしてください：
- 300文字以内で簡潔にまとめる
- その月の主要な出来事や感情を要約する
- ポジティブな視点で書く
- 読み返したときに思い出しやすい内容にする

日記の内容:
%s

月間サマリー:`, diaryContent)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	summary, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected content type")
	}

	return string(summary), nil
}
