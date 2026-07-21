package mcpserver

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/redis/rueidis"
)

// consentRequest はフロントエンドの同意画面から送られるリクエストボディ。
// client_id/redirect_uri/code_challenge/code_challenge_method/state は
// /oauth/authorize がフロントエンドへのリダイレクトに埋め込んだ値をそのまま折り返す。
type consentRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	State               string `json:"state"`
}

// consentResponse はauthorization codeを埋め込んだリダイレクト先URLを返す。
// フロントエンドはこのURLに window.location で遷移させることでMCPクライアントに戻る。
type consentResponse struct {
	RedirectURL string `json:"redirect_url"`
}

// newConsentHandler は POST /oauth/consent を提供する。
// フロントエンドがユーザーの同意操作の後、Authorization: Bearer <JWTアクセストークン>
// を添えてこのエンドポイントを呼び出すと、authorization codeを発行してRedisに保存し、
// redirect_uriへの遷移先URLを返す。JWT検証には既存のmodel.ParseAccessTokenをそのまま使う
// （AuthMiddlewareのJWT分岐と同じロジック。APIキーは対象外 — ブラウザ経由の同意フローで
// APIキーを使う想定はない）。
func newConsentHandler(redisClient rueidis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeOAuthError(w, http.StatusMethodNotAllowed, "invalid_request", "method not allowed")
			return
		}

		token, err := model.ExtractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeOAuthError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid bearer token")
			return
		}
		_, userID, err := model.ParseAccessToken(token)
		if err != nil {
			writeOAuthError(w, http.StatusUnauthorized, "unauthorized", "invalid access token")
			return
		}

		var req consentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid JSON body")
			return
		}
		if req.ClientID == "" || req.CodeChallenge == "" {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "client_id and code_challenge are required")
			return
		}
		if req.CodeChallengeMethod != "S256" {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "code_challenge_method must be S256")
			return
		}
		if !isValidRedirectURI(req.RedirectURI) {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri must be an absolute http(s) URL")
			return
		}

		code, err := generateAuthCode()
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to generate authorization code")
			return
		}
		if err := storeAuthCode(r.Context(), redisClient, code, authCodeData{
			UserID:              userID,
			ClientID:            req.ClientID,
			RedirectURI:         req.RedirectURI,
			CodeChallenge:       req.CodeChallenge,
			CodeChallengeMethod: req.CodeChallengeMethod,
		}); err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to store authorization code")
			return
		}

		dest, err := url.Parse(req.RedirectURI)
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "invalid redirect_uri")
			return
		}
		q := dest.Query()
		q.Set("code", code)
		if req.State != "" {
			q.Set("state", req.State)
		}
		dest.RawQuery = q.Encode()

		writeJSON(w, http.StatusOK, consentResponse{RedirectURL: dest.String()})
	}
}
