package mcpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

func generateValidTokenForTest(t *testing.T, userID string) string {
	t.Helper()
	tokens, err := model.GenerateAuthTokens(userID)
	if err != nil {
		t.Fatalf("トークン生成失敗: %v", err)
	}
	return tokens.AccessToken
}

func TestAuthMiddleware(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUserID   bool
	}{
		{
			name:           "正常系: 有効なBearerトークンでユーザーIDがコンテキストに注入される",
			authHeader:     "Bearer " + generateValidTokenForTest(t, userID),
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
		{
			name:           "異常系: Authorizationヘッダーがない場合は401",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "異常系: Bearerプレフィックスがない場合は401",
			authHeader:     "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "異常系: Bearerの後にトークンがない場合は401",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "異常系: 不正なJWTトークンは401",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedUserID string
			var nextCalled bool
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if val, ok := r.Context().Value(middleware.UserIDKey).(string); ok {
					capturedUserID = val
				}
				w.WriteHeader(http.StatusOK)
			})

			server := httptest.NewServer(AuthMiddleware(next))
			defer server.Close()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, nil)
			if err != nil {
				t.Fatalf("リクエスト作成失敗: %v", err)
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("リクエスト送信失敗: %v", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("ステータスコード: 期待 %d, 実際 %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.expectUserID {
				if !nextCalled {
					t.Error("認証成功時はnextが呼ばれるべき")
				}
				if capturedUserID != userID {
					t.Errorf("ユーザーID: 期待 %v, 実際 %v", userID, capturedUserID)
				}
			} else if nextCalled {
				t.Error("認証失敗時にnextが呼ばれるべきではない")
			}
		})
	}
}
