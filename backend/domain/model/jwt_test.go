package model

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func TestAuthType_Int16(t *testing.T) {
	tests := []struct {
		name     string
		authType AuthType
		expected int16
	}{
		{
			name:     "正常系：AuthTypeEmailPassword",
			authType: AuthTypeEmailPassword,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.authType.Int16()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAuthTypeFromInt16(t *testing.T) {
	tests := []struct {
		name     string
		authType int16
		expected AuthType
	}{
		{
			name:     "正常系：AuthTypeEmailPassword（0）",
			authType: 0,
			expected: AuthTypeEmailPassword,
		},
		{
			name:     "正常系：未知の値（デフォルト）",
			authType: 99,
			expected: AuthTypeEmailPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAuthTypeFromInt16(tt.authType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenUser(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		userName string
		authType AuthType
	}{
		{
			name:     "正常系：ユーザー生成",
			email:    "test@example.com",
			userName: "テストユーザー",
			authType: AuthTypeEmailPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := GenUser(tt.email, tt.userName, tt.authType)

			require.NotNil(t, user)
			assert.NotEqual(t, uuid.Nil, user.ID)
			assert.Equal(t, tt.email, user.Email)
			assert.Equal(t, tt.userName, user.Name)
			assert.Equal(t, tt.authType, user.AuthType)
		})
	}
}

func TestUser_ConvertToDBModel(t *testing.T) {
	tests := []struct {
		name string
		user *User
	}{
		{
			name: "正常系：DBモデルへの変換",
			user: &User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Name:     "テストユーザー",
				AuthType: AuthTypeEmailPassword,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbModel := tt.user.ConvertToDBModel()

			assert.Equal(t, tt.user.ID, dbModel.ID)
			assert.Equal(t, tt.user.Email, dbModel.Email)
			assert.Equal(t, tt.user.Name, dbModel.Name)
			assert.Equal(t, tt.user.AuthType.Int16(), dbModel.AuthType)
			assert.Greater(t, dbModel.CreatedAt, int64(0))
			assert.Greater(t, dbModel.UpdatedAt, int64(0))
			assert.Equal(t, dbModel.CreatedAt, dbModel.UpdatedAt)
		})
	}
}
