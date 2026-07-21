package mcpserver

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/redis/rueidis"
)

// clientIDRandomBytes はDynamic Client Registrationで発行するclient_idの乱数バイト長
const clientIDRandomBytes = 16

// clientRegistrationRequest はRFC7591 (Dynamic Client Registration) のリクエストのうち
// このサーバーが実際に利用するフィールドのみを受け取る。
type clientRegistrationRequest struct {
	RedirectURIs []string `json:"redirect_uris"`
	ClientName   string   `json:"client_name,omitempty"`
}

// clientRegistrationResponse はRFC7591のレスポンスのうち返却するフィールド。
// client_secret は発行しない（PKCEのみで保護するpublic client）。
type clientRegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name,omitempty"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

// newRegisterHandler はDynamic Client Registration（POST /register）を提供する。
//
// umi.mikanのMCPサーバーは個人利用（自分の日記データへのアクセス）が前提であり、
// client_secretの発行やクライアント名の審査などは行わない最小実装だが、
// redirect_urisはclient_idに紐付けてRedisに保存する（clientRegistrationTTL、
// oauth_client_store.go参照）。これにより /oauth/authorize・/oauth/consent が
// 「登録時に申告したredirect_uri以外への遷移を拒否する」検証を行えるようにし、
// 第三者が任意のclient_idを取得して被害者のauthorization codeを自分の
// redirect_uriへ誘導する攻撃（Authorization Code Interception）を防ぐ。
func newRegisterHandler(redisClient rueidis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeOAuthError(w, http.StatusMethodNotAllowed, "invalid_request", "method not allowed")
			return
		}

		var req clientRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeOAuthError(w, http.StatusBadRequest, "invalid_client_metadata", "invalid JSON body")
			return
		}
		if len(req.RedirectURIs) == 0 {
			writeOAuthError(w, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris is required")
			return
		}
		for _, redirectURI := range req.RedirectURIs {
			if !isValidRedirectURI(redirectURI) {
				writeOAuthError(w, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris must be absolute http(s) URLs")
				return
			}
		}

		clientID, err := generateClientID()
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to generate client_id")
			return
		}
		if err := storeClientRegistration(r.Context(), redisClient, clientID, req.RedirectURIs); err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to store client registration")
			return
		}

		writeJSON(w, http.StatusCreated, clientRegistrationResponse{
			ClientID:                clientID,
			ClientName:              req.ClientName,
			RedirectURIs:            req.RedirectURIs,
			TokenEndpointAuthMethod: "none",
			GrantTypes:              []string{"authorization_code"},
			ResponseTypes:           []string{"code"},
		})
	}
}

// generateClientID は暗号論的乱数からclient_idを生成する
func generateClientID() (string, error) {
	buf := make([]byte, clientIDRandomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "mcpclient_" + hex.EncodeToString(buf), nil
}
