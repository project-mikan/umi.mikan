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
	Health       string `json:"health"`        // 体調: "bad", "slight", "normal", "good"
	HealthReason string `json:"health_reason"` // 体調の理由（10文字以内）
	Mood         string `json:"mood"`          // 気分: "bad", "slight", "normal", "good"
	MoodReason   string `json:"mood_reason"`   // 気分の理由（10文字以内）
	Activities   string `json:"activities"`    // 活動・行動（箇条書き・階層構造のテキスト）
}

func (g *GeminiClient) GenerateLatestTrend(ctx context.Context, diaryContent string, yesterday string) (string, error) {
	prompt := fmt.Sprintf(`以下は複数日分の日記です。**前日（最も新しい日）を最も重視**し、それ以前の日記は参考程度に使用して、傾向を分析してください。

【重要な分析方針】
- **昨日の日付は%sです**
- **前日（昨日）と前々日（一昨日）を比較**してください
- 体調と気分は、一昨日と比べてどうだったかという視点で評価してください
- 理由フィールドには、**一昨日との比較**と**具体的な理由**の両方を含めてください
- 理由のフォーマット: 「比較|具体的理由」（例: 前日より穏やか|仕事成果あり）

【出力形式】
以下のJSON形式で出力してください：

{
  "health": "<昨日の体調を4段階で評価: bad / slight / normal / good>",
  "health_reason": "<比較|具体的理由 の形式で20文字以内（例: 前日より改善|よく休めた）>",
  "mood": "<昨日の気分を4段階で評価: bad / slight / normal / good>",
  "mood_reason": "<比較|具体的理由 の形式で20文字以内（例: 前日より穏やか|仕事成果あり）>",
  "activities": "<活動・行動を箇条書き・階層構造で記述>"
}

【評価基準】
health（昨日の体調）:
- bad: 体調が悪い、不調、病気、疲労が激しい
- slight: やや体調が悪い、少し疲れている
- normal: 普通、特に問題なし
- good: 体調が良い、元気、健康
- **一昨日と比較して評価してください**

health_reason（体調の理由）:
- **必ず20文字以内**で記述してください
- **フォーマット**: 「比較|具体的理由」
- **比較**: 一昨日との比較を簡潔に表現（例: 前日より改善、前日より悪化、前日と同様）
- **具体的理由**: なぜそうなったかの理由（例: よく休めた、疲労蓄積、運動した）
- 例: 「前日より改善|よく休めた」「前日より悪化|睡眠不足」「前日と同様|通常通り」

mood（昨日の気分）:
- bad: 気分が悪い、落ち込んでいる、ストレスが多い
- slight: やや気分が悪い、少し憂鬱
- normal: 普通、特に問題なし
- good: 気分が良い、前向き、充実している
- **一昨日と比較して評価してください**

mood_reason（気分の理由）:
- **必ず20文字以内**で記述してください
- **フォーマット**: 「比較|具体的理由」
- **比較**: 一昨日との比較を簡潔に表現（例: 前日より穏やか、前日より低下、前日と同様）
- **具体的理由**: なぜそうなったかの理由（例: 仕事成果あり、ストレス増、楽しい出来事）
- 例: 「前日より穏やか|仕事成果あり」「前日より低下|締切接近」「前日と同様|通常通り」

activities（活動・行動）:
- **必ず改行区切りの箇条書き**で記述してください
- 各項目は改行（\n）で区切り、行頭に「- 」を付けてください
- 階層構造の場合は、ネストレベルごとに半角スペース2つのインデントを追加
- 出力例（改行あり）:
  - 運動
    - 朝のランニング
    - ストレッチ
  - 仕事
    - プロジェクトミーティング
- **重要**: 各項目の間には必ず改行（\n）を入れてください
- **重要**: 「- 項目1- 項目2」のように連結しないでください
- Markdownは使用しないでください

【要件】
- 必ずJSON形式で出力してください
- Markdownは使用しないでください
- 具体的な日付や曜日は含めず、傾向のみを記述
- **前日（昨日）を最も重視**し、最近の様子に注目してください
- **前日（昨日）と前々日（一昨日）を比較**して評価してください
- 客観的かつ優しい語り口で
- 各フィールドは空文字列にせず、必ず内容を記述してください
- health と mood は必ず "bad", "slight", "normal", "good" のいずれか1つを選んでください
- health_reason と mood_reason は**必ず20文字以内**で記述してください
- **health_reason と mood_reason は「比較|具体的理由」の形式で記述してください**
- **比較部分**: 一昨日との比較を含めてください（例: 前日より改善、前日と同様）
- **具体的理由部分**: なぜそうなったかの理由を含めてください（例: よく休めた、仕事成果あり）
- **activities フィールドは必ず改行区切り**で記述してください（連結禁止）

【日記の内容】
%s

`, yesterday, diaryContent)

	contents := genai.Text(prompt)

	// JSON出力を強制するためのスキーマを設定
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"health": {
				Type:        genai.TypeString,
				Description: "体調を4段階で評価: bad, slight, normal, good",
			},
			"health_reason": {
				Type:        genai.TypeString,
				Description: "体調の理由を10文字以内で簡潔に",
			},
			"mood": {
				Type:        genai.TypeString,
				Description: "気分を4段階で評価: bad, slight, normal, good",
			},
			"mood_reason": {
				Type:        genai.TypeString,
				Description: "気分の理由を10文字以内で簡潔に",
			},
			"activities": {
				Type:        genai.TypeString,
				Description: "活動・行動を箇条書き・階層構造で記述",
			},
		},
		Required: []string{"health", "health_reason", "mood", "mood_reason", "activities"},
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
