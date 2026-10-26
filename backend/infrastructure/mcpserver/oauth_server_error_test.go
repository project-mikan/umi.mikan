package mcpserver

import (
	"bytes"
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/service/user"
	"github.com/redis/rueidis"
)

// setupFailingRedisForTest は全コマンドがエラーを返すminiredisに接続したrueidisクライアントを返す。
// Redis障害時に各OAuthエンドポイントが500を返し、原因をログに出すことを検証するために使う。
func setupFailingRedisForTest(t *testing.T) rueidis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis起動失敗: %v", err)
	}
	t.Cleanup(mr.Close)

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("rueidisクライアント作成失敗: %v", err)
	}
	t.Cleanup(client.Close)
	// クライアント接続後にエラーを注入し、以降のコマンドをすべて失敗させる
	mr.SetError("ERR injected failure")
	return client
}

// captureLog はテスト中に標準logへ出力された内容を取得できるようにする
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

// assertServerErrorLogged は500が返り、原因がログに出力されていることを検証する
func assertServerErrorLogged(t *testing.T, w *httptest.ResponseRecorder, logs *bytes.Buffer, wantLog string) {
	t.Helper()
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("ステータスコードが期待と異なる: got %d, want %d, body=%s", w.Code, http.StatusInternalServerError, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "server_error") {
		t.Errorf("エラーコードがserver_errorではない: %s", w.Body.String())
	}
	if !strings.Contains(logs.String(), wantLog) {
		t.Errorf("原因がログに出力されていない: want %q, got %q", wantLog, logs.String())
	}
}

func TestOAuthEndpoints_ServerError(t *testing.T) {
	t.Run("異常系: Redis障害時に/registerは500を返し、原因をログに出力する", func(t *testing.T) {
		logs := captureLog(t)
		handler := newRegisterHandler(setupFailingRedisForTest(t))

		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"redirect_uris":["https://claude.ai/callback"]}`))
		w := httptest.NewRecorder()
		handler(w, req)

		assertServerErrorLogged(t, w, logs, "mcp oauth register: failed to store client registration")
	})

	t.Run("異常系: Redis障害時に/oauth/authorizeは500を返し、原因をログに出力する", func(t *testing.T) {
		logs := captureLog(t)
		handler := newAuthorizeHandler(setupFailingRedisForTest(t), "https://umi-mikan.example.com")

		q := url.Values{
			"client_id":             {"c1"},
			"redirect_uri":          {"https://claude.ai/callback"},
			"code_challenge":        {"abc"},
			"code_challenge_method": {"S256"},
			"response_type":         {"code"},
		}
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil))

		assertServerErrorLogged(t, w, logs, "mcp oauth authorize: failed to verify redirect_uri")
	})

	t.Run("異常系: FRONTEND_BASE_URLが不正だと/oauth/authorizeは500を返し、原因をログに出力する", func(t *testing.T) {
		logs := captureLog(t)
		redisClient := setupTestRedisForOAuthStoreTest(t)
		if err := storeClientRegistration(t.Context(), redisClient, "c1", []string{"https://claude.ai/callback"}); err != nil {
			t.Fatalf("storeClientRegistration失敗: %v", err)
		}
		handler := newAuthorizeHandler(redisClient, "http://[::1")

		q := url.Values{
			"client_id":             {"c1"},
			"redirect_uri":          {"https://claude.ai/callback"},
			"code_challenge":        {"abc"},
			"code_challenge_method": {"S256"},
			"response_type":         {"code"},
		}
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+q.Encode(), nil))

		assertServerErrorLogged(t, w, logs, "mcp oauth authorize: invalid frontend base URL")
	})

	t.Run("異常系: Redis障害時に/oauth/consentは500を返し、原因をログに出力する", func(t *testing.T) {
		logs := captureLog(t)
		handler := newConsentHandler(setupFailingRedisForTest(t))

		body := `{"client_id":"c1","redirect_uri":"https://claude.ai/callback","code_challenge":"abc","code_challenge_method":"S256"}`
		req := httptest.NewRequest(http.MethodPost, "/oauth/consent", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+generateValidTokenForTest(t, uuid.New().String()))
		w := httptest.NewRecorder()
		handler(w, req)

		assertServerErrorLogged(t, w, logs, "mcp oauth consent: failed to verify redirect_uri")
	})

	t.Run("異常系: Redis障害時に/oauth/tokenは500を返し、原因をログに出力する", func(t *testing.T) {
		logs := captureLog(t)
		handler := newTokenHandler(setupFailingRedisForTest(t), &user.UserEntry{})

		w := httptest.NewRecorder()
		handler(w, newTokenRequestForTest("any-code", "verifier", "c1", "https://claude.ai/callback"))

		assertServerErrorLogged(t, w, logs, "mcp oauth token: failed to look up authorization code")
	})

	t.Run("異常系: authorization codeに保存されたユーザーIDが不正だと/oauth/tokenは500を返し、原因をログに出力する", func(t *testing.T) {
		logs := captureLog(t)
		redisClient := setupTestRedisForOAuthStoreTest(t)
		verifier, challenge := pkcePairForTokenTest()
		if err := storeAuthCode(t.Context(), redisClient, "code-invalid-user", authCodeData{
			UserID:              "not-a-uuid",
			ClientID:            "c1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}
		handler := newTokenHandler(redisClient, &user.UserEntry{})

		w := httptest.NewRecorder()
		handler(w, newTokenRequestForTest("code-invalid-user", verifier, "c1", "https://claude.ai/callback"))

		assertServerErrorLogged(t, w, logs, "mcp oauth token: invalid user id in authorization code")
	})

	t.Run("異常系: APIキーの保存に失敗すると/oauth/tokenは500を返し、原因をログに出力する", func(t *testing.T) {
		logs := captureLog(t)
		redisClient := setupTestRedisForOAuthStoreTest(t)
		verifier, challenge := pkcePairForTokenTest()
		if err := storeAuthCode(t.Context(), redisClient, "code-db-down", authCodeData{
			UserID:              uuid.New().String(),
			ClientID:            "c1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}
		// クローズ済みのDBを渡し、INSERTを確実に失敗させる（DB障害の再現）
		db, err := sql.Open("postgres", "")
		if err != nil {
			t.Fatalf("sql.Open失敗: %v", err)
		}
		_ = db.Close()
		handler := newTokenHandler(redisClient, &user.UserEntry{DB: db})

		w := httptest.NewRecorder()
		handler(w, newTokenRequestForTest("code-db-down", verifier, "c1", "https://claude.ai/callback"))

		assertServerErrorLogged(t, w, logs, "mcp oauth token: failed to issue access token")
	})
}

// newTokenRequestForTest は /oauth/token へのauthorization_codeグラントのリクエストを組み立てる
func newTokenRequestForTest(code, verifier, clientID, redirectURI string) *http.Request {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
	}
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}
