package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestAuthSuite_RegisterAndLogin(t *testing.T) {
	runner := testutil.NewTestRunner(t)
	runner.Run(func(suite *testutil.TestSuite) {
		authService := &AuthEntry{DB: suite.DB}
		ctx := context.Background()

		// Register a new user with unique email
		testID := fmt.Sprintf("%d", time.Now().UnixNano())
		registerReq := &g.RegisterByPasswordRequest{
			Email:    fmt.Sprintf("auth-suite-test-%s@example.com", testID),
			Password: "securePassword123",
			Name:     "Suite Test User",
		}
		registerResp, err := authService.RegisterByPassword(ctx, registerReq)
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}
		if registerResp.AccessToken == "" {
			t.Error("Expected access token after registration")
		}

		// Login with the registered user
		loginReq := &g.LoginByPasswordRequest{
			Email:    registerReq.Email,
			Password: registerReq.Password,
		}
		loginResp, err := authService.LoginByPassword(ctx, loginReq)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
		if loginResp.AccessToken == "" {
			t.Error("Expected access token after login")
		}

		// Refresh access token
		refreshReq := &g.RefreshAccessTokenRequest{
			RefreshToken: loginResp.RefreshToken,
		}
		refreshResp, err := authService.RefreshAccessToken(ctx, refreshReq)
		if err != nil {
			t.Fatalf("Token refresh failed: %v", err)
		}
		if refreshResp.AccessToken == "" {
			t.Error("Expected new access token after refresh")
		}
	})
}