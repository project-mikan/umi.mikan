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
