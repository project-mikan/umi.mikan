package mcpserver

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
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
// 悪意あるクライアントがclient_idを取得できても、実際の認可（/oauth/authorize）では
// 本人のログインセッションと同意操作が必須になる。そのためクライアントの真正性検証
// （redirect_uriの事前登録・永続化など）は行わず、client_idをその場で発行するのみの
// 最小実装とする。client_id・redirect_uriの対応関係はサーバー側に保存しないため、
// オープンリダイレクト対策は /oauth/authorize 側でredirect_uriのスキーム・ホストを
// 検証すること（PKCE必須化と合わせ、authorization codeの窃取・再利用を防ぐ）で担保する。
func newRegisterHandler() http.HandlerFunc {
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

		clientID, err := generateClientID()
		if err != nil {
			writeOAuthError(w, http.StatusInternalServerError, "server_error", "failed to generate client_id")
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
