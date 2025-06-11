package request

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"golang.org/x/crypto/bcrypt"
)

type PasswordAuth struct {
	Name           string
	Email          string
	PasswordHashed string
}

func (u *PasswordAuth) ConvertToDBModel(userID uuid.UUID) database.UserPasswordAuthe {
	return database.UserPasswordAuthe{
		UserID:         userID,
		PasswordHashed: u.PasswordHashed,
	}
}

func ValidateRegisterByPasswordRequest(req *g.RegisterByPasswordRequest) (*PasswordAuth, error) {
	// TODO: 細かくやる
	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetName() == "" {
		return nil, fmt.Errorf("email and password and name must not be empty")
	}
	hashedPassword, err := encryptPassword(req.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return &PasswordAuth{
		Name:           req.GetName(),
		Email:          req.GetEmail(),
		PasswordHashed: hashedPassword,
	}, nil
}

// passworrdHasedを返す
func ValidateLoginByPasswordRequest(req *g.LoginByPasswordRequest) (*PasswordAuth, error) {
	// TODO: 細かくやる
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, fmt.Errorf("email and password must not be empty")
	}

	hashedPassword, err := encryptPassword(req.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return &PasswordAuth{
		Name:           "", // 名前はログイン時には必要ないので空文字
		Email:          req.GetEmail(),
		PasswordHashed: hashedPassword,
	}, nil
}

// ValidateRefreshTokenRequest 第一引数はuserID
func ValidateRefreshTokenRequest(req *g.RefreshAccessTokenRequest) (string, error) {
	if req.GetRefreshToken() == "" {
		return "", fmt.Errorf("refresh token must not be empty")
	}

	// --- トークンの検証 ---
	_, userID, err := model.ParseAuthTokens(req.GetRefreshToken())
	if err != nil {
		return "", fmt.Errorf("failed to parse refresh token: %w", err)
	}

	return userID, nil
}

func encryptPassword(password string) (string, error) {
	// パスワードの文字列をハッシュ化する
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
