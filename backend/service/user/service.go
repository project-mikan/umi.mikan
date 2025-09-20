package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/request"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

type UserEntry struct {
	g.UnimplementedUserServiceServer
	DB database.DB
}

func (s *UserEntry) UpdateUserName(ctx context.Context, req *g.UpdateUserNameRequest) (*g.UpdateUserNameResponse, error) {
	// リクエストのバリデーション
	if req.GetNewName() == "" {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "nameRequired",
		}, nil
	}

	// 名前の長さチェック（20文字以内）
	if len([]rune(req.GetNewName())) > 20 {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "nameTooLong",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	// ユーザーの取得
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	userDB, err := database.UserByID(ctx, s.DB, parsedUserID)
	if err != nil {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "userNotFound",
		}, nil
	}

	// ユーザー名を更新
	userDB.Name = req.GetNewName()
	userDB.UpdatedAt = time.Now().Unix()

	if err := userDB.Update(ctx, s.DB); err != nil {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.UpdateUserNameResponse{
		Success: true,
		Message: "usernameUpdateSuccess",
	}, nil
}

func (s *UserEntry) ChangePassword(ctx context.Context, req *g.ChangePasswordRequest) (*g.ChangePasswordResponse, error) {
	// リクエストのバリデーション
	if req.GetCurrentPassword() == "" || req.GetNewPassword() == "" {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "passwordsRequired",
		}, nil
	}

	// 新しいパスワードの強度チェック
	if len(req.GetNewPassword()) < 8 {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "passwordTooShort",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 現在のパスワードを取得して検証
	passwordAuthDB, err := database.UserPasswordAutheByUserID(ctx, s.DB, parsedUserID)
	if err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "userNotFound",
		}, nil
	}

	// 現在のパスワードの検証
	if err := request.VerifyPassword(req.GetCurrentPassword(), passwordAuthDB.PasswordHashed); err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "currentPasswordIncorrect",
		}, nil
	}

	// 新しいパスワードをハッシュ化
	hashedNewPassword, err := request.EncryptPassword(req.GetNewPassword())
	if err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// パスワードを更新
	passwordAuthDB.PasswordHashed = hashedNewPassword
	passwordAuthDB.UpdatedAt = time.Now().Unix()

	if err := passwordAuthDB.Update(ctx, s.DB); err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.ChangePasswordResponse{
		Success: true,
		Message: "passwordChangeSuccess",
	}, nil
}

func (s *UserEntry) UpdateLLMKey(ctx context.Context, req *g.UpdateLLMKeyRequest) (*g.UpdateLLMKeyResponse, error) {
	// リクエストのバリデーション
	if req.GetKey() == "" {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "tokenRequired",
		}, nil
	}

	// トークンの長さチェック（100文字以内）
	if len(req.GetKey()) > 100 {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "tokenTooLong",
		}, nil
	}

	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "invalidProvider",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 既存のLLMトークンを確認
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	currentTime := time.Now().Unix()

	if err != nil && err != sql.ErrNoRows {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	if err == sql.ErrNoRows {
		// 新規作成
		newUserLLM := &database.UserLlm{
			UserID:             parsedUserID,
			LlmProvider:        int16(req.GetLlmProvider()),
			Key:                req.GetKey(),
			AutoSummaryDaily:   false, // デフォルトは無効
			AutoSummaryMonthly: false, // デフォルトは無効
			CreatedAt:          currentTime,
			UpdatedAt:          currentTime,
		}

		if err := newUserLLM.Insert(ctx, s.DB); err != nil {
			return &g.UpdateLLMKeyResponse{
				Success: false,
				Message: "updateFailed",
			}, nil
		}
	} else {
		// 更新
		userLLMDB.Key = req.GetKey()
		userLLMDB.UpdatedAt = currentTime

		if err := userLLMDB.Update(ctx, s.DB); err != nil {
			return &g.UpdateLLMKeyResponse{
				Success: false,
				Message: "updateFailed",
			}, nil
		}
	}

	return &g.UpdateLLMKeyResponse{
		Success: true,
		Message: "llmTokenUpdateSuccess",
	}, nil
}

func (s *UserEntry) GetUserInfo(ctx context.Context, req *g.GetUserInfoRequest) (*g.GetUserInfoResponse, error) {
	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// ユーザー情報を取得
	userDB, err := database.UserByID(ctx, s.DB, parsedUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// LLMキーを取得（存在する場合）
	var llmKeys []*g.LLMKeyInfo

	// 現在はGemini（provider 1）のみサポート
	userLLM, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, 1)
	if err == nil && userLLM != nil {
		llmKeys = append(llmKeys, &g.LLMKeyInfo{
			LlmProvider:        int32(userLLM.LlmProvider),
			Key:                userLLM.Key,
			AutoSummaryDaily:   userLLM.AutoSummaryDaily,
			AutoSummaryMonthly: userLLM.AutoSummaryMonthly,
		})
	}

	return &g.GetUserInfoResponse{
		Name:    userDB.Name,
		Email:   userDB.Email,
		LlmKeys: llmKeys,
	}, nil
}

