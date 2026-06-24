package connect

import (
	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcStatusToConnectError gRPC status エラーを ConnectRPC エラーに変換する。
// 既存サービスは google.golang.org/grpc/status でエラーを返すため、
// ConnectRPC の HTTP ハンドラーが正しいステータスコードを返せるよう変換が必要。
func grpcStatusToConnectError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return connect.NewError(connect.CodeUnknown, err)
	}

	var code connect.Code
	switch st.Code() {
	case codes.OK:
		return nil
	case codes.Canceled:
		code = connect.CodeCanceled
	case codes.Unknown:
		code = connect.CodeUnknown
	case codes.InvalidArgument:
		code = connect.CodeInvalidArgument
	case codes.DeadlineExceeded:
		code = connect.CodeDeadlineExceeded
	case codes.NotFound:
		code = connect.CodeNotFound
	case codes.AlreadyExists:
		code = connect.CodeAlreadyExists
	case codes.PermissionDenied:
		code = connect.CodePermissionDenied
	case codes.ResourceExhausted:
		code = connect.CodeResourceExhausted
	case codes.FailedPrecondition:
		code = connect.CodeFailedPrecondition
	case codes.Aborted:
		code = connect.CodeAborted
	case codes.OutOfRange:
		code = connect.CodeOutOfRange
	case codes.Unimplemented:
		code = connect.CodeUnimplemented
	case codes.Internal:
		code = connect.CodeInternal
	case codes.Unavailable:
		code = connect.CodeUnavailable
	case codes.DataLoss:
		code = connect.CodeDataLoss
	case codes.Unauthenticated:
		code = connect.CodeUnauthenticated
	default:
		code = connect.CodeUnknown
	}

	return connect.NewError(code, st.Err())
}
