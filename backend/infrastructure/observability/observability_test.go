package observability

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/XSAM/otelsql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestIsTracingEnabled(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{name: "正常系: どちらの環境変数も未設定なら無効", env: map[string]string{}, want: false},
		{name: "正常系: OTEL_EXPORTER_OTLP_ENDPOINTがあれば有効", env: map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://tempo:4318"}, want: true},
		{name: "正常系: トレース専用のOTEL_EXPORTER_OTLP_TRACES_ENDPOINTだけでも有効", env: map[string]string{"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://tempo:4318/v1/traces"}, want: true},
		{name: "正常系: 空文字は未設定と同じ扱いで無効", env: map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": ""}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string { return tt.env[key] }
			assert.Equal(t, tt.want, isTracingEnabled(getenv))
		})
	}
}

func TestSetupTracing(t *testing.T) {
	tests := []struct {
		name            string
		env             map[string]string
		wantSDKProvider bool
	}{
		{
			name:            "正常系: 送信先が未設定ならグローバルのトレーサーを差し替えない",
			env:             map[string]string{},
			wantSDKProvider: false,
		},
		{
			name:            "正常系: 送信先が設定されていればSDKのトレーサーを登録する",
			env:             map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318"},
			wantSDKProvider: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 他のテストに影響しないよう、グローバルのトレーサーを元に戻す
			original := otel.GetTracerProvider()
			t.Cleanup(func() { otel.SetTracerProvider(original) })
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			shutdown, err := setupTracing(context.Background(), "test-service", func(key string) string { return tt.env[key] })
			require.NoError(t, err)
			require.NotNil(t, shutdown)

			_, isSDK := otel.GetTracerProvider().(*sdktrace.TracerProvider)
			assert.Equal(t, tt.wantSDKProvider, isSDK)
			assert.NoError(t, shutdown(context.Background()))
		})
	}
}

func TestHasParentSpan(t *testing.T) {
	// 中身のあるスパンコンテキストを持つ ctx を作る
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1},
		SpanID:  trace.SpanID{1},
	})

	tests := []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{name: "正常系: 親スパンがなければスパンを作らない", ctx: context.Background(), want: false},
		{name: "正常系: 親スパンがあればスパンを作る", ctx: trace.ContextWithSpanContext(context.Background(), spanCtx), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasParentSpan(tt.ctx, otelsql.MethodConnQuery, "SELECT 1", nil))
		})
	}
}

func TestSQLOptions(t *testing.T) {
	assert.NotEmpty(t, SQLOptions())
}

func TestStartSpan(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus codes.Code
		wantEvents int
	}{
		{name: "正常系: エラーがなければステータスは未設定のまま閉じる", err: nil, wantStatus: codes.Unset, wantEvents: 0},
		{name: "正常系: エラーがあればエラーとして記録して閉じる", err: errors.New("失敗"), wantStatus: codes.Error, wantEvents: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := tracetest.NewSpanRecorder()
			original := otel.GetTracerProvider()
			otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
			t.Cleanup(func() { otel.SetTracerProvider(original) })

			ctx, end := StartSpan(context.Background(), "test", "test-span")
			assert.True(t, trace.SpanContextFromContext(ctx).IsValid())
			end(tt.err)

			spans := recorder.Ended()
			require.Len(t, spans, 1)
			assert.Equal(t, "test-span", spans[0].Name())
			assert.Equal(t, tt.wantStatus, spans[0].Status().Code)
			assert.Len(t, spans[0].Events(), tt.wantEvents)
		})
	}
}

func TestRegisterPprofHandlers(t *testing.T) {
	mux := http.NewServeMux()
	RegisterPprofHandlers(mux)

	tests := []struct {
		name string
		path string
	}{
		{name: "正常系: 一覧ページを返す", path: "/debug/pprof/"},
		{name: "正常系: ヒーププロファイルを返す", path: "/debug/pprof/heap"},
		{name: "正常系: goroutineプロファイルを返す", path: "/debug/pprof/goroutine"},
		{name: "正常系: コマンドラインを返す", path: "/debug/pprof/cmdline"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
