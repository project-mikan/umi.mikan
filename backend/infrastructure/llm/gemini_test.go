package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// TestNewGeminiClient は GeminiClient の初期化をテスト
func TestNewGeminiClient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "空のAPIキー",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeminiClient(ctx, tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGeminiClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewGeminiClient() returned nil client")
			}
			if client != nil {
				_ = client.Close()
			}
		})
	}
}

// TestGeminiClient_Close は Close メソッドをテスト
func TestGeminiClient_Close(t *testing.T) {
	// Close は現在何もしないが、将来の実装のためにテストを用意
	client := &GeminiClient{}
	if err := client.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// TestLatestTrendAnalysis_Struct は LatestTrendAnalysis 構造体の定義をテスト
func TestLatestTrendAnalysis_Struct(t *testing.T) {
	// 構造体のフィールドが正しく定義されているかテスト
	analysis := LatestTrendAnalysis{
		Health:       "good",
		HealthReason: "よく休めた",
		Mood:         "good",
		MoodReason:   "仕事成果あり",
		Activities:   "- 朝のランニング\n- プロジェクトミーティング",
	}

	if analysis.Health != "good" {
		t.Errorf("Health = %v, want %v", analysis.Health, "good")
	}
	if analysis.HealthReason != "よく休めた" {
		t.Errorf("HealthReason = %v, want %v", analysis.HealthReason, "よく休めた")
	}
	if analysis.Mood != "good" {
		t.Errorf("Mood = %v, want %v", analysis.Mood, "good")
	}
	if analysis.MoodReason != "仕事成果あり" {
		t.Errorf("MoodReason = %v, want %v", analysis.MoodReason, "仕事成果あり")
	}
	if analysis.Activities != "- 朝のランニング\n- プロジェクトミーティング" {
		t.Errorf("Activities = %v, want %v", analysis.Activities, "- 朝のランニング\n- プロジェクトミーティング")
	}
}

// TestMockLLMClient_GenerateSummary はモックを使用した GenerateSummary のテスト
func TestMockLLMClient_GenerateSummary(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		diaryContent  string
		mockResponse  string
		mockError     error
		expectedError bool
	}{
		{
			name:         "正常系: 月間サマリーが生成される",
			diaryContent: "今月は色々なことがありました。",
			mockResponse: "1日：重要な出来事1\n5日：重要な出来事2\n10日：重要な出来事3\n\n今月は充実した一ヶ月でした。",
			mockError:    nil,
			expectedError: false,
		},
		{
			name:          "異常系: API呼び出しエラー",
			diaryContent:  "今月は色々なことがありました。",
			mockResponse:  "",
			mockError:     fmt.Errorf("API error"),
			expectedError: true,
		},
		{
			name:          "異常系: 空のコンテンツ",
			diaryContent:  "",
			mockResponse:  "1日：特になし\n\n特に変化のない月でした。",
			mockError:     nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLLMClient{
				GenerateSummaryFunc: func(ctx context.Context, diaryContent string) (string, error) {
					if tt.mockError != nil {
						return "", tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			result, err := mock.GenerateSummary(ctx, tt.diaryContent)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.mockResponse {
					t.Errorf("expected %q, got %q", tt.mockResponse, result)
				}
			}
		})
	}
}

// TestMockLLMClient_GenerateDailySummary はモックを使用した GenerateDailySummary のテスト
func TestMockLLMClient_GenerateDailySummary(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		diaryContent  string
		mockResponse  string
		mockError     error
		expectedError bool
	}{
		{
			name:         "正常系: 日次サマリーが生成される",
			diaryContent: "今日は会議があり、プロジェクトが進展した。",
			mockResponse: "- 会議でプロジェクト進展\n- 重要な決定事項あり\n\n重要そうな人\n- 田中さん\n- 佐藤さん",
			mockError:    nil,
			expectedError: false,
		},
		{
			name:          "異常系: API呼び出しエラー",
			diaryContent:  "今日は会議があり、プロジェクトが進展した。",
			mockResponse:  "",
			mockError:     fmt.Errorf("API error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLLMClient{
				GenerateDailySummaryFunc: func(ctx context.Context, diaryContent string) (string, error) {
					if tt.mockError != nil {
						return "", tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			result, err := mock.GenerateDailySummary(ctx, tt.diaryContent)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.mockResponse {
					t.Errorf("expected %q, got %q", tt.mockResponse, result)
				}
			}
		})
	}
}

// TestMockLLMClient_GenerateLatestTrend はモックを使用した GenerateLatestTrend のテスト
func TestMockLLMClient_GenerateLatestTrend(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		diaryContent  string
		yesterday     string
		mockResponse  string
		mockError     error
		expectedError bool
		validateJSON  bool
	}{
		{
			name:         "正常系: トレンド分析が生成される（JSON形式）",
			diaryContent: "今日は体調が良く、仕事も順調でした。",
			yesterday:    "2024-01-15",
			mockResponse: `{"health":"good","health_reason":"前日より改善|よく休めた","mood":"good","mood_reason":"前日より穏やか|仕事成果あり","activities":"- 朝のランニング\n- プロジェクトミーティング"}`,
			mockError:    nil,
			expectedError: false,
			validateJSON:  true,
		},
		{
			name:          "異常系: API呼び出しエラー",
			diaryContent:  "今日は体調が良く、仕事も順調でした。",
			yesterday:     "2024-01-15",
			mockResponse:  "",
			mockError:     fmt.Errorf("API error"),
			expectedError: true,
			validateJSON:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLLMClient{
				GenerateLatestTrendFunc: func(ctx context.Context, diaryContent string, yesterday string) (string, error) {
					if tt.mockError != nil {
						return "", tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			result, err := mock.GenerateLatestTrend(ctx, tt.diaryContent, tt.yesterday)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.mockResponse {
					t.Errorf("expected %q, got %q", tt.mockResponse, result)
				}

				// JSON形式の検証
				if tt.validateJSON {
					var trend LatestTrendAnalysis
					if err := json.Unmarshal([]byte(result), &trend); err != nil {
						t.Errorf("failed to parse JSON: %v", err)
					}
					// 必須フィールドの検証
					if trend.Health == "" || trend.Mood == "" {
						t.Error("health or mood is empty")
					}
				}
			}
		})
	}
}

// TestMockLLMClient_GenerateHighlights はモックを使用した GenerateHighlights のテスト
func TestMockLLMClient_GenerateHighlights(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		diaryContent  string
		mockResponse  string
		mockError     error
		expectedError bool
		validateJSON  bool
	}{
		{
			name:         "正常系: ハイライトが生成される（JSON配列形式）",
			diaryContent: "今日は素晴らしい一日でした。朝から気分が良く、午後には重要なプレゼンテーションを成功させました。",
			mockResponse: `[{"start":0,"end":15,"text":"今日は素晴らしい一日でした。"},{"start":30,"end":55,"text":"重要なプレゼンテーションを成功させました。"}]`,
			mockError:    nil,
			expectedError: false,
			validateJSON:  true,
		},
		{
			name:          "異常系: API呼び出しエラー",
			diaryContent:  "今日は素晴らしい一日でした。",
			mockResponse:  "",
			mockError:     fmt.Errorf("API error"),
			expectedError: true,
			validateJSON:  false,
		},
		{
			name:          "正常系: ハイライトなし（空配列）",
			diaryContent:  "今日は特に何もなかった。",
			mockResponse:  `[]`,
			mockError:     nil,
			expectedError: false,
			validateJSON:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLLMClient{
				GenerateHighlightsFunc: func(ctx context.Context, diaryContent string) (string, error) {
					if tt.mockError != nil {
						return "", tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			result, err := mock.GenerateHighlights(ctx, tt.diaryContent)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.mockResponse {
					t.Errorf("expected %q, got %q", tt.mockResponse, result)
				}

				// JSON配列形式の検証
				if tt.validateJSON {
					var highlights []map[string]interface{}
					if err := json.Unmarshal([]byte(result), &highlights); err != nil {
						t.Errorf("failed to parse JSON array: %v", err)
					}
					// 各ハイライトの必須フィールドを検証
					for i, h := range highlights {
						if _, ok := h["start"]; !ok {
							t.Errorf("highlight[%d] missing 'start' field", i)
						}
						if _, ok := h["end"]; !ok {
							t.Errorf("highlight[%d] missing 'end' field", i)
						}
						if _, ok := h["text"]; !ok {
							t.Errorf("highlight[%d] missing 'text' field", i)
						}
					}
				}
			}
		})
	}
}

// TestMockLLMClient_Close はモックの Close メソッドをテスト
func TestMockLLMClient_Close(t *testing.T) {
	tests := []struct {
		name      string
		closeFunc func() error
		wantErr   bool
	}{
		{
			name:      "正常系: CloseFunc未設定（デフォルト動作）",
			closeFunc: nil,
			wantErr:   false,
		},
		{
			name: "正常系: CloseFunc設定済み",
			closeFunc: func() error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "異常系: Closeエラー",
			closeFunc: func() error {
				return fmt.Errorf("close error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockLLMClient{
				CloseFunc: tt.closeFunc,
			}

			err := mock.Close()

			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMockLLMClient_Interface はモックがインターフェースを実装しているかテスト
func TestMockLLMClient_Interface(t *testing.T) {
	var _ LLMClient = &MockLLMClient{}
	var _ LLMClient = &GeminiClient{}
}

// TestMockLLMClient_NotImplemented は未実装の関数呼び出しをテスト
func TestMockLLMClient_NotImplemented(t *testing.T) {
	ctx := context.Background()
	mock := &MockLLMClient{} // 何も設定しない

	t.Run("GenerateSummary未実装", func(t *testing.T) {
		_, err := mock.GenerateSummary(ctx, "test")
		if err == nil {
			t.Error("expected error for unimplemented GenerateSummary")
		}
		if !strings.Contains(err.Error(), "not implemented") {
			t.Errorf("expected 'not implemented' error, got: %v", err)
		}
	})

	t.Run("GenerateDailySummary未実装", func(t *testing.T) {
		_, err := mock.GenerateDailySummary(ctx, "test")
		if err == nil {
			t.Error("expected error for unimplemented GenerateDailySummary")
		}
		if !strings.Contains(err.Error(), "not implemented") {
			t.Errorf("expected 'not implemented' error, got: %v", err)
		}
	})

	t.Run("GenerateLatestTrend未実装", func(t *testing.T) {
		_, err := mock.GenerateLatestTrend(ctx, "test", "2024-01-01")
		if err == nil {
			t.Error("expected error for unimplemented GenerateLatestTrend")
		}
		if !strings.Contains(err.Error(), "not implemented") {
			t.Errorf("expected 'not implemented' error, got: %v", err)
		}
	})

	t.Run("GenerateHighlights未実装", func(t *testing.T) {
		_, err := mock.GenerateHighlights(ctx, "test")
		if err == nil {
			t.Error("expected error for unimplemented GenerateHighlights")
		}
		if !strings.Contains(err.Error(), "not implemented") {
			t.Errorf("expected 'not implemented' error, got: %v", err)
		}
	})
}
