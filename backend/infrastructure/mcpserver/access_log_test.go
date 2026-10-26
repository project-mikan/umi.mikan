package mcpserver

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestAccessLogMiddleware(t *testing.T) {
	t.Run("正常系: メソッド・パス・ステータスがログに出力され、クエリ文字列は出力されない", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		t.Cleanup(func() { log.SetOutput(os.Stderr) })

		handler := AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?state=secret", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("ステータスが想定外: %d", rec.Code)
		}
		out := buf.String()
		if !strings.Contains(out, "method=GET path=/oauth/authorize status=500") {
			t.Errorf("アクセスログが想定外: %q", out)
		}
		if strings.Contains(out, "secret") {
			t.Errorf("クエリ文字列がログに含まれている: %q", out)
		}
	})

	t.Run("正常系: WriteHeaderを呼ばない場合はstatus=200として記録され、Flushが元のWriterに届く", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		t.Cleanup(func() { log.SetOutput(os.Stderr) })

		handler := AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
			if err := http.NewResponseController(w).Flush(); err != nil {
				t.Errorf("Flush失敗: %v", err)
			}
		}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

		if !rec.Flushed {
			t.Error("Flushが元のWriterに届いていない")
		}
		if !strings.Contains(buf.String(), "method=POST path=/ status=200") {
			t.Errorf("アクセスログが想定外: %q", buf.String())
		}
	})
}
