package mcpserver

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

// notifyLastUsedUpdate はlast_used_at更新goroutineの完了をテストに通知するフック。
// テストビルド時はauth_hook_test.goで上書きされる。本番では何もしない。
var notifyLastUsedUpdate = func(err error) {}

// AuthMiddleware は Authorization: Bearer <トークン> ヘッダーを検証し、ユーザーIDをコンテキストに注入する。
// トークンは2種類を受け付ける:
//   - APIキー（umi_プレフィックス）: DBのSHA-256ハッシュと照合する。MCPクライアント向けの長期キー。
//   - JWTアクセストークン: gRPC/ConnectRPCの認証インターセプターと同じロジックで検証する（リフレッシュトークンは拒否）。
func AuthMiddleware(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := model.ExtractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var userID string
		if model.IsAPIKey(token) {
			// APIキー認証: ハッシュでDB照合する
			apiKey, err := database.UserAPIKeyByKeyHash(r.Context(), db, model.HashAPIKey(token))
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "invalid api key", http.StatusUnauthorized)
					return
				}
				// DB接続断など、キーの正当性とは無関係な障害はUnauthorizedと区別する。
				// 401のままだとクライアントが「キーが失効した」と誤解し不要なローテーションを招くため。
				log.Printf("failed to look up api key: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if time.Now().Unix() >= apiKey.ExpiresAt {
				http.Error(w, "api key expired", http.StatusUnauthorized)
				return
			}
			userID = apiKey.UserID.String()

			// 最終使用日時の更新は認証結果に影響しないbest-effort処理なので、
			// レスポンスを遅延させないよう非同期化する（毎リクエストの同期DB書き込みを避ける）。
			go func(keyID uuid.UUID) {
				updateErr := database.UpdateUserAPIKeyLastUsed(context.Background(), db, keyID, time.Now().Unix())
				if updateErr != nil {
					log.Printf("failed to update api key last_used_at: %v", updateErr)
				}
				notifyLastUsedUpdate(updateErr)
			}(apiKey.ID)
		} else {
			// JWTアクセストークン認証（リフレッシュトークンは拒否する）
			_, jwtUserID, err := model.ParseAccessToken(token)
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
