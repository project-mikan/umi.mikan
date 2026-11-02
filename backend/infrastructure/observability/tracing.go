// Package observability は、Grafana の Traces / Profiles Drilldown で見るための
// トレース送信（OpenTelemetry → Tempo）と pprof エンドポイントの公開をまとめる。
package observability

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// otlpEndpointEnvKeys は送信先として参照する環境変数（OpenTelemetry SDK の標準）
var otlpEndpointEnvKeys = []string{
	"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
	"OTEL_EXPORTER_OTLP_ENDPOINT",
}

// ShutdownFunc は送信待ちのスパンを送り切ってからトレーサーを止める関数
type ShutdownFunc func(context.Context) error

// isTracingEnabled はOTLPの送信先が環境変数で指定されているかを返す
func isTracingEnabled(getenv func(string) string) bool {
	for _, key := range otlpEndpointEnvKeys {
		if getenv(key) != "" {
			return true
		}
	}
	return false
}

// SetupTracing はOTLP（HTTP）でトレースを送るトレーサーをグローバルに登録する。
// 送信先の環境変数が未設定ならトレースは無効のままにし、何もしない終了関数を返す。
// serviceName は OTEL_SERVICE_NAME が設定されていればそちらで上書きされる。
func SetupTracing(ctx context.Context, serviceName string) (ShutdownFunc, error) {
	return setupTracing(ctx, serviceName, os.Getenv)
}

// setupTracing は環境変数の読み取りを差し替えられる SetupTracing の本体
func setupTracing(ctx context.Context, serviceName string, getenv func(string) string) (ShutdownFunc, error) {
	if !isTracingEnabled(getenv) {
		return func(context.Context) error { return nil }, nil
	}

	// 送信先・プロトコルなどは OTEL_EXPORTER_OTLP_* の環境変数から読み取られる
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// 後に指定したものほど優先されるため、環境変数（OTEL_SERVICE_NAME など）を最後に置く
	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", serviceName)),
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return provider.Shutdown, nil
}
