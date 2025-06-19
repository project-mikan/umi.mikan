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
			name: "Valid registration",
			request: &g.RegisterByPasswordRequest{
				Email:    generateTestEmail(t, "test"),
				Password: "validPassword123",
				Name:     "Test User",
			},
			shouldSucceed: true,
		},
		{
			name: "Empty email",
			request: &g.RegisterByPasswordRequest{
				Email:    "",
				Password: "validPassword123",
				Name:     "Test User",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name: "Invalid email format",
			request: &g.RegisterByPasswordRequest{
				Email:    "invalid-email",
				Password: "validPassword123",
				Name:     "Test User",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name: "Password too short",
			request: &g.RegisterByPasswordRequest{
				Email:    "test2@example.com",
				Password: "123",
				Name:     "Test User",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name: "Empty name",
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
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if response == nil {
					t.Error("Expected response but got nil")
					return
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
					t.Error("Expected error but got nil")
					return
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

	// First, register a user
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
			name: "Valid login",
			request: &g.LoginByPasswordRequest{
				Email:    testEmail,
				Password: "validPassword123",
			},
			shouldSucceed: true,
		},
		{
			name: "Invalid email",
			request: &g.LoginByPasswordRequest{
				Email:    "nonexistent@example.com",
				Password: "validPassword123",
			},
			shouldSucceed: false,
			expectedError: "user not found",
		},
		{
			name: "Invalid password",
			request: &g.LoginByPasswordRequest{
				Email:    testEmail,
				Password: "wrongPassword",
			},
			shouldSucceed: false,
			expectedError: "password does not match",
		},
		{
			name: "Empty email",
			request: &g.LoginByPasswordRequest{
				Email:    "",
				Password: "validPassword123",
			},
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name: "Empty password",
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
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if response == nil {
					t.Error("Expected response but got nil")
					return
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
					t.Error("Expected error but got nil")
					return
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
			name:          "Valid refresh token",
			refreshToken:  registerResp.RefreshToken,
			shouldSucceed: true,
		},
		{
			name:          "Invalid refresh token",
			refreshToken:  "invalid.token.here",
			shouldSucceed: false,
			expectedError: "validation error",
		},
		{
			name:          "Empty refresh token",
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
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if response == nil {
					t.Error("Expected response but got nil")
					return
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
					t.Error("Expected error but got nil")
					return
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

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}