package mcpserver

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/rueidis"
)

// authCodeTTL はauthorization codeの有効期間。
// OAuth 2.0のベストプラクティス（RFC6749 4.1.2）に従い短命にする。
const authCodeTTL = 5 * time.Minute

// authCodeRandomBytes はauthorization codeの乱数バイト長
const authCodeRandomBytes = 32

// authCodeKeyPrefix はRedis上でauthorization codeを保存する際のキー接頭辞
const authCodeKeyPrefix = "mcp_oauth_code:"

// authCodeData はauthorization codeに紐づけてRedisに保存する情報。
// PKCEのcode_challengeを保存しておき、/oauth/token側でcode_verifierと突き合わせる。
type authCodeData struct {
	UserID              string `json:"user_id"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

// generateAuthCode は暗号論的乱数からauthorization codeを生成する
func generateAuthCode() (string, error) {
	buf := make([]byte, authCodeRandomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate auth code: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// storeAuthCode はauthorization codeとその付随情報をRedisにTTL付きで保存する
func storeAuthCode(ctx context.Context, redisClient rueidis.Client, code string, data authCodeData) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal auth code data: %w", err)
	}
	cmd := redisClient.B().Set().Key(authCodeKeyPrefix + code).Value(string(payload)).Ex(authCodeTTL).Build()
	if err := redisClient.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("failed to store auth code: %w", err)
	}
	return nil
}

// consumeAuthCode はauthorization codeをRedisから取得と同時に削除する（単回使用の強制）。
// GETDELはRedis側でアトミックに実行されるため、同一codeに対する複数リクエストが
// 同時に到達しても、取得できるのはそのうち1件のみとなる（GET+DELの2回のコマンドに
// 分けると、削除が完了する前に複数リクエストがGETを通過してしまい、同じcodeから
// 複数のAPIキーが発行されうる）。
// 存在しない・期限切れの場合はokがfalseになる。
func consumeAuthCode(ctx context.Context, redisClient rueidis.Client, code string) (authCodeData, bool, error) {
	key := authCodeKeyPrefix + code

	cmd := redisClient.B().Getdel().Key(key).Build()
	result := redisClient.Do(ctx, cmd)
	if result.Error() != nil {
		if rueidis.IsRedisNil(result.Error()) {
			return authCodeData{}, false, nil
		}
		return authCodeData{}, false, fmt.Errorf("failed to get and delete auth code: %w", result.Error())
	}
	payload, err := result.ToString()
	if err != nil {
		return authCodeData{}, false, fmt.Errorf("failed to parse auth code: %w", err)
	}

	var data authCodeData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return authCodeData{}, false, fmt.Errorf("failed to unmarshal auth code data: %w", err)
	}
	return data, true, nil
}

// verifyPKCE はPKCE (RFC7636) のcode_verifierがcode_challengeと一致するか検証する。
// S256のみサポートする（plain方式は許可しない）。
func verifyPKCE(codeChallenge, codeChallengeMethod, codeVerifier string) bool {
	if codeChallengeMethod != "S256" {
		return false
	}
	sum := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return computed == codeChallenge
}
