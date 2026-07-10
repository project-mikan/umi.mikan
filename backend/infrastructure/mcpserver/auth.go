package mcpserver

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

// AuthMiddleware は Authorization: Bearer <トークン> ヘッダーを検証し、ユーザーIDをコンテキストに注入する。
// トークンは2種類を受け付ける:
//   - APIキー（umi_プレフィックス）: DBのSHA-256ハッシュと照合する。MCPクライアント向けの長期キー。
//   - JWTアクセストークン: gRPC/ConnectRPCの認証インターセプターと同じロジックで検証する。
func AuthMiddleware(db *sql.DB, next http.Handler) http.Handler {
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

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "empty access token", http.StatusUnauthorized)
			return
		}

		var userID string
		if model.IsAPIKey(token) {
			// APIキー認証: ハッシュでDB照合する
			apiKey, err := database.UserAPIKeyByKeyHash(r.Context(), db, model.HashAPIKey(token))
			if err != nil {
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}
			userID = apiKey.UserID.String()

			// 最終使用日時を更新する（失敗しても認証は継続する）
			if err := database.UpdateUserAPIKeyLastUsed(r.Context(), db, apiKey.ID, time.Now().Unix()); err != nil {
				log.Printf("failed to update api key last_used_at: %v", err)
			}
		} else {
			// JWTアクセストークン認証
			_, jwtUserID, err := model.ParseAuthTokens(token)
			if err != nil {
				http.Error(w, "invalid access token", http.StatusUnauthorized)
				return
			}
			userID = jwtUserID
		}

		ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
