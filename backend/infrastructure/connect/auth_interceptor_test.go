package connect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/grpc/grpcconnect"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

// testAuthHandler は認証インターセプターのテスト用ダミーハンドラー
type testAuthHandler struct {
	capturedUserID string
}

func (h *testAuthHandler) GetRegistrationConfig(_ context.Context, _ *connect.Request[g.GetRegistrationConfigRequest]) (*connect.Response[g.GetRegistrationConfigResponse], error) {
	return connect.NewResponse(&g.GetRegistrationConfigResponse{}), nil
}

func (h *testAuthHandler) RegisterByPassword(_ context.Context, req *connect.Request[g.RegisterByPasswordRequest]) (*connect.Response[g.AuthResponse], error) {
	h.capturedUserID = req.Header().Get("X-Captured-User")
	return connect.NewResponse(&g.AuthResponse{}), nil
}

func (h *testAuthHandler) LoginByPassword(_ context.Context, _ *connect.Request[g.LoginByPasswordRequest]) (*connect.Response[g.AuthResponse], error) {
	return connect.NewResponse(&g.AuthResponse{}), nil
}

func (h *testAuthHandler) RefreshAccessToken(_ context.Context, _ *connect.Request[g.RefreshAccessTokenRequest]) (*connect.Response[g.AuthResponse], error) {
	return connect.NewResponse(&g.AuthResponse{}), nil
}

func generateValidTokenForTest(t *testing.T, userID string) string {
	t.Helper()
	tokens, err := model.GenerateAuthTokens(userID)
	if err != nil {
		t.Fatalf("トークン生成失敗: %v", err)
	}
	return tokens.AccessToken
}

// newTestServer はテスト用の HTTP サーバーを起動する
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	interceptor := NewAuthInterceptor()
	handler := &testAuthHandler{}
	mux := http.NewServeMux()
	path, h := grpcconnect.NewAuthServiceHandler(handler, connect.WithInterceptors(interceptor))
	mux.Handle(path, h)
	return httptest.NewServer(mux)
}

// connectPost は Connect プロトコルの JSON リクエストを送信し HTTP ステータスを返す
func connectPost(t *testing.T, server *httptest.Server, procedure, authHeader string) int {
	t.Helper()
	url := server.URL + procedure
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("リクエスト作成失敗: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("リクエスト送信失敗: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("resp.Body.Close 失敗: %v", err)
		}
	}()
	return resp.StatusCode
}

func TestNewAuthInterceptor_ExemptProcedures(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	exemptProcedures := []struct {
		name      string
		procedure string
	}{
		{"RegisterByPassword", grpcconnect.AuthServiceRegisterByPasswordProcedure},
		{"LoginByPassword", grpcconnect.AuthServiceLoginByPasswordProcedure},
		{"RefreshAccessToken", grpcconnect.AuthServiceRefreshAccessTokenProcedure},
		{"GetRegistrationConfig", grpcconnect.AuthServiceGetRegistrationConfigProcedure},
	}

	for _, tt := range exemptProcedures {
		t.Run("正常系: 認証不要エンドポイント "+tt.name+" はトークンなしで通過する", func(t *testing.T) {
			status := connectPost(t, server, tt.procedure, "")
			// 200 OK（認証を通過してハンドラーが呼ばれた）
			if status != http.StatusOK {
				t.Errorf("HTTP ステータス: 期待 200, 実際 %d", status)
			}
		})
	}
}

func TestNewAuthInterceptor_AuthRequired(t *testing.T) {
	userID := uuid.New().String()
	server := newTestServer(t)
	defer server.Close()

	// 認証が必要なエンドポイントのテストには DiaryService を別途用意する必要があるが、
	// AuthService の認証ロジックはインターセプターで制御される。
	// ここではインターセプターを直接呼び出してテストする。
	interceptorFunc := NewAuthInterceptor()

	tests := []struct {
		name         string
		authHeader   string
		procedure    string
		expectErr    bool
		expectedCode connect.Code
		expectUserID bool
	}{
		{
			name:         "正常系: 有効なBearerトークンでユーザーIDがコンテキストに注入される",
			authHeader:   "Bearer " + generateValidTokenForTest(t, userID),
			procedure:    "/diary.DiaryService/CreateDiaryEntry",
			expectErr:    false,
			expectUserID: true,
		},
		{
			name:         "異常系: Authorizationヘッダーがない場合はUnauthenticatedエラーになる",
			authHeader:   "",
			procedure:    "/diary.DiaryService/CreateDiaryEntry",
			expectErr:    true,
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:         "異常系: Bearer プレフィックスがないトークンはUnauthenticatedエラーになる",
			authHeader:   "invalid-token",
			procedure:    "/diary.DiaryService/CreateDiaryEntry",
			expectErr:    true,
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:         "異常系: Bearerの後にトークンがない場合はUnauthenticatedエラーになる",
			authHeader:   "Bearer ",
			procedure:    "/diary.DiaryService/CreateDiaryEntry",
			expectErr:    true,
			expectedCode: connect.CodeUnauthenticated,
		},
		{
			name:         "異常系: 不正なJWTトークンはUnauthenticatedエラーになる",
			authHeader:   "Bearer invalid.jwt.token",
			procedure:    "/diary.DiaryService/CreateDiaryEntry",
			expectErr:    true,
			expectedCode: connect.CodeUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedUserID string
			var nextCalled bool
			next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				nextCalled = true
				if val, ok := ctx.Value(middleware.UserIDKey).(string); ok {
					capturedUserID = val
				}
				return connect.NewResponse(&g.CreateDiaryEntryResponse{}), nil
			}

			wrappedNext := interceptorFunc(next)

			// connect.Request を生成。Spec の Procedure はパッケージ外から設定不可なので、
			// 認証が必要なパスのテストはヘッダー検証のみを行う。
			req := connect.NewRequest(&g.CreateDiaryEntryRequest{})
			if tt.authHeader != "" {
				req.Header().Set("Authorization", tt.authHeader)
			}

			_, err := wrappedNext(context.Background(), req)

			if tt.expectErr {
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
				if nextCalled {
					t.Error("エラー時にnextが呼ばれるべきではない")
				}
			} else {
				if err != nil {
					t.Fatalf("エラーを期待しなかったが %v が返った", err)
				}
				if !nextCalled {
					t.Error("nextが呼ばれていない")
				}
				if tt.expectUserID && capturedUserID != userID {
					t.Errorf("ユーザーID: 期待 %v, 実際 %v", userID, capturedUserID)
				}
			}
		})
	}
}
