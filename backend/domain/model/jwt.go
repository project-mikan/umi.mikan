package model

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/project-mikan/umi.mikan/backend/constants"

	"github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
)

const (
	accessTokenExpirationMinutes = time.Minute * 15    // アクセストークンの有効期限
	refreshTokenExpirationDays   = 30 * 24 * time.Hour // リフレッシュトークンの有効期限
)

type TokenDetails struct {
	AccessToken  string
	TokenType    string
	ExpiresIn    int64
	RefreshToken string
}

type Claims struct {
	UserID string
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
		UserID: userID,
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
		UserID: userID,
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
		UserID: userID,
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

func (m *TokenDetails) ConvertAuthResponse() *grpc.AuthResponse {
	return &grpc.AuthResponse{
		AccessToken:  m.AccessToken,
		TokenType:    m.TokenType,
		ExpiresIn:    int32(m.ExpiresIn),
		RefreshToken: m.RefreshToken,
	}
}
