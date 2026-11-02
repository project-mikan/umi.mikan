package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// StartSpan はリクエストを起点にしない処理（定期ジョブやPub/Subのメッセージ処理）のスパンを開始する。
// 返す関数に処理結果のエラーを渡すと、エラーを記録してからスパンを閉じる。
func StartSpan(ctx context.Context, tracerName, spanName string) (context.Context, func(error)) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, spanName)
	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}
