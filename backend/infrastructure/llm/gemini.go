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

func (g *GeminiClient) GenerateLatestTrend(ctx context.Context, diaryContent string) (string, error) {
	prompt := fmt.Sprintf(`以下は直近3日間の日記です。この短い期間の傾向を分析し、わかりやすく要約してください。

【出力形式】
以下の形式で出力してください（Markdownは使用しないでください）：

## 最近のあなた

<1行で全体的な様子を簡潔に表現>

### 体調・気分
<体調や気分の傾向を2-3文で>

### 活動・行動
<よくしていた活動や行動パターンを2-3文で>

### 気になること
<特筆すべき変化や注目すべき点を1-2文で>

【要件】
- 「##」「###」の見出しマーカーはそのまま出力してください
- その他のMarkdown記法（太字、リンクなど）は使用しないでください
- 具体的な日付や曜日は含めず、傾向のみを記述
- 3日間という短期間なので、大きな傾向よりも最近の様子に注目してください
- 客観的かつ優しい語り口で
- 合計300-400字程度

【日記の内容】
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
