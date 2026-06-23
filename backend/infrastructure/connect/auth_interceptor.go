package connect

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

// 認証不要なプロシージャの一覧
var authExemptProcedures = map[string]bool{
	"/auth.AuthService/RegisterByPassword":    true,
	"/auth.AuthService/LoginByPassword":       true,
	"/auth.AuthService/RefreshAccessToken":    true,
	"/auth.AuthService/GetRegistrationConfig": true,
}

// NewAuthInterceptor ConnectRPC 用の認証インターセプターを返す。
// gRPC の AuthInterceptor と同じロジックで JWT を検証し、ユーザーIDをコンテキストに注入する。
func NewAuthInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure

			// 認証不要なエンドポイントはそのまま通す
			if authExemptProcedures[procedure] {
				return next(ctx, req)
			}

			// Authorization ヘッダーを取得
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			accessToken := strings.TrimPrefix(authHeader, "Bearer ")
			if accessToken == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			// JWT を検証してユーザーIDを取得
			_, userID, err := model.ParseAuthTokens(accessToken)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// ユーザーIDをコンテキストに注入（gRPC ミドルウェアと同じキーを使う）
			ctx = context.WithValue(ctx, middleware.UserIDKey, userID)

			return next(ctx, req)
		}
	}
}
