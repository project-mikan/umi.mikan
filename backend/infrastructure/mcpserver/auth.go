package mcpserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

// AuthMiddleware は Authorization: Bearer <JWT> ヘッダーを検証し、
// gRPC/ConnectRPCの認証インターセプターと同じロジックでユーザーIDをコンテキストに注入する。
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		if accessToken == "" {
			http.Error(w, "empty access token", http.StatusUnauthorized)
			return
		}

		_, userID, err := model.ParseAuthTokens(accessToken)
		if err != nil {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
