package middleware

import (
	"context"
	"testing"

	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TestAuthInterceptor tests the AuthInterceptor function
func TestAuthInterceptor(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		setupContext   func() context.Context
		expectError    bool
		expectedCode   codes.Code
		expectedUserID string
	}{
		{
			name:   "認証不要なメソッド: RegisterByPassword",
			method: "/auth.AuthService/RegisterByPassword",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectError: false,
		},
		{
			name:   "認証不要なメソッド: LoginByPassword",
			method: "/auth.AuthService/LoginByPassword",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectError: false,
		},
		{
			name:   "認証不要なメソッド: RefreshAccessToken",
			method: "/auth.AuthService/RefreshAccessToken",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectError: false,
		},
		{
			name:   "認証不要なメソッド: GetRegistrationConfig",
			method: "/auth.AuthService/GetRegistrationConfig",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectError: false,
		},
		{
			name:   "メタデータがない場合",
			method: "/diary.DiaryService/CreateDiaryEntry",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectError:  true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:   "authorizationヘッダーがない場合",
			method: "/diary.DiaryService/CreateDiaryEntry",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectError:  true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:   "Bearer プレフィックスがない場合",
			method: "/diary.DiaryService/CreateDiaryEntry",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "InvalidToken",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectError:  true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:   "空のアクセストークンの場合",
			method: "/diary.DiaryService/CreateDiaryEntry",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer ",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectError:  true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:   "無効なトークンの場合",
			method: "/diary.DiaryService/CreateDiaryEntry",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer invalid-token",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectError:  true,
			expectedCode: codes.Unauthenticated,
		},
		{
			name:   "有効なトークンの場合",
			method: "/diary.DiaryService/CreateDiaryEntry",
			setupContext: func() context.Context {
				userID := "test-user-id"
				tokens, err := model.GenerateAuthTokens(userID)
				if err != nil {
					t.Fatalf("failed to generate tokens: %v", err)
				}
				md := metadata.New(map[string]string{
					"authorization": "Bearer " + tokens.AccessToken,
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectError:    false,
			expectedUserID: "test-user-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupContext()
			info := &grpc.UnaryServerInfo{
				FullMethod: tc.method,
			}

			// ダミーハンドラーを作成
			handler := func(ctx context.Context, req any) (any, error) {
				// コンテキストからユーザーIDを取得できるか確認
				if tc.expectedUserID != "" {
					userID, err := GetUserIDFromContext(ctx)
					if err != nil {
						t.Errorf("expected user ID in context, got error: %v", err)
					}
					if userID != tc.expectedUserID {
						t.Errorf("expected user ID %s, got %s", tc.expectedUserID, userID)
					}
				}
				return "success", nil
			}

			// インターセプターを実行
			_, err := AuthInterceptor(ctx, nil, info, handler)

			// エラーチェック
			if tc.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else {
					st, ok := status.FromError(err)
					if !ok {
						t.Errorf("expected gRPC status error, got %v", err)
					}
					if st.Code() != tc.expectedCode {
						t.Errorf("expected code %v, got %v", tc.expectedCode, st.Code())
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

// TestIsAuthExempt tests the isAuthExempt function
func TestIsAuthExempt(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		expected bool
	}{
		{
			name:     "RegisterByPassword は認証不要",
			method:   "/auth.AuthService/RegisterByPassword",
			expected: true,
		},
		{
			name:     "LoginByPassword は認証不要",
			method:   "/auth.AuthService/LoginByPassword",
			expected: true,
		},
		{
			name:     "RefreshAccessToken は認証不要",
			method:   "/auth.AuthService/RefreshAccessToken",
			expected: true,
		},
		{
			name:     "GetRegistrationConfig は認証不要",
			method:   "/auth.AuthService/GetRegistrationConfig",
			expected: true,
		},
		{
			name:     "CreateDiaryEntry は認証が必要",
			method:   "/diary.DiaryService/CreateDiaryEntry",
			expected: false,
		},
		{
			name:     "GetDiaryEntry は認証が必要",
			method:   "/diary.DiaryService/GetDiaryEntry",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isAuthExempt(tc.method)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestGetUserIDFromContext tests the GetUserIDFromContext function
func TestGetUserIDFromContext(t *testing.T) {
	testCases := []struct {
		name        string
		setupCtx    func() context.Context
		expectError bool
		expectedID  string
	}{
		{
			name: "正常なユーザーIDの取得",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "test-user-id")
			},
			expectError: false,
			expectedID:  "test-user-id",
		},
		{
			name: "ユーザーIDがコンテキストにない場合",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectError: true,
		},
		{
			name: "ユーザーIDが空文字列の場合",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "")
			},
			expectError: true,
		},
		{
			name: "ユーザーIDが文字列でない場合",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, 123)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx()
			userID, err := GetUserIDFromContext(ctx)

			if tc.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if userID != tc.expectedID {
					t.Errorf("expected user ID %s, got %s", tc.expectedID, userID)
				}
			}
		})
	}
}
