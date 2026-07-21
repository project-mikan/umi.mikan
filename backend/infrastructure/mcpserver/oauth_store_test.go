package mcpserver

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/rueidis"
)

// setupTestRedisForOAuthStoreTest はテスト用のminiredisクライアントを起動してrueidisクライアントを返す
func setupTestRedisForOAuthStoreTest(t *testing.T) rueidis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis起動失敗: %v", err)
	}
	t.Cleanup(mr.Close)

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("rueidisクライアント作成失敗: %v", err)
	}
	t.Cleanup(client.Close)
	return client
}

func TestGenerateAuthCode(t *testing.T) {
	t.Run("正常系: 生成される度に異なるcodeが返る", func(t *testing.T) {
		code1, err := generateAuthCode()
		if err != nil {
			t.Fatalf("generateAuthCode失敗: %v", err)
		}
		code2, err := generateAuthCode()
		if err != nil {
			t.Fatalf("generateAuthCode失敗: %v", err)
		}
		if code1 == code2 {
			t.Errorf("2回の生成で同じcodeが返った: %s", code1)
		}
		if len(code1) != authCodeRandomBytes*2 {
			t.Errorf("codeの長さが期待と異なる: got %d, want %d", len(code1), authCodeRandomBytes*2)
		}
	})
}

func TestStoreAndConsumeAuthCode(t *testing.T) {
	t.Run("正常系: 保存したauth codeを取得できる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		data := authCodeData{
			UserID:              "user-1",
			ClientID:            "client-1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       "challenge",
			CodeChallengeMethod: "S256",
		}
		if err := storeAuthCode(t.Context(), redisClient, "code-1", data); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		got, ok, err := consumeAuthCode(t.Context(), redisClient, "code-1")
		if err != nil {
			t.Fatalf("consumeAuthCode失敗: %v", err)
		}
		if !ok {
			t.Fatal("codeが見つからなかった")
		}
		if got != data {
			t.Errorf("取得したデータが一致しない: got %+v, want %+v", got, data)
		}
	})

	t.Run("異常系: 一度consumeしたcodeは再利用するとokがfalseになるので、二重使用（トークン窃取時のリプレイ攻撃）を防げる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		data := authCodeData{UserID: "user-1", ClientID: "client-1", RedirectURI: "https://claude.ai/callback", CodeChallenge: "challenge", CodeChallengeMethod: "S256"}
		if err := storeAuthCode(t.Context(), redisClient, "code-1", data); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		if _, ok, err := consumeAuthCode(t.Context(), redisClient, "code-1"); err != nil || !ok {
			t.Fatalf("1回目のconsumeに失敗: ok=%v, err=%v", ok, err)
		}

		_, ok, err := consumeAuthCode(t.Context(), redisClient, "code-1")
		if err != nil {
			t.Fatalf("2回目のconsumeでエラー: %v", err)
		}
		if ok {
			t.Error("2回目のconsumeでokがtrueになった（単回使用が保証されていない）")
		}
	})

	t.Run("異常系: 存在しないcodeを指定するとokがfalseになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		_, ok, err := consumeAuthCode(t.Context(), redisClient, "nonexistent-code")
		if err != nil {
			t.Fatalf("consumeAuthCodeでエラー: %v", err)
		}
		if ok {
			t.Error("存在しないcodeに対してokがtrueになった")
		}
	})
}

func TestVerifyPKCE(t *testing.T) {
	verifier := "test-code-verifier-1234567890"
	sum := sha256.Sum256([]byte(verifier))
	validChallenge := base64.RawURLEncoding.EncodeToString(sum[:])

	tests := []struct {
		name                string
		codeChallenge       string
		codeChallengeMethod string
		codeVerifier        string
		want                bool
	}{
		{
			name:                "正常系: S256で正しいverifierを渡すとtrueになる",
			codeChallenge:       validChallenge,
			codeChallengeMethod: "S256",
			codeVerifier:        verifier,
			want:                true,
		},
		{
			name:                "異常系: verifierがchallengeに一致しないのでfalseになる",
			codeChallenge:       validChallenge,
			codeChallengeMethod: "S256",
			codeVerifier:        "wrong-verifier",
			want:                false,
		},
		{
			name:                "異常系: codeChallengeMethodがplainだとS256未対応のためfalseになる",
			codeChallenge:       validChallenge,
			codeChallengeMethod: "plain",
			codeVerifier:        verifier,
			want:                false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := verifyPKCE(tt.codeChallenge, tt.codeChallengeMethod, tt.codeVerifier)
			if got != tt.want {
				t.Errorf("verifyPKCE() = %v, want %v", got, tt.want)
			}
		})
	}
}
