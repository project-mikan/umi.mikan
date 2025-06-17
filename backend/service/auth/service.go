package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/domain/request"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
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
