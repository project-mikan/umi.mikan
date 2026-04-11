package user

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/testutil"
	"github.com/redis/rueidis"
)

func setupUserTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestUserEntry_UpdateUserName(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-update-name@example.com", "Old Name")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	tests := []struct {
		name            string
		newName         string
		expectedSuccess bool
		expectedMessage string
	}{
		{
			name:            "正常系：名前を更新",
			newName:         "New Name",
			expectedSuccess: true,
			expectedMessage: "usernameUpdateSuccess",
		},
		{
			name:            "異常系：空の名前",
			newName:         "",
			expectedSuccess: false,
			expectedMessage: "nameRequired",
		},
		{
			name:            "異常系：21文字の名前（上限超過）",
			newName:         "あいうえおかきくけこさしすせそたちつてとな",
			expectedSuccess: false,
			expectedMessage: "nameTooLong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.UpdateUserName(ctx, &g.UpdateUserNameRequest{NewName: tt.newName})
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Success != tt.expectedSuccess {
				t.Errorf("Success: got %v, want %v", resp.Success, tt.expectedSuccess)
			}
			if resp.Message != tt.expectedMessage {
				t.Errorf("Message: got %q, want %q", resp.Message, tt.expectedMessage)
			}
		})
	}
}

func TestUserEntry_UpdateUserName_Unauthenticated(t *testing.T) {
	db := setupUserTestDB(t)
	svc := &UserEntry{DB: db}

	resp, err := svc.UpdateUserName(context.Background(), &g.UpdateUserNameRequest{NewName: "New Name"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if resp.Success {
		t.Error("認証なしでSuccessがtrueになっている")
	}
	if resp.Message != "unauthorized" {
		t.Errorf("Message: got %q, want %q", resp.Message, "unauthorized")
	}
}

func TestUserEntry_ChangePassword(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUserWithPassword(t, db, "user-change-pass@example.com", "Change Pass User", "oldPassword123")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	tests := []struct {
		name            string
		currentPassword string
		newPassword     string
		expectedSuccess bool
		expectedMessage string
	}{
		{
			name:            "異常系：空のパスワード",
			currentPassword: "",
			newPassword:     "",
			expectedSuccess: false,
			expectedMessage: "passwordsRequired",
		},
		{
			name:            "異常系：新パスワードが短すぎる",
			currentPassword: "oldPassword123",
			newPassword:     "short",
			expectedSuccess: false,
			expectedMessage: "passwordTooShort",
		},
		{
			name:            "異常系：現在のパスワードが間違っている",
			currentPassword: "wrongPassword123",
			newPassword:     "newPassword123",
			expectedSuccess: false,
			expectedMessage: "currentPasswordIncorrect",
		},
		{
			name:            "正常系：パスワードを変更",
			currentPassword: "oldPassword123",
			newPassword:     "newPassword123",
			expectedSuccess: true,
			expectedMessage: "passwordChangeSuccess",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.ChangePassword(ctx, &g.ChangePasswordRequest{
				CurrentPassword: tt.currentPassword,
				NewPassword:     tt.newPassword,
			})
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Success != tt.expectedSuccess {
				t.Errorf("Success: got %v, want %v", resp.Success, tt.expectedSuccess)
			}
			if resp.Message != tt.expectedMessage {
				t.Errorf("Message: got %q, want %q", resp.Message, tt.expectedMessage)
			}
		})
	}
}

func TestUserEntry_ChangePassword_Unauthenticated(t *testing.T) {
	db := setupUserTestDB(t)
	svc := &UserEntry{DB: db}

	resp, err := svc.ChangePassword(context.Background(), &g.ChangePasswordRequest{
		CurrentPassword: "old",
		NewPassword:     "newPassword123",
	})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if resp.Success {
		t.Error("認証なしでSuccessがtrueになっている")
	}
	if resp.Message != "unauthorized" {
		t.Errorf("Message: got %q, want %q", resp.Message, "unauthorized")
	}
}

func TestUserEntry_UpdateLLMKey(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-llm-key@example.com", "LLM Key User")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	tests := []struct {
		name            string
		key             string
		llmProvider     int32
		expectedSuccess bool
		expectedMessage string
	}{
		{
			name:            "異常系：空のキー",
			key:             "",
			llmProvider:     1,
			expectedSuccess: false,
			expectedMessage: "tokenRequired",
		},
		{
			name:            "正常系：新規LLMキーを作成",
			key:             "test-api-key-12345",
			llmProvider:     1,
			expectedSuccess: true,
			expectedMessage: "llmTokenUpdateSuccess",
		},
		{
			name:            "正常系：既存LLMキーを更新",
			key:             "updated-api-key-12345",
			llmProvider:     1,
			expectedSuccess: true,
			expectedMessage: "llmTokenUpdateSuccess",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.UpdateLLMKey(ctx, &g.UpdateLLMKeyRequest{
				Key:         tt.key,
				LlmProvider: tt.llmProvider,
			})
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp.Success != tt.expectedSuccess {
				t.Errorf("Success: got %v, want %v", resp.Success, tt.expectedSuccess)
			}
			if resp.Message != tt.expectedMessage {
				t.Errorf("Message: got %q, want %q", resp.Message, tt.expectedMessage)
			}
		})
	}
}

