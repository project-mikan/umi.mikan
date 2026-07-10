package mcpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/service/user"
	"github.com/project-mikan/umi.mikan/backend/testutil"
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

			// HTTPクライアント経由だとヘッダー末尾の空白がトリムされ「Bearer 」のケースを
			// 検証できないため、ハンドラーを直接呼び出す
			// JWT認証のみのケースなのでDBは使用されない（nilを渡す）
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			AuthMiddleware(nil, next).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("ステータスコード: 期待 %d, 実際 %d", tt.expectedStatus, rec.Code)
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

func TestAuthMiddleware_APIKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "mcp-api-key-auth@example.com", "MCPAPIKeyUser")
	userService := &user.UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	created, err := userService.CreateApiKey(ctx, &g.CreateApiKeyRequest{Name: "MCP認証テスト用キー"})
	if err != nil {
		t.Fatalf("APIキー発行に失敗: %v", err)
	}

	callWithToken := func(t *testing.T, token string) (int, string) {
		t.Helper()
		var capturedUserID string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if val, ok := r.Context().Value(middleware.UserIDKey).(string); ok {
				capturedUserID = val
			}
			w.WriteHeader(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		AuthMiddleware(db, next).ServeHTTP(rec, req)
		return rec.Code, capturedUserID
	}

	t.Run("正常系: 有効なAPIキーで認証されユーザーIDが注入される", func(t *testing.T) {
		status, capturedUserID := callWithToken(t, created.ApiKey)
		if status != http.StatusOK {
			t.Fatalf("ステータスコード: 期待 200, 実際 %d", status)
		}
		if capturedUserID != userID.String() {
			t.Errorf("ユーザーID: 期待 %v, 実際 %v", userID, capturedUserID)
		}

		// 最終使用日時が更新されている
		listResp, err := userService.ListApiKeys(ctx, &g.ListApiKeysRequest{})
		if err != nil {
			t.Fatalf("一覧取得に失敗: %v", err)
		}
		if len(listResp.ApiKeys) != 1 || listResp.ApiKeys[0].LastUsedAt == 0 {
			t.Errorf("認証後にLastUsedAtが更新されていない: %+v", listResp.ApiKeys)
		}
	})

	t.Run("異常系: 存在しないAPIキーは401", func(t *testing.T) {
		status, _ := callWithToken(t, "umi_0000000000000000000000000000000000000000000000000000000000000000")
		if status != http.StatusUnauthorized {
			t.Errorf("ステータスコード: 期待 401, 実際 %d", status)
		}
	})

	t.Run("異常系: 削除済みAPIキーは401", func(t *testing.T) {
		toDelete, err := userService.CreateApiKey(ctx, &g.CreateApiKeyRequest{Name: "削除予定キー"})
		if err != nil {
			t.Fatalf("APIキー発行に失敗: %v", err)
		}
		if _, err := userService.DeleteApiKey(ctx, &g.DeleteApiKeyRequest{Id: toDelete.Info.Id}); err != nil {
			t.Fatalf("APIキー削除に失敗: %v", err)
		}

		status, _ := callWithToken(t, toDelete.ApiKey)
		if status != http.StatusUnauthorized {
			t.Errorf("ステータスコード: 期待 401, 実際 %d", status)
		}
	})
}
