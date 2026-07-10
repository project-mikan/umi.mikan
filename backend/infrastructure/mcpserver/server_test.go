package mcpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