func TestUserEntry_UpdateLLMKey_TokenTooLong(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-llm-long@example.com", "LLM Long User")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	// 101文字のキー
	var longKey strings.Builder
	longKey.WriteString("a")
	for range 100 {
		longKey.WriteString("a")
	}

	resp, err := svc.UpdateLLMKey(ctx, &g.UpdateLLMKeyRequest{
		Key:         longKey.String(),
		LlmProvider: 1,
	})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if resp.Success {
		t.Error("長すぎるキーでSuccessがtrueになっている")
	}
	if resp.Message != "tokenTooLong" {
		t.Errorf("Message: got %q, want %q", resp.Message, "tokenTooLong")
	}
}

func TestUserEntry_GetUserInfo(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-info@example.com", "Info User")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	t.Run("正常系：ユーザー情報を取得", func(t *testing.T) {
		resp, err := svc.GetUserInfo(ctx, &g.GetUserInfoRequest{})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if resp.Name != "Info User" {
			t.Errorf("Name: got %q, want %q", resp.Name, "Info User")
		}
	})

	t.Run("正常系：LLMキー付きユーザー情報を取得", func(t *testing.T) {
		testutil.CreateTestUserLLM(t, db, userID, "test-api-key")

		resp, err := svc.GetUserInfo(ctx, &g.GetUserInfoRequest{})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(resp.LlmKeys) == 0 {
			t.Error("LLMキーが存在するはずがLlmKeysが空")
		}
	})

	t.Run("異常系：未認証でユーザー情報を取得", func(t *testing.T) {
		_, err := svc.GetUserInfo(context.Background(), &g.GetUserInfoRequest{})
		if err == nil {
			t.Error("認証なしでエラーが返らなかった")
		}
	})
}

func TestUserEntry_DeleteLLMKey(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-del-llm@example.com", "Del LLM User")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	t.Run("異常系：存在しないLLMキーを削除", func(t *testing.T) {
		resp, err := svc.DeleteLLMKey(ctx, &g.DeleteLLMKeyRequest{LlmProvider: 1})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if resp.Success {
			t.Error("存在しないキーの削除でSuccessがtrueになっている")
		}
		if resp.Message != "tokenNotFound" {
			t.Errorf("Message: got %q, want %q", resp.Message, "tokenNotFound")
		}
	})

	t.Run("正常系：LLMキーを削除", func(t *testing.T) {
		testutil.CreateTestUserLLM(t, db, userID, "test-api-key")

		resp, err := svc.DeleteLLMKey(ctx, &g.DeleteLLMKeyRequest{LlmProvider: 1})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if !resp.Success {
			t.Errorf("Success: got false, want true (message: %s)", resp.Message)
		}
		if resp.Message != "llmTokenDeleteSuccess" {
			t.Errorf("Message: got %q, want %q", resp.Message, "llmTokenDeleteSuccess")
		}
	})
}

func TestUserEntry_DeleteAccount(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUserWithPassword(t, db, "user-delete-account@example.com", "Delete User", "password123")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	resp, err := svc.DeleteAccount(ctx, &g.DeleteAccountRequest{})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if !resp.Success {
		t.Errorf("Success: got false, want true (message: %s)", resp.Message)
	}
	if resp.Message != "accountDeleteSuccess" {
		t.Errorf("Message: got %q, want %q", resp.Message, "accountDeleteSuccess")
	}
}

func TestUserEntry_DeleteAccount_Unauthenticated(t *testing.T) {
	db := setupUserTestDB(t)
	svc := &UserEntry{DB: db}

	resp, err := svc.DeleteAccount(context.Background(), &g.DeleteAccountRequest{})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if resp.Success {
		t.Error("認証なしでSuccessがtrueになっている")
	}
}

