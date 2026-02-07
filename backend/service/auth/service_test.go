package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/testutil"
	"google.golang.org/grpc/metadata"
)

func setupTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func generateTestEmail(t *testing.T, prefix string) string {
	testID := fmt.Sprintf("%s-%d-%d", t.Name(), os.Getpid(), time.Now().UnixNano())
	testID = strings.ReplaceAll(testID, "/", "-")
	testID = strings.ReplaceAll(testID, " ", "-")
	return fmt.Sprintf("%s-%s@example.com", prefix, testID)
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestAuthEntry_RegisterByPassword(t *testing.T) {
	db := setupTestDB(t)

	authService := &AuthEntry{DB: db}
	ctx := context.Background()

	tests := []struct {
		name          string
		request       *g.RegisterByPasswordRequest
		shouldSucceed bool
		expectedError string
	}{
		{
			name: "正常系：正常な登録",
			request: &g.RegisterByPasswordRequest{
				Email:    generateTestEmail(t, "test"),
				Password: "validPassword123",
				Name:     "Test User",
			},
			shouldSucceed: true,
		},
		{
			name: "異常系：空のメールアドレス",
			request: &g.RegisterByPasswordRequest{
				Email:    "",
				Password: "validPassword123",
				Name:     "Test User",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name: "異常系：無効なメールアドレス形式",
			request: &g.RegisterByPasswordRequest{
				Email:    "invalid-email",
				Password: "validPassword123",
				Name:     "Test User",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name: "異常系：パスワードが短すぎる",
			request: &g.RegisterByPasswordRequest{
				Email:    "test2@example.com",
				Password: "123",
				Name:     "Test User",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name: "異常系：空の名前",
			request: &g.RegisterByPasswordRequest{
				Email:    "test3@example.com",
				Password: "validPassword123",
				Name:     "",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := authService.RegisterByPassword(ctx, tt.request)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if response == nil {
					t.Fatal("Expected response but got nil")
				}
				if response.AccessToken == "" {
					t.Error("Expected access token but got empty string")
				}
				if response.RefreshToken == "" {
					t.Error("Expected refresh token but got empty string")
				}
				if response.TokenType != "Bearer" {
					t.Errorf("Expected token type 'Bearer' but got '%s'", response.TokenType)
				}
				if response.ExpiresIn <= 0 {
					t.Errorf("Expected positive expires_in but got %d", response.ExpiresIn)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
				if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s' but got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

func TestAuthEntry_LoginByPassword(t *testing.T) {
	db := setupTestDB(t)

	authService := &AuthEntry{DB: db}
	ctx := context.Background()

	// まず、ユーザーを登録
	testEmail := generateTestEmail(t, "login-test")
	registerReq := &g.RegisterByPasswordRequest{
		Email:    testEmail,
		Password: "validPassword123",
		Name:     "Login Test User",
	}
	_, err := authService.RegisterByPassword(ctx, registerReq)
	if err != nil {
		t.Fatalf("Failed to register user for login test: %v", err)
	}

	tests := []struct {
		name          string
		request       *g.LoginByPasswordRequest
		shouldSucceed bool
		expectedError string
	}{
		{
			name: "正常系：正常なログイン",
			request: &g.LoginByPasswordRequest{
				Email:    testEmail,
				Password: "validPassword123",
			},
			shouldSucceed: true,
		},
		{
			name: "異常系：無効なメールアドレス",
			request: &g.LoginByPasswordRequest{
				Email:    "nonexistent@example.com",
				Password: "validPassword123",
			},
			shouldSucceed: false,
			expectedError: "user not found",
		},
		{
			name: "異常系：無効なパスワード",
			request: &g.LoginByPasswordRequest{
				Email:    testEmail,
				Password: "wrongPassword",
			},
			shouldSucceed: false,
			expectedError: "password does not match",
		},
		{
			name: "異常系：空のメールアドレス",
			request: &g.LoginByPasswordRequest{
				Email:    "",
				Password: "validPassword123",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name: "異常系：空のパスワード",
			request: &g.LoginByPasswordRequest{
				Email:    "login-test@example.com",
				Password: "",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := authService.LoginByPassword(ctx, tt.request)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if response == nil {
					t.Fatal("Expected response but got nil")
				}
				if response.AccessToken == "" {
					t.Error("Expected access token but got empty string")
				}
				if response.RefreshToken == "" {
					t.Error("Expected refresh token but got empty string")
				}
				if response.TokenType != "Bearer" {
					t.Errorf("Expected token type 'Bearer' but got '%s'", response.TokenType)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
				if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s' but got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

func TestAuthEntry_RefreshAccessToken(t *testing.T) {
	db := setupTestDB(t)

	authService := &AuthEntry{DB: db}
	ctx := context.Background()

	// First, register a user and get tokens
	registerReq := &g.RegisterByPasswordRequest{
		Email:    generateTestEmail(t, "refresh-test"),
		Password: "validPassword123",
		Name:     "Refresh Test User",
	}
	registerResp, err := authService.RegisterByPassword(ctx, registerReq)
	if err != nil {
		t.Fatalf("Failed to register user for refresh test: %v", err)
	}

	tests := []struct {
		name          string
		refreshToken  string
		shouldSucceed bool
		expectedError string
	}{
		{
			name:          "正常系：正常なリフレッシュトークン",
			refreshToken:  registerResp.RefreshToken,
			shouldSucceed: true,
		},
		{
			name:          "異常系：無効なリフレッシュトークン",
			refreshToken:  "invalid.token.here",
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name:          "異常系：空のリフレッシュトークン",
			refreshToken:  "",
			shouldSucceed: false,
			expectedError: "validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshReq := &g.RefreshAccessTokenRequest{
				RefreshToken: tt.refreshToken,
			}
			response, err := authService.RefreshAccessToken(ctx, refreshReq)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if response == nil {
					t.Fatal("Expected response but got nil")
				}
				if response.AccessToken == "" {
					t.Error("Expected access token but got empty string")
				}
				if response.TokenType != "Bearer" {
					t.Errorf("Expected token type 'Bearer' but got '%s'", response.TokenType)
				}
				// Refresh token should be empty for access token refresh
				if response.RefreshToken != "" {
					t.Error("Expected empty refresh token for access token refresh")
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
				if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s' but got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

func TestAuthEntry_DuplicateRegistration(t *testing.T) {
	db := setupTestDB(t)

	authService := &AuthEntry{DB: db}
	ctx := context.Background()

	// Register a user
	testEmail := generateTestEmail(t, "duplicate-test")
	registerReq := &g.RegisterByPasswordRequest{
		Email:    testEmail,
		Password: "validPassword123",
		Name:     "Duplicate Test User",
	}
	_, err := authService.RegisterByPassword(ctx, registerReq)
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	// Try to register the same user again
	_, err = authService.RegisterByPassword(ctx, registerReq)
	if err == nil {
		t.Error("Expected error for duplicate registration but got nil")
	} else if !containsString(err.Error(), "already exists") {
		t.Errorf("Expected error about user already existing but got: %v", err)
	}
}

func TestAuthEntry_GetRegistrationConfig(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	tests := []struct {
		name                string
		registerKey         string
		expectedKeyRequired bool
	}{
		{
			name:                "登録キーが設定されている場合",
			registerKey:         "test-secret-key",
			expectedKeyRequired: true,
		},
		{
			name:                "登録キーが設定されていない場合",
			registerKey:         "",
			expectedKeyRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &AuthEntry{
				DB:          db,
				RegisterKey: tt.registerKey,
			}

			req := &g.GetRegistrationConfigRequest{}
			resp, err := authService.GetRegistrationConfig(ctx, req)

			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}
			if resp == nil {
				t.Fatal("Expected response but got nil")
			}
			if resp.RegisterKeyRequired != tt.expectedKeyRequired {
				t.Errorf("Expected RegisterKeyRequired=%v but got %v", tt.expectedKeyRequired, resp.RegisterKeyRequired)
			}
		})
	}
}

func TestAuthEntry_RegisterByPasswordWithRegisterKey(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	const correctKey = "correct-secret-key"

	tests := []struct {
		name          string
		registerKey   string // サーバー側に設定される登録キー
		requestKey    string // リクエストで送信される登録キー
		email         string
		password      string
		userName      string
		shouldSucceed bool
		expectedError string
	}{
		{
			name:          "正常系：登録キーが設定されており、正しいキーを送信",
			registerKey:   correctKey,
			requestKey:    correctKey,
			email:         generateTestEmail(t, "with-key-valid"),
			password:      "validPassword123",
			userName:      "Test User",
			shouldSucceed: true,
		},
		{
			name:          "異常系：登録キーが設定されているが、キーが空",
			registerKey:   correctKey,
			requestKey:    "",
			email:         generateTestEmail(t, "with-key-empty"),
			password:      "validPassword123",
			userName:      "Test User",
			shouldSucceed: false,
			expectedError: "registration failed",
		},
		{
			name:          "異常系：登録キーが設定されているが、間違ったキーを送信",
			registerKey:   correctKey,
			requestKey:    "wrong-key",
			email:         generateTestEmail(t, "with-key-wrong"),
			password:      "validPassword123",
			userName:      "Test User",
			shouldSucceed: false,
			expectedError: "registration failed",
		},
		{
			name:          "正常系：登録キーが設定されていない場合、キー無しで登録可能",
			registerKey:   "",
			requestKey:    "",
			email:         generateTestEmail(t, "without-key-empty"),
			password:      "validPassword123",
			userName:      "Test User",
			shouldSucceed: true,
		},
		{
			name:          "正常系：登録キーが設定されていない場合、キーがあっても無視される",
			registerKey:   "",
			requestKey:    "some-key",
			email:         generateTestEmail(t, "without-key-with-value"),
			password:      "validPassword123",
			userName:      "Test User",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &AuthEntry{
				DB:          db,
				RegisterKey: tt.registerKey,
			}

			req := &g.RegisterByPasswordRequest{
				Email:       tt.email,
				Password:    tt.password,
				Name:        tt.userName,
				RegisterKey: tt.requestKey,
			}

			resp, err := authService.RegisterByPassword(ctx, req)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if resp == nil {
					t.Fatal("Expected response but got nil")
				}
				if resp.AccessToken == "" {
					t.Error("Expected access token but got empty string")
				}
				if resp.RefreshToken == "" {
					t.Error("Expected refresh token but got empty string")
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
				if tt.expectedError != "" && !containsString(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s' but got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	authService := &AuthEntry{}

	tests := []struct {
		name     string
		setupCtx func() context.Context
		expected string
	}{
		{
			name: "X-Forwarded-Forヘッダーがある場合",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{
					"x-forwarded-for": "192.168.1.1, 10.0.0.1",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expected: "192.168.1.1",
		},
		{
			name: "X-Real-IPヘッダーがある場合",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{
					"x-real-ip": "192.168.1.2",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expected: "192.168.1.2",
		},
		{
			name: "ヘッダーがない場合",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := authService.getClientIP(ctx)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetUserAgent(t *testing.T) {
	authService := &AuthEntry{}

	tests := []struct {
		name     string
		setupCtx func() context.Context
		expected string
	}{
		{
			name: "User-Agentヘッダーがある場合",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{
					"user-agent": "Mozilla/5.0",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expected: "Mozilla/5.0",
		},
		{
			name: "grpcgateway-user-agentヘッダーがある場合",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{
					"grpcgateway-user-agent": "gRPC-Gateway/1.0",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expected: "gRPC-Gateway/1.0",
		},
		{
			name: "ヘッダーがない場合",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := authService.getUserAgent(ctx)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetClientIdentifier(t *testing.T) {
	authService := &AuthEntry{}

	tests := []struct {
		name     string
		setupCtx func() context.Context
		expected string
	}{
		{
			name: "IPとUser-Agent両方がある場合",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{
					"x-forwarded-for": "192.168.1.1",
					"user-agent":      "Mozilla/5.0",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expected: "192.168.1.1:Mozilla/5.0",
		},
		{
			name: "X-Real-IPとgrpcgateway-user-agentがある場合",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{
					"x-real-ip":               "192.168.1.2",
					"grpcgateway-user-agent": "gRPC-Gateway/1.0",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expected: "192.168.1.2:gRPC-Gateway/1.0",
		},
		{
			name: "ヘッダーがない場合",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expected: "unknown:unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := authService.getClientIdentifier(ctx)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
