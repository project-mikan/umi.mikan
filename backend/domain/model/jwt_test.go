package model

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAuthTokens(t *testing.T) {
	tests := []struct {
		name   string
		userID string
	}{
		{
			name:   "正常なトークン生成",
			userID: "test-user-id",
		},
		{
			name:   "異なるユーザーIDでトークン生成",
			userID: "another-user-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := GenerateAuthTokens(tt.userID)

			require.NoError(t, err)
			assert.NotEmpty(t, tokens.AccessToken)
			assert.NotEmpty(t, tokens.RefreshToken)
			assert.Equal(t, "Bearer", tokens.TokenType)
			assert.Greater(t, tokens.ExpiresIn, int64(0))
		})
	}
}

func TestGenerateAccessToken(t *testing.T) {
	tests := []struct {
		name   string
		userID string
	}{
		{
			name:   "正常なアクセストークン生成",
			userID: "test-user-id",
		},
		{
			name:   "異なるユーザーIDでアクセストークン生成",
			userID: "another-user-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := GenerateAccessToken(tt.userID)

			require.NoError(t, err)
			assert.NotEmpty(t, tokens.AccessToken)
			assert.Empty(t, tokens.RefreshToken) // RefreshTokenは空
			assert.Equal(t, "Bearer", tokens.TokenType)
			assert.Greater(t, tokens.ExpiresIn, int64(0))
		})
	}
}

func TestParseAuthTokens(t *testing.T) {
	userID := "test-user-id"
	tokens, err := GenerateAuthTokens(userID)
	require.NoError(t, err)

	tests := []struct {
		name        string
		tokenString string
		expectError bool
	}{
		{
			name:        "正常なトークンのパース",
			tokenString: tokens.AccessToken,
			expectError: false,
		},
		{
			name:        "無効なトークン",
			tokenString: "invalid-token",
			expectError: true,
		},
		{
			name:        "空のトークン",
			tokenString: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedTokens, parsedUserID, err := ParseAuthTokens(tt.tokenString)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, parsedTokens)
				assert.Empty(t, parsedUserID)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, parsedTokens)
				assert.Equal(t, userID, parsedUserID)
				assert.Equal(t, "Bearer", parsedTokens.TokenType)
			}
		})
	}
}

func TestParseAuthTokens_ExpiredToken(t *testing.T) {
	// 期限切れトークンを生成（過去の時刻で生成）
	jwtSecret := []byte("hogehoge") // テスト用のシークレット

	expiredClaims := &Claims{
		UserID: "test-user-id",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1時間前に期限切れ
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "test-user-id",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := token.SignedString(jwtSecret)
	require.NoError(t, err)

	tests := []struct {
		name        string
		tokenString string
	}{
		{
			name:        "期限切れトークン",
			tokenString: expiredTokenString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedTokens, parsedUserID, err := ParseAuthTokens(tt.tokenString)

			require.Error(t, err)
			assert.Nil(t, parsedTokens)
			assert.Empty(t, parsedUserID)
		})
	}
}

func TestConvertAuthResponse(t *testing.T) {
	tests := []struct {
		name         string
		tokenDetails *TokenDetails
	}{
		{
			name: "正常な変換",
			tokenDetails: &TokenDetails{
				AccessToken:  "test-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "test-refresh-token",
			},
		},
		{
			name: "RefreshTokenなしの変換",
			tokenDetails: &TokenDetails{
				AccessToken:  "test-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResponse := tt.tokenDetails.ConvertAuthResponse()

			assert.Equal(t, tt.tokenDetails.AccessToken, authResponse.AccessToken)
			assert.Equal(t, tt.tokenDetails.TokenType, authResponse.TokenType)
			assert.Equal(t, int32(tt.tokenDetails.ExpiresIn), authResponse.ExpiresIn)
			assert.Equal(t, tt.tokenDetails.RefreshToken, authResponse.RefreshToken)
		})
	}
}
