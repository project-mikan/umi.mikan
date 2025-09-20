package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	clientConfig := &genai.ClientConfig{
		APIKey: apiKey,
	}
	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
	}, nil
}

func (g *GeminiClient) Close() error {
	return nil
}

func (g *GeminiClient) GenerateSummary(ctx context.Context, diaryContent string) (string, error) {
	prompt := fmt.Sprintf(`以下の日記の内容を読んで、月間サマリーを生成してください。
サマリーは以下の要件を満たしてください：
- 冒頭に箇条書きで特筆すべき日付と内容を最大3つ挙げる(Markdown非対応のため箇条書きは「- 」で始める)
- 次にその月全体の傾向を300文字以内で簡潔にまとめる

日記の内容:
%s

`, diaryContent)

	contents := genai.Text(prompt)

	resp, err := g.client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// The response parts contain the generated text
	if textPart := resp.Candidates[0].Content.Parts[0]; textPart != nil {
		return textPart.Text, nil
	}

	return "", fmt.Errorf("unexpected content type")
}

func (g *GeminiClient) GenerateDailySummary(ctx context.Context, diaryContent string) (string, error) {
	prompt := fmt.Sprintf(`以下の1日の日記の内容を読んで、日次サマリーを生成してください。
サマリーは以下の要件を満たしてください：
- 最大3つまで要点を列挙(Markdown非対応のため箇条書きは「- 」で始める)
- 出てきた人物を列挙(Markdown非対応のため箇条書きは「- 」で始める)

日記の内容:
%s

`, diaryContent)

	contents := genai.Text(prompt)

	resp, err := g.client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// The response parts contain the generated text
	if textPart := resp.Candidates[0].Content.Parts[0]; textPart != nil {
		return textPart.Text, nil
	}

	return "", fmt.Errorf("unexpected content type")
}