func TestUserEntry_UpdateAutoSummarySettings(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-auto-summary@example.com", "Auto Summary User")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	t.Run("異常系：LLMキーが存在しない", func(t *testing.T) {
		resp, err := svc.UpdateAutoSummarySettings(ctx, &g.UpdateAutoSummarySettingsRequest{
			LlmProvider:        1,
			AutoSummaryMonthly: true,
		})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if resp.Success {
			t.Error("LLMキーが存在しないのにSuccessがtrueになっている")
		}
		if resp.Message != "llmKeyNotFound" {
			t.Errorf("Message: got %q, want %q", resp.Message, "llmKeyNotFound")
		}
	})

	t.Run("正常系：自動要約設定を更新", func(t *testing.T) {
		testutil.CreateTestUserLLM(t, db, userID, "test-api-key")

		resp, err := svc.UpdateAutoSummarySettings(ctx, &g.UpdateAutoSummarySettingsRequest{
			LlmProvider:           1,
			AutoSummaryMonthly:    false,
			SemanticSearchEnabled: true,
		})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if !resp.Success {
			t.Errorf("Success: got false, want true (message: %s)", resp.Message)
		}
		if resp.Message != "autoSummarySettingsUpdateSuccess" {
			t.Errorf("Message: got %q, want %q", resp.Message, "autoSummarySettingsUpdateSuccess")
		}
	})
}

func TestUserEntry_GetAutoSummarySettings(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-get-auto@example.com", "Get Auto User")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	t.Run("正常系：LLMキーが存在しない場合はデフォルト値を返す", func(t *testing.T) {
		resp, err := svc.GetAutoSummarySettings(ctx, &g.GetAutoSummarySettingsRequest{LlmProvider: 1})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if resp.AutoSummaryMonthly {
			t.Error("LLMキーが存在しない場合はデフォルト値がfalseであるべき")
		}
	})

	t.Run("正常系：LLMキーが存在する場合は設定を返す", func(t *testing.T) {
		testutil.CreateTestUserLLM(t, db, userID, "test-api-key")

		resp, err := svc.GetAutoSummarySettings(ctx, &g.GetAutoSummarySettingsRequest{LlmProvider: 1})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		// CreateTestUserLLMはauto_summary_monthly=trueで設定する
		if !resp.AutoSummaryMonthly {
			t.Error("AutoSummaryMonthlyがtrueであるべき")
		}
	})
}

// setupTestRedis はテスト用のminiredisクライアントを起動してrueidisクライアントを返す
func setupTestRedis(t *testing.T) rueidis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis起動失敗: %v", err)
	}
	t.Cleanup(mr.Close)

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("rueidisクライアント作成失敗: %v", err)
	}
	t.Cleanup(client.Close)
	return client
}

func TestUserEntry_GetHourlyMetrics(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-hourly-metrics@example.com", "Metrics User")
	redisClient := setupTestRedis(t)
	svc := &UserEntry{DB: db, RedisClient: redisClient}
	ctx := context.Background()

	t.Run("正常系：時間別メトリクスを取得", func(t *testing.T) {
		metrics, err := svc.getHourlyMetrics(ctx, userID)
		if err != nil {
			t.Fatalf("getHourlyMetrics失敗: %v", err)
		}
		// 過去24時間分のデータが返る
		if len(metrics) != 24 {
			t.Errorf("期待 24件, 実際 %d件", len(metrics))
		}
	})
}

func TestUserEntry_GetMetricsSummary(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-metrics-summary@example.com", "Metrics Summary User")
	redisClient := setupTestRedis(t)
	svc := &UserEntry{DB: db, RedisClient: redisClient}
	ctx := context.Background()

	t.Run("正常系：メトリクスサマリーを取得（データなし）", func(t *testing.T) {
		summary, err := svc.getMetricsSummary(ctx, userID)
		if err != nil {
			t.Fatalf("getMetricsSummary失敗: %v", err)
		}
		if summary.TotalMonthlySummaries != 0 {
			t.Errorf("TotalMonthlySummaries: 期待 0, 実際 %d", summary.TotalMonthlySummaries)
		}
	})
}

func TestUserEntry_GetPubSubMetrics(t *testing.T) {
	db := setupUserTestDB(t)
	userID := testutil.CreateTestUser(t, db, "user-pubsub-metrics@example.com", "PubSub Metrics User")
	redisClient := setupTestRedis(t)
	svc := &UserEntry{DB: db, RedisClient: redisClient}
	ctx := testutil.CreateAuthenticatedContext(userID)

	t.Run("正常系：PubSubメトリクスを取得", func(t *testing.T) {
		resp, err := svc.GetPubSubMetrics(ctx, &g.GetPubSubMetricsRequest{})
		if err != nil {
			t.Fatalf("GetPubSubMetrics失敗: %v", err)
		}
		if resp == nil {
			t.Fatal("レスポンスがnil")
		}
		if len(resp.HourlyMetrics) != 24 {
			t.Errorf("HourlyMetrics: 期待 24件, 実際 %d件", len(resp.HourlyMetrics))
		}
	})
}
