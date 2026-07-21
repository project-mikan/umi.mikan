package mcpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/service/user"
	"github.com/redis/rueidis"
)

// oauthApiKeyNamePrefix はOAuthフロー経由で発行されるAPIキーの名前接頭辞。
// user_api_keys.name に保存され、設定ページのAPIキー一覧にも他のキーと並んで表示される
// （どのキーがどのMCPクライアント接続由来か利用者が判別できるようにするため）。
// 発行日時（JST・分単位）を付与するのは、同じユーザーが複数回接続・再認可した際に
// user_api_keys.name にUNIQUE制約が無く同名レコードが複数並んでしまい、
// 設定ページでどれを失効すべきか判別できなくなるのを防ぐため。
const oauthApiKeyNamePrefix = "MCP OAuth (Claude connector)"

// newOAuthApiKeyName は現在時刻（JST）を付与したOAuth用APIキー名を生成する。
func newOAuthApiKeyName(now time.Time) string {
	jst := time.FixedZone("JST", 9*60*60)
	return fmt.Sprintf("%s %s", oauthApiKeyNamePrefix, now.In(jst).Format("2006-01-02 15:04"))
}

// tokenResponse はRFC6749 5.1節 (Successful Response) の必要最小限のフィールド。
// token_typeはBearer固定。refresh_tokenは発行しない
// （既存のAPIキーは90日有効・DeleteApiKeyで即時失効できるため、リフレッシュの仕組みを
// 別途実装するメリットが薄い。期限切れ後はMCPクライアント側で再度/authorizeからやり直す）。
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// newTokenHandler は POST /oauth/token を提供する。
// authorization_code グラントのみをサポートする。codeをRedisから取得しPKCE検証したうえで、
// 既存のAPIキー発行ロジック（user.UserEntry.CreateApiKeyForUser）を呼び出し、
// 発行されたAPIキー本体をOAuthの access_token としてそのまま返す。これにより
// MCPサーバー側の認証ミドルウェア（auth.go AuthMiddleware）は無改造のまま、
// OAuth経由で取得したトークンも既存のAPIキー authentication パスをそのまま通過できる。
func newTokenHandler(redisClient rueidis.Client, userService *user.UserEntry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeOAuthError(w, http.StatusMethodNotAllowed, "invalid_request", "method not allowed")
			return
		}
		if err := r.ParseForm(); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid form body")
			return
		}

		if r.PostForm.Get("grant_type") != "authorization_code" {
			writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "only authorization_code is supported")
			return
		}
		code := r.PostForm.Get("code")
		codeVerifier := r.PostForm.Get("code_verifier")
		if code == "" || codeVerifier == "" {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "code and code_verifier are required")
			return
		}

		data, ok, err := consumeAuthCode(r.Context(), redisClient, code)
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to look up authorization code")
			return
		}
		if !ok {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "authorization code is invalid or expired")
			return
		}

		// client_id・redirect_uriの一致確認。この実装では/authorize・/consent双方で
		// 常にclient_id/redirect_uriを必須としているため、RFC6749 4.1.3節が求める
		// 「/authorize時に指定された値との一致確認」は値の省略を許さず常に行う
		// （省略時に検証をスキップすると、redirect_uri不一致検出という多層防御が失われるため）。
		if r.PostForm.Get("client_id") != data.ClientID {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "client_id mismatch")
			return
		}
		if r.PostForm.Get("redirect_uri") != data.RedirectURI {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
			return
		}

		if !verifyPKCE(data.CodeChallenge, data.CodeChallengeMethod, codeVerifier) {
			writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
			return
		}

		userID, err := uuid.Parse(data.UserID)
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "invalid user id")
			return
		}

		key, plainKey, err := userService.CreateApiKeyForUser(r.Context(), userID, newOAuthApiKeyName(time.Now()))
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to issue access token")
			return
		}

		writeJSON(w, http.StatusOK, tokenResponse{
			AccessToken: plainKey,
			TokenType:   "Bearer",
			ExpiresIn:   key.ExpiresAt - key.CreatedAt,
		})
	}
}
