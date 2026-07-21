package mcpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProtectedResourceMetadataHandler(t *testing.T) {
	t.Run("正常系: resourceとauthorization_serversがbaseURLから組み立てられる", func(t *testing.T) {
		handler := newProtectedResourceMetadataHandler("https://umi-mikan-api.usuyuki.net")
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusOK)
		}
		var resp protectedResourceMetadata
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("レスポンスのJSONパース失敗: %v", err)
		}
		if len(resp.AuthorizationServers) != 1 || resp.AuthorizationServers[0] != "https://umi-mikan-api.usuyuki.net" {
			t.Errorf("authorization_serversが期待と異なる: %v", resp.AuthorizationServers)
		}
	})
}

func TestNewAuthorizationServerMetadataHandler(t *testing.T) {
	t.Run("正常系: 各エンドポイントURLがbaseURLから組み立てられる", func(t *testing.T) {
		handler := newAuthorizationServerMetadataHandler("https://umi-mikan-api.usuyuki.net")
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusOK)
		}
		var resp authorizationServerMetadata
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("レスポンスのJSONパース失敗: %v", err)
		}
		if resp.AuthorizationEndpoint != "https://umi-mikan-api.usuyuki.net"+oauthAuthorizePath {
			t.Errorf("authorization_endpointが期待と異なる: %s", resp.AuthorizationEndpoint)
		}
		if resp.TokenEndpoint != "https://umi-mikan-api.usuyuki.net"+oauthTokenPath {
			t.Errorf("token_endpointが期待と異なる: %s", resp.TokenEndpoint)
		}
		if resp.RegistrationEndpoint != "https://umi-mikan-api.usuyuki.net"+oauthRegisterPath {
			t.Errorf("registration_endpointが期待と異なる: %s", resp.RegistrationEndpoint)
		}
		if len(resp.CodeChallengeMethodsSupported) != 1 || resp.CodeChallengeMethodsSupported[0] != "S256" {
			t.Errorf("code_challenge_methods_supportedが期待と異なる: %v", resp.CodeChallengeMethodsSupported)
		}
	})
}
