package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/redis/rueidis"
)

// clientRegistrationTTL はDynamic Client Registrationで登録したredirect_urisの保存期間。
// authorization codeのような単回使用の値ではなく、MCPクライアントが繰り返し接続する間
// 有効である必要があるため長めに設定する。期限切れ後は再度 POST /register が必要になる
// （MCPクライアントは接続の都度Discoveryからやり直すため実害はない）。
const clientRegistrationTTL = 30 * 24 * time.Hour

// clientRegistrationKeyPrefix はRedis上でクライアント登録情報を保存する際のキー接頭辞
const clientRegistrationKeyPrefix = "mcp_oauth_client:"

// storeClientRegistration はDynamic Client Registrationで登録されたclient_idと
// redirect_uris の対応関係をRedisにTTL付きで保存する。
// これにより /oauth/authorize・/oauth/consent で「登録時に申告したredirect_uri以外への
// リダイレクトを拒否する」検証（Authorization Code Interception対策）が可能になる。
func storeClientRegistration(ctx context.Context, redisClient rueidis.Client, clientID string, redirectURIs []string) error {
	payload, err := json.Marshal(redirectURIs)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect_uris: %w", err)
	}
	cmd := redisClient.B().Set().Key(clientRegistrationKeyPrefix + clientID).Value(string(payload)).Ex(clientRegistrationTTL).Build()
	if err := redisClient.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("failed to store client registration: %w", err)
	}
	return nil
}

// isRegisteredRedirectURI は、指定されたclient_idの登録時に申告されたredirect_uris
// の中に redirectURI が完全一致で含まれているかを検証する。
// client_idが未登録（期限切れ含む）の場合はfalseを返す。
func isRegisteredRedirectURI(ctx context.Context, redisClient rueidis.Client, clientID, redirectURI string) (bool, error) {
	getCmd := redisClient.B().Get().Key(clientRegistrationKeyPrefix + clientID).Build()
	result := redisClient.Do(ctx, getCmd)
	if result.Error() != nil {
		if rueidis.IsRedisNil(result.Error()) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get client registration: %w", result.Error())
	}
	payload, err := result.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to parse client registration: %w", err)
	}

	var redirectURIs []string
	if err := json.Unmarshal([]byte(payload), &redirectURIs); err != nil {
		return false, fmt.Errorf("failed to unmarshal redirect_uris: %w", err)
	}

	if slices.Contains(redirectURIs, redirectURI) {
		return true, nil
	}
	return false, nil
}
