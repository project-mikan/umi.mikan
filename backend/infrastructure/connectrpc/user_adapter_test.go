package connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/service/user"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func setupUserAdapter(t *testing.T) *UserServiceAdapter {
	t.Helper()
	db := testutil.SetupTestDB(t)
	svc := &user.UserEntry{DB: db}
	return &UserServiceAdapter{svc: svc}
}

func createAuthContext(userID string) context.Context {
	return context.WithValue(context.Background(), middleware.UserIDKey, userID)
}

func TestUserServiceAdapter_UpdateUserName(t *testing.T) {
	tests := []struct {
		name          string
		newName       string
		expectSuccess bool
		expectMsg     string
	}{
		{
			name:          "異常系: 空の名前はSuccess:falseを返す",
			newName:       "",
			expectSuccess: false,
			expectMsg:     "nameRequired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupUserAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.UpdateUserNameRequest{NewName: tt.newName})
			resp, err := adapter.UpdateUserName(ctx, req)

			// UpdateUserName はgRPCステータスエラーを返さず Success フラグで制御する
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Msg.GetSuccess() != tt.expectSuccess {
				t.Errorf("Success: 期待 %v, 実際 %v (msg: %v)", tt.expectSuccess, resp.Msg.GetSuccess(), resp.Msg.GetMessage())
			}
		})
	}
}

func TestUserServiceAdapter_GetUserInfo_Error(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		expectedCode connect.Code
	}{
		{
			name:         "異常系: 存在しないユーザーIDはNotFoundエラーになる",
			ctx:          createAuthContext(uuid.New().String()),
			expectedCode: connect.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupUserAdapter(t)
			req := connect.NewRequest(&g.GetUserInfoRequest{})
			_, err := adapter.GetUserInfo(tt.ctx, req)

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

func TestUserServiceAdapter_GetAutoSummarySettings(t *testing.T) {
	tests := []struct {
		name                string
		expectedAutoMonthly bool
		expectedAutoTrend   bool
	}{
		{
			name:                "正常系: 存在しないユーザーはデフォルト値を返す",
			expectedAutoMonthly: false,
			expectedAutoTrend:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupUserAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.GetAutoSummarySettingsRequest{})
			resp, err := adapter.GetAutoSummarySettings(ctx, req)

			// エラーを返さずデフォルト値を返す
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Msg.GetAutoSummaryMonthly() != tt.expectedAutoMonthly {
				t.Errorf("AutoSummaryMonthly: 期待 %v, 実際 %v", tt.expectedAutoMonthly, resp.Msg.GetAutoSummaryMonthly())
			}
			if resp.Msg.GetAutoLatestTrendEnabled() != tt.expectedAutoTrend {
				t.Errorf("AutoLatestTrendEnabled: 期待 %v, 実際 %v", tt.expectedAutoTrend, resp.Msg.GetAutoLatestTrendEnabled())
			}
		})
	}
}

func TestUserServiceAdapter_ChangePassword(t *testing.T) {
	tests := []struct {
		name          string
		request       *g.ChangePasswordRequest
		expectSuccess bool
	}{
		{
			name: "異常系: 空のパスワードはSuccess:falseを返す",
			request: &g.ChangePasswordRequest{
				CurrentPassword: "",
				NewPassword:     "",
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupUserAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(tt.request)
			resp, err := adapter.ChangePassword(ctx, req)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Msg.GetSuccess() != tt.expectSuccess {
				t.Errorf("Success: 期待 %v, 実際 %v (msg: %v)", tt.expectSuccess, resp.Msg.GetSuccess(), resp.Msg.GetMessage())
			}
		})
	}
}

func TestUserServiceAdapter_DeleteLLMKey(t *testing.T) {
	tests := []struct {
		name          string
		expectSuccess bool
	}{
		{
			name:          "異常系: LLMキーが未登録の場合はSuccess:falseを返す",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupUserAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.DeleteLLMKeyRequest{})
			resp, err := adapter.DeleteLLMKey(ctx, req)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Msg.GetSuccess() != tt.expectSuccess {
				t.Errorf("Success: 期待 %v, 実際 %v (msg: %v)", tt.expectSuccess, resp.Msg.GetSuccess(), resp.Msg.GetMessage())
			}
		})
	}
}

func TestUserServiceAdapter_DeleteAccount(t *testing.T) {
	tests := []struct {
		name          string
		expectSuccess bool
	}{
		{
			name:          "異常系: 存在しないユーザーの削除はSuccess:falseを返す",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupUserAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.DeleteAccountRequest{})
			resp, err := adapter.DeleteAccount(ctx, req)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Msg.GetSuccess() != tt.expectSuccess {
				t.Errorf("Success: 期待 %v, 実際 %v (msg: %v)", tt.expectSuccess, resp.Msg.GetSuccess(), resp.Msg.GetMessage())
			}
		})
	}
}

func TestUserServiceAdapter_UpdateAutoSummarySettings(t *testing.T) {
	tests := []struct {
		name          string
		expectSuccess bool
	}{
		{
			name:          "異常系: LLMキーが未登録のユーザーは Success:falseを返す",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupUserAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.UpdateAutoSummarySettingsRequest{})
			resp, err := adapter.UpdateAutoSummarySettings(ctx, req)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Msg.GetSuccess() != tt.expectSuccess {
				t.Errorf("Success: 期待 %v, 実際 %v (msg: %v)", tt.expectSuccess, resp.Msg.GetSuccess(), resp.Msg.GetMessage())
			}
		})
	}
}

func TestUserServiceAdapter_UpdateLLMKey(t *testing.T) {
	tests := []struct {
		name          string
		request       *g.UpdateLLMKeyRequest
		expectSuccess bool
	}{
		{
			name: "異常系: 空のキーはSuccess:falseを返す",
			request: &g.UpdateLLMKeyRequest{
				Key: "",
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupUserAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(tt.request)
			resp, err := adapter.UpdateLLMKey(ctx, req)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Msg.GetSuccess() != tt.expectSuccess {
				t.Errorf("Success: 期待 %v, 実際 %v (msg: %v)", tt.expectSuccess, resp.Msg.GetSuccess(), resp.Msg.GetMessage())
			}
		})
	}
}
