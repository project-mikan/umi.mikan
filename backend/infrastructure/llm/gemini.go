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
- Markdownは非対応
- 冒頭に箇条書きで特筆すべき日付と内容を最大3つ挙げる(箇条書きは「n日：」で始める。月は不要)
- 次にその月全体の傾向を300文字以内で簡潔にまとめる

形式は以下の通りにしてください：
n日：箇条書き1
n日：箇条書き2
n日：箇条書き3

<300文字以内の月全体の傾向>


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
- Markdownは非対応
- 最大3つまで簡潔に要点を列挙(箇条書きは「- 」で始める)
- 出てきた人物を文脈から重要な順に最大3人列挙(箇条書きは「- 」で始める)

形式は以下の通りにしてください：
- 箇条書き1
- 箇条書き2
- 箇条書き3

重要そうな人
- 人物1
- 人物2
- 人物3

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

// LatestTrendAnalysis はトレンド分析のJSON構造体
type LatestTrendAnalysis struct {
	OverallSummary string `json:"overall_summary"` // 全体的な様子（1行）
	HealthMood     string `json:"health_mood"`     // 体調・気分の傾向（2-3文）
	Activities     string `json:"activities"`      // 活動・行動パターン（2-3文）
	Concerns       string `json:"concerns"`        // 気になること（1-2文）
}

func (g *GeminiClient) GenerateLatestTrend(ctx context.Context, diaryContent string) (string, error) {
	prompt := fmt.Sprintf(`以下は直近3日間の日記です。この短い期間の傾向を分析し、わかりやすく要約してください。

【出力形式】
以下のJSON形式で出力してください：

{
  "overall_summary": "<1行で全体的な様子を簡潔に表現>",
  "health_mood": "<体調や気分の傾向を2-3文で>",
  "activities": "<よくしていた活動や行動パターンを2-3文で>",
  "concerns": "<特筆すべき変化や注目すべき点を1-2文で>"
}

【要件】
- 必ずJSON形式で出力してください
- Markdownは使用しないでください
- 具体的な日付や曜日は含めず、傾向のみを記述
- 3日間という短期間なので、大きな傾向よりも最近の様子に注目してください
- 客観的かつ優しい語り口で
- 各フィールドは空文字列にせず、必ず内容を記述してください

【日記の内容】
%s

`, diaryContent)

	contents := genai.Text(prompt)

	// JSON出力を強制するためのスキーマを設定
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"overall_summary": {
				Type:        genai.TypeString,
				Description: "全体的な様子を1行で簡潔に表現",
			},
			"health_mood": {
				Type:        genai.TypeString,
				Description: "体調や気分の傾向を2-3文で",
			},
			"activities": {
				Type:        genai.TypeString,
				Description: "よくしていた活動や行動パターンを2-3文で",
			},
			"concerns": {
				Type:        genai.TypeString,
				Description: "特筆すべき変化や注目すべき点を1-2文で",
			},
		},
		Required: []string{"overall_summary", "health_mood", "activities", "concerns"},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	resp, err := g.client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
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
