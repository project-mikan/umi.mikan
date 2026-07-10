package model

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	t.Run("正常系: umi_プレフィックス付きのキーとハッシュが生成される", func(t *testing.T) {
		generated, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if !strings.HasPrefix(generated.Key, APIKeyPrefix) {
			t.Errorf("キーが %q で始まらない: %q", APIKeyPrefix, generated.Key)
		}
		// umi_(4文字) + hex 64文字
		if len(generated.Key) != len(APIKeyPrefix)+64 {
			t.Errorf("キー長: 期待 %d, 実際 %d", len(APIKeyPrefix)+64, len(generated.Key))
		}
		if generated.Hash != HashAPIKey(generated.Key) {
			t.Error("HashがHashAPIKey(Key)と一致しない")
		}
		if !strings.HasPrefix(generated.Key, generated.DisplayPrefix) {
			t.Errorf("DisplayPrefix %q がキーの先頭と一致しない", generated.DisplayPrefix)
		}
		if len(generated.DisplayPrefix) != 12 {
			t.Errorf("DisplayPrefix長: 期待 12, 実際 %d", len(generated.DisplayPrefix))
		}
	})

	t.Run("正常系: 生成ごとに異なるキーが返る", func(t *testing.T) {
		first, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		second, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if first.Key == second.Key {
			t.Error("2回の生成で同じキーが返った")
		}
	})
}

func TestHashAPIKey(t *testing.T) {
	t.Run("正常系: 同じ入力は同じハッシュ、異なる入力は異なるハッシュを返す", func(t *testing.T) {
		hash1 := HashAPIKey("umi_test-key")
		hash2 := HashAPIKey("umi_test-key")
		hash3 := HashAPIKey("umi_other-key")
		if hash1 != hash2 {
			t.Error("同じ入力で異なるハッシュが返った")
		}
		if hash1 == hash3 {
			t.Error("異なる入力で同じハッシュが返った")
		}
		// SHA-256 hex は64文字
		if len(hash1) != 64 {
			t.Errorf("ハッシュ長: 期待 64, 実際 %d", len(hash1))
		}
	})
}

func TestIsAPIKey(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"正常系: umi_プレフィックス付きはtrue", "umi_abc123", true},
		{"正常系: JWTトークンはfalse", "eyJhbGciOiJIUzI1NiJ9.xxx.yyy", false},
		{"正常系: 空文字はfalse", "", false},
		{"正常系: プレフィックスのみはtrue", "umi_", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAPIKey(tt.token); got != tt.want {
				t.Errorf("IsAPIKey(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}
