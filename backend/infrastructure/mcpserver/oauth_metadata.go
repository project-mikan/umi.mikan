package mcpserver

import (
	"encoding/json"
	"net/http"
)

// oauthMetadataHandlers は、MCP仕様のAuthorization
// （https://spec.modelcontextprotocol.io/specification/basic/authorization/）が要求する
// Discoveryエンドポイントを提供する。Claude.aiなどのMCPクライアントは接続時にまず
// /.well-known/oauth-protected-resource → /.well-known/oauth-authorization-server の順に
// アクセスして、認可・トークンエンドポイントのURLを解決する。

// protectedResourceMetadata はRFC9728 (OAuth 2.0 Protected Resource Metadata) の必要最小限の応答。
type protectedResourceMetadata struct {
	Resource             string   `json:"resource"`
	AuthorizationServers []string `json:"authorization_servers"`
}

// authorizationServerMetadata はRFC8414 (OAuth 2.0 Authorization Server Metadata) の必要最小限の応答。
type authorizationServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RegistrationEndpoint              string   `json:"registration_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

// newProtectedResourceMetadataHandler は /.well-known/oauth-protected-resource を提供する。
// baseURL はこのMCPサーバー自身から見た公開URL（例: https://umi-mikan-api.usuyuki.net）。
func newProtectedResourceMetadataHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, protectedResourceMetadata{
			Resource:             baseURL + mcpPath,
			AuthorizationServers: []string{baseURL},
		})
	}
}

// newAuthorizationServerMetadataHandler は /.well-known/oauth-authorization-server を提供する。
func newAuthorizationServerMetadataHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, authorizationServerMetadata{
			Issuer:                            baseURL,
			AuthorizationEndpoint:             baseURL + oauthAuthorizePath,
			TokenEndpoint:                     baseURL + oauthTokenPath,
			RegistrationEndpoint:              baseURL + oauthRegisterPath,
			ResponseTypesSupported:            []string{"code"},
			GrantTypesSupported:               []string{"authorization_code"},
			CodeChallengeMethodsSupported:     []string{"S256"},
			TokenEndpointAuthMethodsSupported: []string{"none"},
		})
	}
}

// writeJSON はJSONレスポンスを書き込む共通ヘルパー
func writeJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

// writeOAuthError はOAuth 2.0エラーレスポンス（RFC6749 5.2節）形式でエラーを返す
func writeOAuthError(w http.ResponseWriter, statusCode int, errCode, description string) {
	writeJSON(w, statusCode, map[string]string{
		"error":             errCode,
		"error_description": description,
	})
}
