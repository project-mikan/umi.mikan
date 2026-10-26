package mcpserver

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder はレスポンスのステータスコードを記録するためのResponseWriterラッパー
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Flush はMCP Streamable HTTPのSSEレスポンスを逐次送信できるよう、元のWriterのFlushを呼ぶ
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap はhttp.NewResponseController（MCP SDKが内部で使用）が元のWriterに到達できるようにする
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// AccessLogMiddleware はMCPサーバーへの全リクエストのメソッド・パス・ステータス・処理時間をログに出力する。
// MCPサーバーはOAuthフローを含め複数のエンドポイントを持ち、Claude.aiなど外部クライアントが
// どの段階で失敗したかをGrafana（Loki）上で追えるようにするために使う。
// クエリ文字列やボディにはauthorization codeやstateが含まれるため、パスのみを記録する。
func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("mcp access: method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
