package mcpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRegisterHandler(t *testing.T) {
	t.Run("正常系: redirect_urisを指定するとclient_idが発行される", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newRegisterHandler(redisClient)
		body := `{"redirect_uris":["https://claude.ai/api/mcp/callback"],"client_name":"Claude"}`
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d, body=%s", w.Code, http.StatusCreated, w.Body.String())
		}
		var resp clientRegistrationResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("レスポンスのJSONパース失敗: %v", err)
		}
		if resp.ClientID == "" {
			t.Error("client_idが空文字だった")
		}
		if resp.TokenEndpointAuthMethod != "none" {
			t.Errorf("token_endpoint_auth_methodが期待と異なる: got %s", resp.TokenEndpointAuthMethod)
		}

		registered, err := isRegisteredRedirectURI(t.Context(), redisClient, resp.ClientID, "https://claude.ai/api/mcp/callback")
		if err != nil {
			t.Fatalf("isRegisteredRedirectURI失敗: %v", err)
		}
		if !registered {
			t.Error("登録したredirect_uriがRedisに保存されていない")
		}
	})

	t.Run("異常系: redirect_urisを指定しないとinvalid_redirect_uriエラーになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newRegisterHandler(redisClient)
		body := `{"client_name":"Claude"}`
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: 不正なJSONを送るとinvalid_client_metadataエラーになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newRegisterHandler(redisClient)
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("{invalid"))
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: GETメソッドで呼ぶとmethod not allowedになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newRegisterHandler(redisClient)
		req := httptest.NewRequest(http.MethodGet, "/register", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("異常系: redirect_urisに不正なスキームが含まれるとinvalid_redirect_uriエラーになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newRegisterHandler(redisClient)
		body := `{"redirect_uris":["javascript:alert(1)"],"client_name":"Claude"}`
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
