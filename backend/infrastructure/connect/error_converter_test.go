package connect

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGrpcStatusToConnectError(t *testing.T) {
	tests := []struct {
		name         string
		inputErr     error
		expectedCode connect.Code
		expectNil    bool
	}{
		{
			name:      "正常系: nilを渡すとnilが返る",
			inputErr:  nil,
			expectNil: true,
		},
		{
			name:      "正常系: OK ステータスはnilが返る",
			inputErr:  status.Error(codes.OK, "ok"),
			expectNil: true,
		},
		{
			name:         "正常系: Canceled → CodeCanceled に変換される",
			inputErr:     status.Error(codes.Canceled, "canceled"),
			expectedCode: connect.CodeCanceled,
		},
		{
			name:         "正常系: Unknown → CodeUnknown に変換される",
			inputErr:     status.Error(codes.Unknown, "unknown"),
			expectedCode: connect.CodeUnknown,
		},
		{
			name:         "正常系: InvalidArgument → CodeInvalidArgument に変換される",
			inputErr:     status.Error(codes.InvalidArgument, "invalid argument"),
			expectedCode: connect.CodeInvalidArgument,
		},
		{
			name:         "正常系: DeadlineExceeded → CodeDeadlineExceeded に変換される",
			inputErr:     status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			expectedCode: connect.CodeDeadlineExceeded,
		},
		{
			name:         "正常系: NotFound → CodeNotFound に変換される",
			inputErr:     status.Error(codes.NotFound, "not found"),
			expectedCode: connect.CodeNotFound,
		},
		{
			name:         "正常系: AlreadyExists → CodeAlreadyExists に変換される",
			inputErr:     status.Error(codes.AlreadyExists, "already exists"),
			expectedCode: connect.CodeAlreadyExists,
		},
		{
			name:         "正常系: PermissionDenied → CodePermissionDenied に変換される",
			inputErr:     status.Error(codes.PermissionDenied, "permission denied"),
			expectedCode: connect.CodePermissionDenied,
		},
		{
			name:         "正常系: ResourceExhausted → CodeResourceExhausted に変換される",
			inputErr:     status.Error(codes.ResourceExhausted, "resource exhausted"),
			expectedCode: connect.CodeResourceExhausted,
		},
		{
			name:         "正常系: FailedPrecondition → CodeFailedPrecondition に変換される",
			inputErr:     status.Error(codes.FailedPrecondition, "failed precondition"),
			expectedCode: connect.CodeFailedPrecondition,
		},
		{
			name:         "正常系: Aborted → CodeAborted に変換される",
			inputErr:     status.Error(codes.Aborted, "aborted"),
			expectedCode: connect.CodeAborted,
		},
		{
			name:         "正常系: OutOfRange → CodeOutOfRange に変換される",
			inputErr:     status.Error(codes.OutOfRange, "out of range"),
			expectedCode: connect.CodeOutOfRange,
		},
		{
			name:         "正常系: Unimplemented → CodeUnimplemented に変換される",
			inputErr:     status.Error(codes.Unimplemented, "unimplemented"),
			expectedCode: connect.CodeUnimplemented,
		},
		{
			name:         "正常系: Internal → CodeInternal に変換される",
			inputErr:     status.Error(codes.Internal, "internal"),
			expectedCode: connect.CodeInternal,
		},
		{
			name:         "正常系: Unavailable → CodeUnavailable に変換される",
			inputErr:     status.Error(codes.Unavailable, "unavailable"),
			expectedCode: connect.CodeUnavailable,
		},
		{
			name:         "正常系: DataLoss → CodeDataLoss に変換される",
			inputErr:     status.Error(codes.DataLoss, "data loss"),
			expectedCode: connect.CodeDataLoss,
		},
		{
			name:         "正常系: Unauthenticated → CodeUnauthenticated に変換される",
			inputErr:     status.Error(codes.Unauthenticated, "unauthenticated"),
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:         "異常系: gRPCステータスでないエラーはCodeUnknownに変換される",
			inputErr:     errors.New("non-grpc error"),
			expectedCode: connect.CodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grpcStatusToConnectError(tt.inputErr)

			if tt.expectNil {
				if result != nil {
					t.Errorf("nilを期待したが %v が返った", result)
				}
				return
			}

			if result == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}

			var connectErr *connect.Error
			if !errors.As(result, &connectErr) {
				t.Fatalf("*connect.Error を期待したが %T が返った", result)
			}

			if connectErr.Code() != tt.expectedCode {
				t.Errorf("コード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}
