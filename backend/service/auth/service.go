package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/domain/request"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
)

type AuthEntry struct {
	g.UnimplementedAuthServiceServer
	DB database.DB
}

func (s *AuthEntry) RegisterByPassword(ctx context.Context, req *g.RegisterByPasswordRequest) (*g.AuthResponse, error) {
	passwordAuth, err := request.ValidateRegisterByPasswordRequest(req)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// --- 既存ユーザーの確認 ---
	existingUser, err := database.UserByEmail(ctx, s.DB, passwordAuth.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", passwordAuth.Email)
	}

	// --- 登録 ---
	user := model.GenUser(passwordAuth.Email, passwordAuth.Name, model.AuthTypeEmailPassword)
	// トランザクション内でユーザー作成とパスワード認証を同時に実行
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		userDB := user.ConvertToDBModel()
		if err := userDB.Save(ctx, tx); err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
		passwordAuthDB := passwordAuth.ConvertToDBModel(user.ID)
		if err := passwordAuthDB.Save(ctx, tx); err != nil {
			return fmt.Errorf("failed to insert password auth: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// --- JWTトークンの生成 ---
	token, err := model.GenerateAuthTokens(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth tokens: %w", err)
	}

	return token.ConvertAuthResponse(), nil
}

func (s *AuthEntry) LoginByPassword(ctx context.Context, req *g.LoginByPasswordRequest) (*g.AuthResponse, error) {
	passwordAuth, err := request.ValidateLoginByPasswordRequest(req)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// --- ユーザーの取得 ---
	userDB, err := database.UserByEmail(ctx, s.DB, passwordAuth.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// --- パスワードの検証 ---
	passwordAuthDB, err := database.UserPasswordAutheByUserID(ctx, s.DB, userDB.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get password auth: %w", err)
	}
	// bcryptを使って平文パスワードとハッシュを比較
	if err := request.VerifyPassword(passwordAuth.Password, passwordAuthDB.PasswordHashed); err != nil {
		return nil, fmt.Errorf("password does not match")
	}

	// --- JWTトークンの生成 ---
	token, err := model.GenerateAuthTokens(userDB.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth tokens: %w", err)
	}

	return token.ConvertAuthResponse(), nil
}

func (s *AuthEntry) RefreshAccessToken(ctx context.Context, req *g.RefreshAccessTokenRequest) (*g.AuthResponse, error) {
	userID, err := request.ValidateRefreshTokenRequest(req)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// --- ユーザーの取得 ---
	// tokenから引っ張ってきたUserIDはUUID形式(でないとぶっ壊れて取れないのでMustParseでよい)
	userDB, err := database.UserByID(ctx, s.DB, uuid.MustParse(userID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	if userDB == nil {
		return nil, fmt.Errorf("user not found")
	}

	// --- AccessTokenだけ再生成 ---
	newToken, err := model.GenerateAccessToken(userDB.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate new auth tokens: %w", err)
	}

	return newToken.ConvertAuthResponse(), nil
}

func (s *AuthEntry) UpdateUserName(ctx context.Context, req *g.UpdateUserNameRequest) (*g.UpdateUserNameResponse, error) {
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

func (s *AuthEntry) ChangePassword(ctx context.Context, req *g.ChangePasswordRequest) (*g.ChangePasswordResponse, error) {
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

func (s *AuthEntry) UpdateLLMToken(ctx context.Context, req *g.UpdateLLMTokenRequest) (*g.UpdateLLMTokenResponse, error) {
	// リクエストのバリデーション
	if req.GetToken() == "" {
		return &g.UpdateLLMTokenResponse{
			Success: false,
			Message: "tokenRequired",
		}, nil
	}

	// トークンの長さチェック（100文字以内）
	if len(req.GetToken()) > 100 {
		return &g.UpdateLLMTokenResponse{
			Success: false,
			Message: "tokenTooLong",
		}, nil
	}

	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.UpdateLLMTokenResponse{
			Success: false,
			Message: "invalidProvider",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.UpdateLLMTokenResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.UpdateLLMTokenResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 既存のLLMトークンを確認
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	currentTime := time.Now().Unix()

	if err != nil && err != sql.ErrNoRows {
		return &g.UpdateLLMTokenResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	if err == sql.ErrNoRows {
		// 新規作成
		newUserLLM := &database.UserLlm{
			UserID:      parsedUserID,
			LlmProvider: int16(req.GetLlmProvider()),
			Token:       req.GetToken(),
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
		}

		if err := newUserLLM.Insert(ctx, s.DB); err != nil {
			return &g.UpdateLLMTokenResponse{
				Success: false,
				Message: "updateFailed",
			}, nil
		}
	} else {
		// 更新
		userLLMDB.Token = req.GetToken()
		userLLMDB.UpdatedAt = currentTime

		if err := userLLMDB.Update(ctx, s.DB); err != nil {
			return &g.UpdateLLMTokenResponse{
				Success: false,
				Message: "updateFailed",
			}, nil
		}
	}

	return &g.UpdateLLMTokenResponse{
		Success: true,
		Message: "llmTokenUpdateSuccess",
	}, nil
}

func (s *AuthEntry) GetUserInfo(ctx context.Context, req *g.GetUserInfoRequest) (*g.GetUserInfoResponse, error) {
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

	// LLMトークンを取得（存在する場合）
	var llmTokens []*g.LLMTokenInfo

	// 現在はGemini（provider 0）のみサポート
	userLLM, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, 0)
	if err == nil && userLLM != nil {
		llmTokens = append(llmTokens, &g.LLMTokenInfo{
			LlmProvider: int32(userLLM.LlmProvider),
			Token:       userLLM.Token,
		})
	}

	return &g.GetUserInfoResponse{
		Name:      userDB.Name,
		Email:     userDB.Email,
		LlmTokens: llmTokens,
	}, nil
}

func (s *AuthEntry) DeleteLLMToken(ctx context.Context, req *g.DeleteLLMTokenRequest) (*g.DeleteLLMTokenResponse, error) {
	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.DeleteLLMTokenResponse{
			Success: false,
			Message: "invalidProvider",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.DeleteLLMTokenResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.DeleteLLMTokenResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 既存のLLMトークンを取得
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	if err != nil {
		if err == sql.ErrNoRows {
			return &g.DeleteLLMTokenResponse{
				Success: false,
				Message: "tokenNotFound",
			}, nil
		}
		return &g.DeleteLLMTokenResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// LLMトークンを削除
	if err := userLLMDB.Delete(ctx, s.DB); err != nil {
		return &g.DeleteLLMTokenResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.DeleteLLMTokenResponse{
		Success: true,
		Message: "llmTokenDeleteSuccess",
	}, nil
}

func (s *AuthEntry) DeleteAccount(ctx context.Context, req *g.DeleteAccountRequest) (*g.DeleteAccountResponse, error) {
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
