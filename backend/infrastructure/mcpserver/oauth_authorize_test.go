package mcpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsValidRedirectURI(t *testing.T) {
	tests := []struct {
		name        string
		redirectURI string
		want        bool
	}{
		{
			name:        "正常系: httpsの絶対URLはtrueになる",
			redirectURI: "https://claude.ai/api/mcp/callback",
			want:        true,
		},
		{
			name:        "正常系: httpの絶対URLはtrueになる（開発環境向け）",
			redirectURI: "http://localhost:3000/callback",
			want:        true,
		},
		{
			name:        "異常系: javascriptスキームだとオープンリダイレクトの危険があるためfalseになる",
			redirectURI: "javascript:alert(1)",
			want:        false,
		},
		{
			name:        "異常系: 相対パスだとホストがないためfalseになる",
			redirectURI: "/callback",
			want:        false,
		},
		{
			name:        "異常系: 空文字だとfalseになる",
			redirectURI: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRedirectURI(tt.redirectURI)
			if got != tt.want {
				t.Errorf("isValidRedirectURI(%q) = %v, want %v", tt.redirectURI, got, tt.want)
			}
		})
	}
}

func TestNewAuthorizeHandler(t *testing.T) {
	t.Run("正常系: 必須パラメータが揃っているとフロントエンドの同意画面へ302リダイレクトする", func(t *testing.T) {
		handler := newAuthorizeHandler("http://localhost:2000")
		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?client_id=c1&redirect_uri=https://claude.ai/callback&code_challenge=abc&code_challenge_method=S256&response_type=code&state=xyz", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusFound)
		}
		location := w.Header().Get("Location")
		if location == "" {
			t.Fatal("Locationヘッダーが空だった")
		}
	})

	t.Run("異常系: client_idがないとinvalid_requestエラーになる", func(t *testing.T) {
		handler := newAuthorizeHandler("http://localhost:2000")
		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?redirect_uri=https://claude.ai/callback&code_challenge=abc&code_challenge_method=S256&response_type=code", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: response_typeがcode以外だとunsupported_response_typeエラーになる", func(t *testing.T) {
		handler := newAuthorizeHandler("http://localhost:2000")
		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?client_id=c1&redirect_uri=https://claude.ai/callback&code_challenge=abc&code_challenge_method=S256&response_type=token", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: code_challenge_methodがS256以外だと拒否される（plain方式は脆弱なため非対応）", func(t *testing.T) {
		handler := newAuthorizeHandler("http://localhost:2000")
		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?client_id=c1&redirect_uri=https://claude.ai/callback&code_challenge=abc&code_challenge_method=plain&response_type=code", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("異常系: redirect_uriが不正なスキームだとinvalid_requestエラーになる", func(t *testing.T) {
		handler := newAuthorizeHandler("http://localhost:2000")
		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?client_id=c1&redirect_uri=javascript:alert(1)&code_challenge=abc&code_challenge_method=S256&response_type=code", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("ステータスコードが期待と異なる: got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
