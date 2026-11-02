package observability

import (
	"net/http"
	"net/http/pprof"
)

// RegisterPprofHandlers は Alloy の pyroscope.scrape が取りに来る pprof エンドポイントを mux に登録する。
// メトリクス用のポート（外部に公開していない）にだけ登録すること。
// net/http/pprof を import すると http.DefaultServeMux にも自動で登録されるため、
// http.DefaultServeMux を渡すとパスが重複して panic する。
func RegisterPprofHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
