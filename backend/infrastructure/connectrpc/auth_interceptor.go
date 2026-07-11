package connect

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

// 認証不要なプロシージャの一覧（書き換えを防ぐため関数で返す）
func isAuthExemptProcedure(procedure string) bool {
	switch procedure {
	case "/auth.AuthService/RegisterByPassword",
		"/auth.AuthService/LoginByPassword",
		"/auth.AuthService/RefreshAccessToken",
		"/auth.AuthService/GetRegistrationConfig":
		return true
	default:
		return false
	}
}

// NewAuthInterceptor ConnectRPC 用の認証インターセプターを返す。
// gRPC の AuthInterceptor と同じロジックで JWT を検証し、ユーザーIDをコンテキストに注入する。
// HTTPヘッダーからクライアントIPとUser-Agentも抽出してコンテキストに注入する（レートリミット用）。
func NewAuthInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := req.Spec().Procedure

			// HTTPヘッダーからクライアント識別情報を取得してコンテキストに注入する。
			// gRPC metadata の代わりに ConnectRPC では HTTP ヘッダーを参照する必要があるため、
			// サービス層が metadata.FromIncomingContext で取れない情報をここで補完する。
			clientIP := extractClientIP(req.Header())
			if clientIP != "" {
				ctx = context.WithValue(ctx, middleware.ConnectClientIPKey, clientIP)
			}
			userAgent := req.Header().Get("User-Agent")
			if userAgent != "" {
				ctx = context.WithValue(ctx, middleware.ConnectUserAgentKey, userAgent)
			}

			// 認証不要なエンドポイントはそのまま通す
			if isAuthExemptProcedure(procedure) {
				return next(ctx, req)
			}

			// Authorization ヘッダーからBearerトークンを抽出
			accessToken, err := model.ExtractBearerToken(req.Header().Get("Authorization"))
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// JWT を検証してユーザーIDを取得（リフレッシュトークンは拒否する）
			_, userID, err := model.ParseAccessToken(accessToken)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// ユーザーIDをコンテキストに注入（gRPC ミドルウェアと同じキーを使う）
			ctx = context.WithValue(ctx, middleware.UserIDKey, userID)

			return next(ctx, req)
		}
	}
}

// extractClientIP HTTP ヘッダーからクライアントIPを取得する。
// X-Forwarded-For → X-Real-IP の順で探し、見つからなければ空文字を返す。
func extractClientIP(header interface{ Get(string) string }) string {
	if xff := header.Get("X-Forwarded-For"); xff != "" {
		ip := strings.TrimSpace(strings.Split(xff, ",")[0])
		if ip != "" {
			return ip
		}
	}
	if xri := header.Get("X-Real-Ip"); xri != "" {
		return strings.TrimSpace(xri)
	}
	return ""
}