func (s *UserEntry) DeleteLLMKey(ctx context.Context, req *g.DeleteLLMKeyRequest) (*g.DeleteLLMKeyResponse, error) {
	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "invalidProvider",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 既存のLLMトークンを取得
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	if err != nil {
		if err == sql.ErrNoRows {
			return &g.DeleteLLMKeyResponse{
				Success: false,
				Message: "tokenNotFound",
			}, nil
		}
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// LLMトークンを削除
	if err := userLLMDB.Delete(ctx, s.DB); err != nil {
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.DeleteLLMKeyResponse{
		Success: true,
		Message: "llmTokenDeleteSuccess",
	}, nil
}

func (s *UserEntry) DeleteAccount(ctx context.Context, req *g.DeleteAccountRequest) (*g.DeleteAccountResponse, error) {
	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.DeleteAccountResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.DeleteAccountResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// ユーザーの存在確認
	userDB, err := database.UserByID(ctx, s.DB, parsedUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &g.DeleteAccountResponse{
				Success: false,
				Message: "userNotFound",
			}, nil
		}
		return &g.DeleteAccountResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// トランザクション内で関連データを削除
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		// 1. 日記データを削除 (個別に削除するためのクエリを実行)
		_, err := tx.ExecContext(ctx, "DELETE FROM diaries WHERE user_id = $1", parsedUserID)
		if err != nil {
			return fmt.Errorf("failed to delete diary entries: %w", err)
		}

		// 2. LLMトークンを削除
		_, err = tx.ExecContext(ctx, "DELETE FROM user_llms WHERE user_id = $1", parsedUserID)
		if err != nil {
			return fmt.Errorf("failed to delete user LLMs: %w", err)
		}

		// 3. パスワード認証を削除
		_, err = tx.ExecContext(ctx, "DELETE FROM user_password_authes WHERE user_id = $1", parsedUserID)
		if err != nil {
			return fmt.Errorf("failed to delete password auth: %w", err)
		}

		// 4. ユーザーを削除
		if err := userDB.Delete(ctx, tx); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})
	if err != nil {
		return &g.DeleteAccountResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.DeleteAccountResponse{
		Success: true,
		Message: "accountDeleteSuccess",
	}, nil
}

func (s *UserEntry) UpdateAutoSummarySettings(ctx context.Context, req *g.UpdateAutoSummarySettingsRequest) (*g.UpdateAutoSummarySettingsResponse, error) {
	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "invalidProvider",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 既存のLLM設定を取得
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	if err != nil {
		if err == sql.ErrNoRows {
			return &g.UpdateAutoSummarySettingsResponse{
				Success: false,
				Message: "llmKeyNotFound",
			}, nil
		}
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// 自動要約設定を更新
	userLLMDB.AutoSummaryDaily = req.GetAutoSummaryDaily()
	userLLMDB.AutoSummaryMonthly = req.GetAutoSummaryMonthly()
	userLLMDB.UpdatedAt = time.Now().Unix()

	if err := userLLMDB.Update(ctx, s.DB); err != nil {
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.UpdateAutoSummarySettingsResponse{
		Success: true,
		Message: "autoSummarySettingsUpdateSuccess",
	}, nil
}

func (s *UserEntry) GetAutoSummarySettings(ctx context.Context, req *g.GetAutoSummarySettingsRequest) (*g.GetAutoSummarySettingsResponse, error) {
	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.GetAutoSummarySettingsResponse{
			AutoSummaryDaily:   false,
			AutoSummaryMonthly: false,
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.GetAutoSummarySettingsResponse{
			AutoSummaryDaily:   false,
			AutoSummaryMonthly: false,
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.GetAutoSummarySettingsResponse{
			AutoSummaryDaily:   false,
			AutoSummaryMonthly: false,
		}, nil
	}

	// LLM設定を取得
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	if err != nil {
		// 設定が存在しない場合はデフォルト値を返す
		return &g.GetAutoSummarySettingsResponse{
			AutoSummaryDaily:   false,
			AutoSummaryMonthly: false,
		}, nil
	}

	return &g.GetAutoSummarySettingsResponse{
		AutoSummaryDaily:   userLLMDB.AutoSummaryDaily,
		AutoSummaryMonthly: userLLMDB.AutoSummaryMonthly,
	}, nil
}
