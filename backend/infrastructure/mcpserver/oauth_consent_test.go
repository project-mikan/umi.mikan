package mcpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewConsentHandler(t *testing.T) {
	t.Run("正常系: 有効なトークンと必須パラメータでauthorization codeを含むリダイレクトURLが返る", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		if err := storeClientRegistration(t.Context(), redisClient, "c1", []string{"https://claude.ai/callback"}); err != nil {
			t.Fatalf("storeClientRegistration失敗: %v", err)
		}
		handler := newConsentHandler(redisClient)
		token := generateValidTokenForTest(t, uuid.New().String())

		body := `{"client_id":"c1","redirect_uri":"https://claude.ai/callback","code_challenge":"abc","code_challenge_method":"S256","state":"xyz"}`
		req := httptest.NewRequest(http.MethodPost, "/oauth/consent", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
		}
		var resp consentResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("レスポンスのJSONパース失敗: %v", err)
		}
		if !strings.HasPrefix(resp.RedirectURL, "https://claude.ai/callback?") {
			t.Errorf("redirect_urlが期待するプレフィックスで始まっていない: %s", resp.RedirectURL)
		}
		if !strings.Contains(resp.RedirectURL, "code=") {
			t.Errorf("redirect_urlにcodeパラメータが含まれていない: %s", resp.RedirectURL)
		}
		if !strings.Contains(resp.RedirectURL, "state=xyz") {
			t.Errorf("redirect_urlにstateパラメータが含まれていない: %s", resp.RedirectURL)
		}
	})

	t.Run("異常系: Authorizationヘッダーがないと401になる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newConsentHandler(redisClient)

		body := `{"client_id":"c1","redirect_uri":"https://claude.ai/callback","code_challenge":"abc","code_challenge_method":"S256"}`
		req := httptest.NewRequest(http.MethodPost, "/oauth/consent", strings.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("異常系: code_challengeがないとinvalid_requestになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newConsentHandler(redisClient)
		token := generateValidTokenForTest(t, uuid.New().String())

		body := `{"client_id":"c1","redirect_uri":"https://claude.ai/callback","code_challenge_method":"S256"}`
		req := httptest.NewRequest(http.MethodPost, "/oauth/consent", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: redirect_uriが不正だとinvalid_requestになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newConsentHandler(redisClient)
		token := generateValidTokenForTest(t, uuid.New().String())

		body := `{"client_id":"c1","redirect_uri":"javascript:alert(1)","code_challenge":"abc","code_challenge_method":"S256"}`
		req := httptest.NewRequest(http.MethodPost, "/oauth/consent", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: client_id登録時のredirect_urisに含まれない値を指定すると、authorization code横取り攻撃を防ぐためinvalid_requestになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		if err := storeClientRegistration(t.Context(), redisClient, "c1", []string{"https://claude.ai/callback"}); err != nil {
			t.Fatalf("storeClientRegistration失敗: %v", err)
		}
		handler := newConsentHandler(redisClient)
		token := generateValidTokenForTest(t, uuid.New().String())

		body := `{"client_id":"c1","redirect_uri":"https://evil.example.com/collect","code_challenge":"abc","code_challenge_method":"S256"}`
		req := httptest.NewRequest(http.MethodPost, "/oauth/consent", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
		}
	})
}
