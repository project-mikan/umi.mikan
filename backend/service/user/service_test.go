package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUpdateUserName(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	service := &UserEntry{DB: db}

	// テスト用ユーザーを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		Name:      "元の名前",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	// コンテキストにユーザーIDを設定
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	tests := []struct {
		name            string
		newName         string
		expectSuccess   bool
		expectedMessage string
	}{
		{
			name:            "正常なユーザー名更新",
			newName:         "新しい名前",
			expectSuccess:   true,
			expectedMessage: "usernameUpdateSuccess",
		},
		{
			name:            "空のユーザー名",
			newName:         "",
			expectSuccess:   false,
			expectedMessage: "nameRequired",
		},
		{
			name:            "20文字を超えるユーザー名",
			newName:         "これは20文字を超える非常に長いユーザー名です",
			expectSuccess:   false,
			expectedMessage: "nameTooLong",
		},
		{
			name:            "ちょうど20文字のユーザー名",
			newName:         "１２３４５６７８９０１２３４５６７８９０",
			expectSuccess:   true,
			expectedMessage: "usernameUpdateSuccess",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.UpdateUserNameRequest{
				NewName: tt.newName,
			}

			resp, err := service.UpdateUserName(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, resp.Success)
			assert.Equal(t, tt.expectedMessage, resp.Message)
		})
	}
}

func TestChangePassword(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	service := &UserEntry{DB: db}

	// テスト用ユーザーを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     fmt.Sprintf("test-%s@example.com", userID.String()),
		Name:      "テストユーザー",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	// パスワード認証情報を作成
	currentPassword := "currentPass123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	auth := &database.UserPasswordAuthe{
		UserID:         userID,
		PasswordHashed: string(hashedPassword),
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}
	require.NoError(t, auth.Insert(context.Background(), db))

	// コンテキストにユーザーIDを設定
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	tests := []struct {
		name            string
		currentPassword string
		newPassword     string
		expectSuccess   bool
		expectedMessage string
	}{
		{
			name:            "正常なパスワード変更",
			currentPassword: currentPassword,
			newPassword:     "newPassword123",
			expectSuccess:   true,
			expectedMessage: "passwordChangeSuccess",
		},
		{
			name:            "空のパスワード",
			currentPassword: "",
			newPassword:     "",
			expectSuccess:   false,
			expectedMessage: "passwordsRequired",
		},
		{
			name:            "短すぎる新パスワード",
			currentPassword: currentPassword,
			newPassword:     "short",
			expectSuccess:   false,
			expectedMessage: "passwordTooShort",
		},
		{
			name:            "間違った現在のパスワード",
			currentPassword: "wrongPassword",
			newPassword:     "newPassword123",
			expectSuccess:   false,
			expectedMessage: "currentPasswordIncorrect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.ChangePasswordRequest{
				CurrentPassword: tt.currentPassword,
				NewPassword:     tt.newPassword,
			}

			resp, err := service.ChangePassword(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, resp.Success)
			assert.Equal(t, tt.expectedMessage, resp.Message)
		})
	}
}
