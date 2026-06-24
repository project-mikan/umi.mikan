package connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/service/auth"
	"github.com/project-mikan/umi.mikan/backend/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupAuthAdapter(t *testing.T) *AuthServiceAdapter {
	t.Helper()
	db := testutil.SetupTestDB(t)
	svc := &auth.AuthEntry{DB: db}
	return &AuthServiceAdapter{svc: svc}
}

func TestAuthServiceAdapter_GetRegistrationConfig(t *testing.T) {
	tests := []struct {
		name             string
		registerKey      string
		expectedRequired bool
	}{
		{
			name:             "正常系: REGISTER_KEYが未設定の場合はregister_key_requiredがfalse",
			registerKey:      "",
			expectedRequired: false,
		},
		{
			name:             "正常系: REGISTER_KEYが設定されている場合はregister_key_requiredがtrue",
			registerKey:      "secret-key",
			expectedRequired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			svc := &auth.AuthEntry{DB: db, RegisterKey: tt.registerKey}
			adapter := &AuthServiceAdapter{svc: svc}

			req := connect.NewRequest(&g.GetRegistrationConfigRequest{})
			resp, err := adapter.GetRegistrationConfig(context.Background(), req)

			if err != nil {
				t.Fatalf("エラーを期待しなかったが %v が返った", err)
			}
			if resp.Msg.GetRegisterKeyRequired() != tt.expectedRequired {
				t.Errorf("register_key_required: 期待 %v, 実際 %v", tt.expectedRequired, resp.Msg.GetRegisterKeyRequired())
			}
		})
	}
}

func TestAuthServiceAdapter_RegisterByPassword_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.RegisterByPasswordRequest
		expectedCode connect.Code
	}{
		{
			name: "異常系: 空のメールアドレスはInvalidArgumentエラーになる",
			request: &g.RegisterByPasswordRequest{
				Email:    "",
				Password: "password123",
				Name:     "Test User",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
		{
			name: "異常系: 空のパスワードはInvalidArgumentエラーになる",
			request: &g.RegisterByPasswordRequest{
				Email:    "test@example.com",
				Password: "",
				Name:     "Test User",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupAuthAdapter(t)
			req := connect.NewRequest(tt.request)
			_, err := adapter.RegisterByPassword(context.Background(), req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestAuthServiceAdapter_LoginByPassword_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.LoginByPasswordRequest
		expectedCode connect.Code
	}{
		{
			name: "異常系: 存在しないユーザーのログインはUnauthenticatedエラーになる",
			request: &g.LoginByPasswordRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedCode: connect.CodeUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupAuthAdapter(t)
			req := connect.NewRequest(tt.request)
			_, err := adapter.LoginByPassword(context.Background(), req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestAuthServiceAdapter_RefreshAccessToken_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.RefreshAccessTokenRequest
		expectedCode connect.Code
	}{
		{
			name:         "異常系: 空のリフレッシュトークンはInvalidArgumentエラーになる",
			request:      &g.RefreshAccessTokenRequest{},
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupAuthAdapter(t)
			req := connect.NewRequest(tt.request)

			// RefreshAccessToken はメタデータからトークンを取得するためコンテキストのみで呼ぶ
			_, err := adapter.RefreshAccessToken(context.Background(), req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

// grpcStatusToConnectError の委譲が正しく動作することを確認する統合テスト
func TestAuthServiceAdapter_ErrorConversion(t *testing.T) {
	// gRPC status エラーが Connect エラーに変換されることを確認
	grpcErr := status.Error(codes.NotFound, "user not found")
	connectErr := grpcStatusToConnectError(grpcErr)

	if connectErr == nil {
		t.Fatal("エラーを期待したがnilが返った")
	}
	ce, ok := connectErr.(*connect.Error)
	if !ok {
		t.Fatalf("*connect.Error を期待したが %T が返った", connectErr)
	}
	if ce.Code() != connect.CodeNotFound {
		t.Errorf("コード: 期待 CodeNotFound, 実際 %v", ce.Code())
	}
}
