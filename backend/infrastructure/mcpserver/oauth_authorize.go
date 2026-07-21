package mcpserver

import (
	"net/http"
	"net/url"

	"github.com/redis/rueidis"
)

// newAuthorizeHandler は GET /oauth/authorize を提供する。
// MCPクライアント（ブラウザ経由）からのAuthorization Requestを受け取り、
// パラメータを検証したうえでフロントエンド（SvelteKit）のログイン/同意画面
// （/oauth/authorize）にそのままリダイレクトする。実際のログイン確認・同意取得・
// authorization code発行はフロントエンドが POST /oauth/consent（このパッケージの
// oauth_consent.go）を叩くことで行う。
//
// frontendBaseURL はフロントエンドの公開URL（例: https://umi-mikan.usuyuki.net）。
func newAuthorizeHandler(redisClient rueidis.Client, frontendBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		clientID := query.Get("client_id")
		redirectURI := query.Get("redirect_uri")
		codeChallenge := query.Get("code_challenge")
		codeChallengeMethod := query.Get("code_challenge_method")
		state := query.Get("state")
		responseType := query.Get("response_type")

		if clientID == "" || redirectURI == "" || codeChallenge == "" {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "client_id, redirect_uri, code_challenge are required")
			return
		}
		if responseType != "code" {
			writeOAuthError(w, http.StatusBadRequest, "unsupported_response_type", "only response_type=code is supported")
			return
		}
		if codeChallengeMethod != "S256" {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "code_challenge_method must be S256")
			return
		}
		if !isValidRedirectURI(redirectURI) {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri must be an absolute http(s) URL")
			return
		}
		// client_id登録時に申告されたredirect_uris以外への遷移を拒否する。
		// これがないと、第三者が自分のclient_idを取得したうえで任意のredirect_uriを
		// 指定し、被害者のauthorization codeを自分のサーバーに誘導できてしまう
		// （Authorization Code Interception。PKCEは攻撃者が自分でcode_challenge/
		// code_verifierを用意するこのシナリオを防げない）。
		registered, err := isRegisteredRedirectURI(r.Context(), redisClient, clientID, redirectURI)
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to verify redirect_uri")
			return
		}
		if !registered {
			writeOAuthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri is not registered for this client_id")
			return
		}

		dest, err := url.Parse(frontendBaseURL + frontendConsentPath)
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "invalid frontend base URL")
			return
		}
		q := dest.Query()
		q.Set("client_id", clientID)
		q.Set("redirect_uri", redirectURI)
		q.Set("code_challenge", codeChallenge)
		q.Set("code_challenge_method", codeChallengeMethod)
		q.Set("state", state)
		dest.RawQuery = q.Encode()

		http.Redirect(w, r, dest.String(), http.StatusFound)
	}
}

// isValidRedirectURI はredirect_uriが絶対URLかつhttp(s)スキームであることを検証する。
// オープンリダイレクト対策として、javascript:等の危険なスキームを排除する。
// クライアントの事前登録は行わない設計（oauth_register.go参照）のため、ホスト名までは
// 制限せずスキームのみ検証する。
func isValidRedirectURI(redirectURI string) bool {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}
