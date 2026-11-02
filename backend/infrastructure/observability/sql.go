package observability

import (
	"context"
	"database/sql/driver"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// hasParentSpan はリクエストやジョブのスパンの中で呼ばれたかを返す。
// 起動時の接続確認などリクエスト外のクエリで、SQLだけの細切れのトレースが大量にできるのを防ぐ。
func hasParentSpan(ctx context.Context, _ otelsql.Method, _ string, _ []driver.NamedValue) bool {
	return trace.SpanContextFromContext(ctx).IsValid()
}

// SQLOptions は otelsql でDBを開くときのオプションを返す。
// クエリ単位のスパンだけを残し、行の読み出しやセッションのリセットなど件数の多いスパンは省く。
func SQLOptions() []otelsql.Option {
	return []otelsql.Option{
		otelsql.WithAttributes(attribute.String("db.system.name", "postgresql")),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			OmitConnPrepare:      true,
			OmitRows:             true,
			OmitConnectorConnect: true,
			SpanFilter:           hasParentSpan,
		}),
	}
}
