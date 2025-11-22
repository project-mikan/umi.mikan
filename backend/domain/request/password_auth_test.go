package request

import (
	"testing"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestValidateRegisterByPasswordRequest(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		userName    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "正常な登録リクエスト",
			email:       "test@example.com",
			password:    "password123",
			userName:    "テストユーザー",
			expectError: false,
		},
		{
			name:        "空のメールアドレス",
			email:       "",
			password:    "password123",
			userName:    "テストユーザー",
			expectError: true,
			errorMsg:    "email must not be empty",
		},
		{
			name:        "空のパスワード",
			email:       "test@example.com",
			password:    "",
			userName:    "テストユーザー",
			expectError: true,
			errorMsg:    "password must not be empty",
		},
		{
			name:        "空の名前",
			email:       "test@example.com",
			password:    "password123",
			userName:    "",
			expectError: true,
			errorMsg:    "name must not be empty",
		},
		{
			name:        "無効なメールアドレス形式",
			email:       "invalid-email",
			password:    "password123",
			userName:    "テストユーザー",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "短すぎるパスワード",
			email:       "test@example.com",
			password:    "short",
			userName:    "テストユーザー",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "ちょうど8文字のパスワード",
			email:       "test@example.com",
			password:    "pass1234",
			userName:    "テストユーザー",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.RegisterByPasswordRequest{
				Email:    tt.email,
				Password: tt.password,
				Name:     tt.userName,
			}

			result, err := ValidateRegisterByPasswordRequest(req)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Email)
				assert.Equal(t, tt.userName, result.Name)
				assert.NotEmpty(t, result.PasswordHashed)
				// パスワードがハッシュ化されていることを確認
				err := bcrypt.CompareHashAndPassword([]byte(result.PasswordHashed), []byte(tt.password))
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLoginByPasswordRequest(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "正常なログインリクエスト",
			email:       "test@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "空のメールアドレス",
			email:       "",
			password:    "password123",
			expectError: true,
			errorMsg:    "email and password must not be empty",
		},
		{
			name:        "空のパスワード",
			email:       "test@example.com",
			password:    "",
			expectError: true,
			errorMsg:    "email and password must not be empty",
		},
		{
			name:        "両方とも空",
			email:       "",
			password:    "",
			expectError: true,
			errorMsg:    "email and password must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.LoginByPasswordRequest{
				Email:    tt.email,
				Password: tt.password,
			}

			result, err := ValidateLoginByPasswordRequest(req)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.email, result.Email)
				assert.Equal(t, tt.password, result.Password)
				assert.Empty(t, result.Name) // ログイン時は名前は空
			}
		})
	}
}

func TestValidateRefreshTokenRequest(t *testing.T) {
	// 有効なトークンを生成
	userID := uuid.New().String()
	tokens, err := generateTestTokens(userID)
	require.NoError(t, err)

	tests := []struct {
		name         string
		refreshToken string
		expectError  bool
		errorMsg     string
		expectedID   string
	}{
		{
			name:         "正常なリフレッシュトークン",
			refreshToken: tokens.RefreshToken,
			expectError:  false,
			expectedID:   userID,
		},
		{
			name:         "空のリフレッシュトークン",
			refreshToken: "",
			expectError:  true,
			errorMsg:     "refresh token must not be empty",
		},
		{
			name:         "無効なリフレッシュトークン",
			refreshToken: "invalid-token",
			expectError:  true,
			errorMsg:     "failed to parse refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.RefreshAccessTokenRequest{
				RefreshToken: tt.refreshToken,
			}

			resultUserID, err := ValidateRefreshTokenRequest(req)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Empty(t, resultUserID)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, resultUserID)
			}
		})
	}
}

func TestEncryptPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "通常のパスワード",
			password: "password123",
		},
		{
			name:     "長いパスワード",
			password: "this-is-a-very-long-password-with-many-characters",
		},
		{
			name:     "特殊文字を含むパスワード",
			password: "p@ssw0rd!#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := EncryptPassword(tt.password)

			require.NoError(t, err)
			assert.NotEmpty(t, hashedPassword)
			assert.NotEqual(t, tt.password, hashedPassword)

			// ハッシュ化されたパスワードが元のパスワードと一致することを確認
			err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(tt.password))
			assert.NoError(t, err)
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	correctPassword := "password123"
	hashedPassword, err := EncryptPassword(correctPassword)
	require.NoError(t, err)

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "正しいパスワード",
			password:    correctPassword,
			expectError: false,
		},
		{
			name:        "間違ったパスワード",
			password:    "wrongpassword",
			expectError: true,
		},
		{
			name:        "空のパスワード",
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.password, hashedPassword)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "正常なメールアドレス",
			email:    "test@example.com",
			expected: true,
		},
		{
			name:     "サブドメイン付きメールアドレス",
			email:    "user@mail.example.com",
			expected: true,
		},
		{
			name:     "数字を含むメールアドレス",
			email:    "user123@example.com",
			expected: true,
		},
		{
			name:     "特殊文字を含むメールアドレス",
			email:    "user.name+tag@example.com",
			expected: true,
		},
		{
			name:     "@がないメールアドレス",
			email:    "invalid-email",
			expected: false,
		},
		{
			name:     "ドメインがないメールアドレス",
			email:    "user@",
			expected: false,
		},
		{
			name:     "トップレベルドメインがないメールアドレス",
			email:    "user@example",
			expected: false,
		},
		{
			name:     "空のメールアドレス",
			email:    "",
			expected: false,
		},
		{
			name:     "スペースを含むメールアドレス",
			email:    "user name@example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertToDBModel(t *testing.T) {
	userID := uuid.New()
	passwordAuth := &PasswordAuth{
		Name:           "テストユーザー",
		Email:          "test@example.com",
		PasswordHashed: "hashed-password",
	}

	tests := []struct {
		name   string
		userID uuid.UUID
	}{
		{
			name:   "正常な変換",
			userID: userID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := passwordAuth.ConvertToDBModel(tt.userID)

			assert.Equal(t, tt.userID, result.UserID)
			assert.Equal(t, passwordAuth.PasswordHashed, result.PasswordHashed)
			assert.Greater(t, result.CreatedAt, int64(0))
			assert.Greater(t, result.UpdatedAt, int64(0))
			assert.Equal(t, result.CreatedAt, result.UpdatedAt)
		})
	}
}

// ヘルパー関数: テスト用のトークンを生成
func generateTestTokens(userID string) (*model.TokenDetails, error) {
	// model.GenerateAuthTokensを使用してトークンを生成
	return model.GenerateAuthTokens(userID)
}
