package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthInterceptor gRPCの認証インターセプター
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 認証が不要なメソッドをスキップ
	if isAuthExempt(info.FullMethod) {
		return handler(ctx, req)
	}

	// メタデータからAuthorizationヘッダーを取得
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	// "Bearer " プレフィックスを確認
	token := authHeader[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization format")
	}

	// トークンを抽出
	accessToken := strings.TrimPrefix(token, "Bearer ")
	if accessToken == "" {
		return nil, status.Errorf(codes.Unauthenticated, "empty access token")
	}

	// JWTトークンの検証
	_, userID, err := model.ParseAuthTokens(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid access token: %v", err)
	}

	// ユーザーIDをコンテキストに追加
	ctx = context.WithValue(ctx, UserIDKey, userID)

	// 認証済みのリクエストを処理
	return handler(ctx, req)
}

// isAuthExempt 認証が不要なメソッドかどうかを判定
func isAuthExempt(method string) bool {
	exemptMethods := []string{
		"/auth.AuthService/RegisterByPassword",
		"/auth.AuthService/LoginByPassword",
		"/auth.AuthService/RefreshAccessToken",
	}

	for _, exemptMethod := range exemptMethods {
		if method == exemptMethod {
			return true
		}
	}
	return false
}

// GetUserIDFromContext コンテキストからユーザーIDを取得
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}