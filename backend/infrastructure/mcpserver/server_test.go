package mcpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
)

func TestNewServer_RegistersTools(t *testing.T) {
	server := NewServer(&diary.DiaryEntry{})
	if server == nil {
		t.Fatal("サーバーがnilで返された")
	}
}

func TestNewHTTPHandler_RequiresAuth(t *testing.T) {
	handler := NewHTTPHandler(&diary.DiaryEntry{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Post(ts.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("リクエスト送信失敗: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("認証ヘッダーなしのリクエストは401を期待したが %d が返った", resp.StatusCode)
	}
}

func TestNewHTTPHandler_AuthenticatedInitialize(t *testing.T) {
	handler := NewHTTPHandler(&diary.DiaryEntry{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	token := generateValidTokenForTest(t, uuid.New().String())
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, ts.URL, strings.NewReader(body))
	if err != nil {
		t.Fatalf("リクエスト作成失敗: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("リクエスト送信失敗: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("有効なトークンでのinitializeは200を期待したが %d が返った", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("レスポンス読み取り失敗: %v", err)
	}
	if !strings.Contains(string(respBody), serverName) {
		t.Errorf("initializeレスポンスにサーバー名 %q が含まれていない: %s", serverName, string(respBody))
	}
}
