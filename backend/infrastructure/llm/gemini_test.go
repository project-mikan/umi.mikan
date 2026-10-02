package llm

import (
	"context"
	"testing"
)

func TestNewGeminiClient(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		location string
	}{
		{
			name:     "異常系: Projectを空にするとVertex AIの接続先が決まらないのでエラーになる",
			project:  "",
			location: "global",
		},
		{
			name:     "異常系: Locationを空にするとVertex AIの接続先が決まらないのでエラーになる",
			project:  "umi-mikan-test",
			location: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeminiClient(context.Background(), tt.project, tt.location)
			if err == nil {
				t.Fatal("エラーが返ることを期待したが nil だった")
			}
			if client != nil {
				t.Error("エラー時はクライアントが nil であることを期待")
			}
		})
	}
}
