package llm

import (
	"context"
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
