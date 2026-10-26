package mcpserver

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/project-mikan/umi.mikan/backend/service/user"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

// pkcePairForTokenTest はテスト用にPKCEのverifier/challengeペアを生成する
func pkcePairForTokenTest() (verifier, challenge string) {
	verifier = "test-verifier-0123456789abcdefghijklmn"
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge
}

func TestNewTokenHandler(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)
	userID := testutil.CreateTestUser(t, db, "mcp-oauth-token@example.com", "MCP OAuth User")
	userService := &user.UserEntry{DB: db}

	t.Run("正常系: 正しいcodeとcode_verifierでaccess_tokenが発行される", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)
		verifier, challenge := pkcePairForTokenTest()

		if err := storeAuthCode(t.Context(), redisClient, "valid-code", authCodeData{
			UserID:              userID.String(),
			ClientID:            "client-1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"valid-code"},
			"code_verifier": {verifier},
			"client_id":     {"client-1"},
			"redirect_uri":  {"https://claude.ai/callback"},
		}
		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
		}
		var resp tokenResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("レスポンスのJSONパース失敗: %v", err)
		}
		if !strings.HasPrefix(resp.AccessToken, "umi_") {
			t.Errorf("access_tokenがumi_プレフィックスで始まっていない: %s", resp.AccessToken)
		}
		if resp.TokenType != "Bearer" {
			t.Errorf("token_typeが期待と異なる: got %s", resp.TokenType)
		}
	})

	t.Run("異常系: 同じcodeを2回使うと2回目はinvalid_grantになるので、authorization codeの使い回し（リプレイ攻撃）を防げる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)
		verifier, challenge := pkcePairForTokenTest()

		if err := storeAuthCode(t.Context(), redisClient, "reused-code", authCodeData{
			UserID:              userID.String(),
			ClientID:            "client-1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		makeReq := func() *http.Request {
			form := url.Values{
				"grant_type":    {"authorization_code"},
				"code":          {"reused-code"},
				"code_verifier": {verifier},
				"client_id":     {"client-1"},
				"redirect_uri":  {"https://claude.ai/callback"},
			}
			req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return req
		}

		w1 := httptest.NewRecorder()
		handler(w1, makeReq())
		if w1.Code != http.StatusOK {
			t.Fatalf("1回目のリクエストが失敗: %d, body=%s", w1.Code, w1.Body.String())
		}

		w2 := httptest.NewRecorder()
		handler(w2, makeReq())
		if w2.Code != http.StatusBadRequest {
			t.Errorf("2回目のリクエストのステータスコードが期待と異なる: got %d, want %d", w2.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: code_verifierがchallengeに一致しないとinvalid_grantになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)
		_, challenge := pkcePairForTokenTest()

		if err := storeAuthCode(t.Context(), redisClient, "mismatch-code", authCodeData{
			UserID:              userID.String(),
			ClientID:            "client-1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"mismatch-code"},
			"code_verifier": {"wrong-verifier"},
			"client_id":     {"client-1"},
			"redirect_uri":  {"https://claude.ai/callback"},
		}
		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: 存在しないcodeを指定するとinvalid_grantになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"nonexistent-code"},
			"code_verifier": {"anything"},
			"client_id":     {"client-1"},
			"redirect_uri":  {"https://claude.ai/callback"},
		}
		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: grant_typeがauthorization_code以外だとunsupported_grant_typeになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)

		form := url.Values{"grant_type": {"client_credentials"}}
		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: client_idがstoreされた値と異なるとinvalid_grantになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)
		verifier, challenge := pkcePairForTokenTest()

		if err := storeAuthCode(t.Context(), redisClient, "client-mismatch-code", authCodeData{
			UserID:              userID.String(),
			ClientID:            "client-1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"client-mismatch-code"},
			"code_verifier": {verifier},
			"client_id":     {"different-client"},
			"redirect_uri":  {"https://claude.ai/callback"},
		}
		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: redirect_uriがstoreされた値と異なるとinvalid_grantになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)
		verifier, challenge := pkcePairForTokenTest()

		if err := storeAuthCode(t.Context(), redisClient, "redirect-mismatch-code", authCodeData{
			UserID:              userID.String(),
			ClientID:            "client-1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"redirect-mismatch-code"},
			"code_verifier": {verifier},
			"client_id":     {"client-1"},
			"redirect_uri":  {"https://attacker.example.com/callback"},
		}
		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: client_idを省略するとinvalid_grantになるので、認可コード漏洩時に検証をスキップして横取りされることを防げる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)
		verifier, challenge := pkcePairForTokenTest()

		if err := storeAuthCode(t.Context(), redisClient, "client-omitted-code", authCodeData{
			UserID:              userID.String(),
			ClientID:            "client-1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"client-omitted-code"},
			"code_verifier": {verifier},
			"redirect_uri":  {"https://claude.ai/callback"},
		}
		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: redirect_uriを省略するとinvalid_grantになるので、認可コード漏洩時に検証をスキップして横取りされることを防げる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		handler := newTokenHandler(redisClient, userService)
		verifier, challenge := pkcePairForTokenTest()

		if err := storeAuthCode(t.Context(), redisClient, "redirect-omitted-code", authCodeData{
			UserID:              userID.String(),
			ClientID:            "client-1",
			RedirectURI:         "https://claude.ai/callback",
			CodeChallenge:       challenge,
			CodeChallengeMethod: "S256",
		}); err != nil {
			t.Fatalf("storeAuthCode失敗: %v", err)
		}

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"redirect-omitted-code"},
			"code_verifier": {verifier},
			"client_id":     {"client-1"},
		}
		req := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestNewOAuthApiKeyName(t *testing.T) {
	tests := []struct {
		name        string
		redirectURI string
		now         time.Time
		want        string
	}{
		{
			name:        "正常系: redirect_uriのホストとJSTの日時がキー名に付与される",
			redirectURI: "https://claude.ai/api/mcp/auth_callback",
			now:         time.Date(2026, 7, 21, 12, 34, 0, 0, time.UTC),
			want:        "MCP OAuth (claude.ai) 2026-07-21 21:34",
		},
		{
			name:        "正常系: UTCで日付が変わる時刻でもJSTの日付に変換される",
			redirectURI: "https://agent.meta.ai/api/hatch/oauth/callback",
			now:         time.Date(2026, 7, 21, 15, 30, 0, 0, time.UTC),
			want:        "MCP OAuth (agent.meta.ai) 2026-07-22 00:30",
		},
		{
			name:        "正常系: ポート番号はキー名に含めない",
			redirectURI: "http://localhost:6274/oauth/callback",
			now:         time.Date(2026, 7, 21, 12, 34, 0, 0, time.UTC),
			want:        "MCP OAuth (localhost) 2026-07-21 21:34",
		},
		{
			name:        "異常系: ホストを取り出せないredirect_uriはunknownになる",
			redirectURI: "%%invalid",
			now:         time.Date(2026, 7, 21, 12, 34, 0, 0, time.UTC),
			want:        "MCP OAuth (unknown) 2026-07-21 21:34",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newOAuthApiKeyName(tt.redirectURI, tt.now)
			if got != tt.want {
				t.Errorf("newOAuthApiKeyName() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("正常系: ホスト名が長くてもキー名はuser_api_keys.nameの上限（100文字）に収まるため、トークン発行が失敗しない", func(t *testing.T) {
		longHost := strings.Repeat("a", 200) + ".example.com"
		got := newOAuthApiKeyName("https://"+longHost+"/callback", time.Date(2026, 7, 21, 12, 34, 0, 0, time.UTC))
		if n := len([]rune(got)); n != oauthApiKeyNameMaxLength {
			t.Errorf("キー名の文字数が上限ちょうどになっていない: got %d, want %d (%q)", n, oauthApiKeyNameMaxLength, got)
		}
		if !strings.HasPrefix(got, "MCP OAuth (aaaa") || !strings.HasSuffix(got, ") 2026-07-21 21:34") {
			t.Errorf("切り詰め後のキー名の形式が想定外: %q", got)
		}
	})
}
