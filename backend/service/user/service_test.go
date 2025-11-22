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

func TestUpdateLLMKey(t *testing.T) {
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

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	tests := []struct {
		name            string
		key             string
		llmProvider     int32
		expectSuccess   bool
		expectedMessage string
	}{
		{
			name:            "正常なLLMキー登録",
			key:             "test-api-key-12345",
			llmProvider:     1,
			expectSuccess:   true,
			expectedMessage: "llmTokenUpdateSuccess",
		},
		{
			name:            "空のキー",
			key:             "",
			llmProvider:     1,
			expectSuccess:   false,
			expectedMessage: "tokenRequired",
		},
		{
			name:            "長すぎるキー",
			key:             string(make([]byte, 101)),
			llmProvider:     1,
			expectSuccess:   false,
			expectedMessage: "tokenTooLong",
		},
		{
			name:            "無効なプロバイダー",
			key:             "test-api-key",
			llmProvider:     -1,
			expectSuccess:   false,
			expectedMessage: "invalidProvider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.UpdateLLMKeyRequest{
				Key:         tt.key,
				LlmProvider: tt.llmProvider,
			}

			resp, err := service.UpdateLLMKey(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, resp.Success)
			assert.Equal(t, tt.expectedMessage, resp.Message)
		})
	}
}

func TestGetUserInfo(t *testing.T) {
	db := testkit.Setup(t)
	defer testkit.Teardown(db)

	service := &UserEntry{DB: db}

	// テスト用ユーザーを作成
	userID := uuid.New()
	user := &database.User{
		ID:        userID,
		Email:     "test@example.com",
		Name:      "テストユーザー",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, user.Insert(context.Background(), db))

	// LLMキーを作成
	llm := &database.UserLlm{
		UserID:             userID,
		LlmProvider:        1,
		Key:                "test-key",
		AutoSummaryDaily:   true,
		AutoSummaryMonthly: false,
		CreatedAt:          time.Now().Unix(),
		UpdatedAt:          time.Now().Unix(),
	}
	require.NoError(t, llm.Insert(context.Background(), db))

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	tests := []struct {
		name string
	}{
		{
			name: "正常なユーザー情報取得",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.GetUserInfoRequest{}

			resp, err := service.GetUserInfo(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, user.Name, resp.Name)
			assert.Equal(t, user.Email, resp.Email)
			assert.Len(t, resp.LlmKeys, 1)
			assert.Equal(t, int32(1), resp.LlmKeys[0].LlmProvider)
			assert.True(t, resp.LlmKeys[0].AutoSummaryDaily)
			assert.False(t, resp.LlmKeys[0].AutoSummaryMonthly)
		})
	}
}

func TestDeleteLLMKey(t *testing.T) {
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

	// LLMキーを作成
	llm := &database.UserLlm{
		UserID:      userID,
		LlmProvider: 1,
		Key:         "test-key",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	require.NoError(t, llm.Insert(context.Background(), db))

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	tests := []struct {
		name            string
		llmProvider     int32
		expectSuccess   bool
		expectedMessage string
	}{
		{
			name:            "正常なLLMキー削除",
			llmProvider:     1,
			expectSuccess:   true,
			expectedMessage: "llmTokenDeleteSuccess",
		},
		{
			name:            "存在しないキーの削除",
			llmProvider:     2,
			expectSuccess:   false,
			expectedMessage: "tokenNotFound",
		},
		{
			name:            "無効なプロバイダー",
			llmProvider:     -1,
			expectSuccess:   false,
			expectedMessage: "invalidProvider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.DeleteLLMKeyRequest{
				LlmProvider: tt.llmProvider,
			}

			resp, err := service.DeleteLLMKey(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, resp.Success)
			assert.Equal(t, tt.expectedMessage, resp.Message)
		})
	}
}

func TestUpdateAutoSummarySettings(t *testing.T) {
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

	// LLMキーを作成
	llm := &database.UserLlm{
		UserID:      userID,
		LlmProvider: 1,
		Key:         "test-key",
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	require.NoError(t, llm.Insert(context.Background(), db))

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	tests := []struct {
		name                   string
		llmProvider            int32
		autoSummaryDaily       bool
		autoSummaryMonthly     bool
		autoLatestTrendEnabled bool
		expectSuccess          bool
		expectedMessage        string
	}{
		{
			name:                   "正常な設定更新",
			llmProvider:            1,
			autoSummaryDaily:       true,
			autoSummaryMonthly:     true,
			autoLatestTrendEnabled: true,
			expectSuccess:          true,
			expectedMessage:        "autoSummarySettingsUpdateSuccess",
		},
		{
			name:                   "無効なプロバイダー",
			llmProvider:            -1,
			autoSummaryDaily:       true,
			autoSummaryMonthly:     false,
			autoLatestTrendEnabled: false,
			expectSuccess:          false,
			expectedMessage:        "invalidProvider",
		},
		{
			name:                   "存在しないLLMキー",
			llmProvider:            2,
			autoSummaryDaily:       true,
			autoSummaryMonthly:     false,
			autoLatestTrendEnabled: false,
			expectSuccess:          false,
			expectedMessage:        "llmKeyNotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.UpdateAutoSummarySettingsRequest{
				LlmProvider:            tt.llmProvider,
				AutoSummaryDaily:       tt.autoSummaryDaily,
				AutoSummaryMonthly:     tt.autoSummaryMonthly,
				AutoLatestTrendEnabled: tt.autoLatestTrendEnabled,
			}

			resp, err := service.UpdateAutoSummarySettings(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, resp.Success)
			assert.Equal(t, tt.expectedMessage, resp.Message)
		})
	}
}

func TestGetAutoSummarySettings(t *testing.T) {
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

	// LLMキーを作成
	llm := &database.UserLlm{
		UserID:                 userID,
		LlmProvider:            1,
		Key:                    "test-key",
		AutoSummaryDaily:       true,
		AutoSummaryMonthly:     false,
		AutoLatestTrendEnabled: true,
		CreatedAt:              time.Now().Unix(),
		UpdatedAt:              time.Now().Unix(),
	}
	require.NoError(t, llm.Insert(context.Background(), db))

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID.String())

	tests := []struct {
		name                        string
		llmProvider                 int32
		expectedAutoDaily           bool
		expectedAutoMonthly         bool
		expectedAutoLatestTrend     bool
	}{
		{
			name:                        "正常な設定取得",
			llmProvider:                 1,
			expectedAutoDaily:           true,
			expectedAutoMonthly:         false,
			expectedAutoLatestTrend:     true,
		},
		{
			name:                        "存在しない設定（デフォルト値）",
			llmProvider:                 2,
			expectedAutoDaily:           false,
			expectedAutoMonthly:         false,
			expectedAutoLatestTrend:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &g.GetAutoSummarySettingsRequest{
				LlmProvider: tt.llmProvider,
			}

			resp, err := service.GetAutoSummarySettings(ctx, req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedAutoDaily, resp.AutoSummaryDaily)
			assert.Equal(t, tt.expectedAutoMonthly, resp.AutoSummaryMonthly)
			assert.Equal(t, tt.expectedAutoLatestTrend, resp.AutoLatestTrendEnabled)
		})
	}
}
