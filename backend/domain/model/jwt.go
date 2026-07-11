package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/project-mikan/umi.mikan/backend/constants"

	"github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
)

const (
	accessTokenExpirationMinutes = time.Minute * 15    // アクセストークンの有効期限
	refreshTokenExpirationDays   = 30 * 24 * time.Hour // リフレッシュトークンの有効期限

	// tokenUseAccess / tokenUseRefresh はJWTがアクセストークン/リフレッシュトークンの
	// どちらであるかをClaims.TokenUseに埋め込むための識別子。
	// アクセストークン専用の検証箇所（gRPC/ConnectRPC/MCPの認証ミドルウェア）が
	// 有効期限の長いリフレッシュトークンを誤って受理しないようにするために使う。
	tokenUseAccess  = "access"
	tokenUseRefresh = "refresh"
)

type TokenDetails struct {
	AccessToken  string
	TokenType    string
	ExpiresIn    int64
	RefreshToken string
}

type Claims struct {
	UserID string
	// TokenUse は "access" または "refresh"。アクセストークン専用の検証で
	// リフレッシュトークンを弾くために使用する（空文字は旧トークンとの後方互換用）。
	TokenUse string
	jwt.RegisteredClaims
}

func GenerateAuthTokens(userID string) (*TokenDetails, error) {
	jwtS, err := constants.LoadJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to load JWT Secret: %w", err)
	}
	jwtSecret := []byte(jwtS)

	// --- Access Token の生成 ---
	accessTokenExpiration := time.Now().Add(accessTokenExpirationMinutes)
	accessClaims := &Claims{
		UserID:   userID,
		TokenUse: tokenUseAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// --- Refresh Token の生成 ---
	refreshTokenExpiration := time.Now().Add(refreshTokenExpirationDays)
	refreshClaims := &Claims{
		UserID:   userID,
		TokenUse: tokenUseRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID,
			Issuer:    "your-app-issuer",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenDetails{
		AccessToken:  signedAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    accessTokenExpiration.Unix() - time.Now().Unix(), // 秒単位の残り時間を返す
		RefreshToken: signedRefreshToken,
	}, nil
}

// GenerateAccessToken Refresh時などRefreshTokenを変更せずにAccessTokenのみを生成する関数
func GenerateAccessToken(userID string) (*TokenDetails, error) {
	jwtS, err := constants.LoadJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to load JWT Secret: %w", err)
	}
	jwtSecret := []byte(jwtS)

	// --- Access Token の生成 ---
	accessTokenExpiration := time.Now().Add(accessTokenExpirationMinutes)
	accessClaims := &Claims{
		UserID:   userID,
		TokenUse: tokenUseAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}
	return &TokenDetails{
		AccessToken:  signedAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    accessTokenExpiration.Unix() - time.Now().Unix(), // 秒単位の残り時間を返す
		RefreshToken: "",                                               // RefreshTokenは変更しない
	}, nil
}

func ParseAuthTokens(tokenString string) (*TokenDetails, string, error) {
	jwtS, err := constants.LoadJWTSecret()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load JWT Secret: %w", err)
	}
	jwtSecret := []byte(jwtS)

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// アルゴリズムがHS256であることを確認
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &TokenDetails{
			AccessToken:  tokenString,
			TokenType:    "Bearer",
			ExpiresIn:    claims.ExpiresAt.Unix() - time.Now().Unix(),
			RefreshToken: "", // リフレッシュトークンはここでは不要
		}, claims.UserID, nil
	}
	return nil, "", fmt.Errorf("invalid token")
}

// ParseAccessToken はBearerトークンとして送られてきた文字列を検証し、
// それがアクセストークン（リフレッシュトークンではない）であることを確認した上でユーザーIDを返す。
// gRPC/ConnectRPC/MCPの認証ミドルウェアはすべてこの関数を使うべきで、
// ParseAuthTokens を直接使うとリフレッシュトークン（30日有効）がBearerトークンとして
// 受理されてしまい、短命であるべきアクセストークンの前提が崩れる。
// TokenUse が空文字の場合はこの変更以前に発行された旧アクセストークンとみなし、
// 後方互換のため許可する（旧トークンは最長15分で自然に失効するため実害は限定的）。
func ParseAccessToken(tokenString string) (*TokenDetails, string, error) {
	jwtS, err := constants.LoadJWTSecret()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load JWT Secret: %w", err)
	}
	jwtSecret := []byte(jwtS)

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, "", fmt.Errorf("invalid token")
	}
	if claims.TokenUse == tokenUseRefresh {
		return nil, "", fmt.Errorf("refresh token is not allowed as an access token")
	}

	return &TokenDetails{
		AccessToken:  tokenString,
		TokenType:    "Bearer",
		ExpiresIn:    claims.ExpiresAt.Unix() - time.Now().Unix(),
		RefreshToken: "",
	}, claims.UserID, nil
}

// ErrMissingAuthHeader / ErrInvalidAuthFormat / ErrEmptyBearerToken は
// ExtractBearerToken が返すエラーの種別を呼び出し側で判定できるようにするための番兵。
var (
	ErrMissingAuthHeader = fmt.Errorf("missing authorization header")
	ErrInvalidAuthFormat = fmt.Errorf("invalid authorization format")
	ErrEmptyBearerToken  = fmt.Errorf("empty bearer token")
)

// ExtractBearerToken は "Authorization: Bearer <token>" ヘッダーの値からトークン本体を取り出す。
// gRPC（metadata経由）、ConnectRPC・MCP（HTTPヘッダー経由）のいずれの認証ミドルウェアでも
// 同じ抽出ロジックを使うことで、Bearerヘッダーのパース仕様が実装ごとに乖離するのを防ぐ。
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingAuthHeader
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrInvalidAuthFormat
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", ErrEmptyBearerToken
	}
	return token, nil
}

func (m *TokenDetails) ConvertAuthResponse() *grpc.AuthResponse {
	return &grpc.AuthResponse{
		AccessToken:  m.AccessToken,
		TokenType:    m.TokenType,
		ExpiresIn:    int32(m.ExpiresIn),
		RefreshToken: m.RefreshToken,
	}
}
