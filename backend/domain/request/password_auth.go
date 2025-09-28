package request

import (
	"fmt"
	"regexp"
	"time"

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
	Password       string // ログイン時の平文パスワード(DBには保存しない)
}

func (u *PasswordAuth) ConvertToDBModel(userID uuid.UUID) database.UserPasswordAuthe {
	currentTime := time.Now().Unix()
	return database.UserPasswordAuthe{
		UserID:         userID,
		PasswordHashed: u.PasswordHashed,
		CreatedAt:      currentTime,
		UpdatedAt:      currentTime,
	}
}

func ValidateRegisterByPasswordRequest(req *g.RegisterByPasswordRequest) (*PasswordAuth, error) {
	// Check if required fields are not empty
	if req.GetEmail() == "" {
		return nil, fmt.Errorf("email must not be empty")
	}
	if req.GetPassword() == "" {
		return nil, fmt.Errorf("password must not be empty")
	}
	if req.GetName() == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	// Validate email format
	if !isValidEmail(req.GetEmail()) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Validate password strength
	if len(req.GetPassword()) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters long")
	}

	hashedPassword, err := EncryptPassword(req.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return &PasswordAuth{
		Name:           req.GetName(),
		Email:          req.GetEmail(),
		PasswordHashed: hashedPassword,
	}, nil
}

// ログイン時は平文パスワードを返す
func ValidateLoginByPasswordRequest(req *g.LoginByPasswordRequest) (*PasswordAuth, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, fmt.Errorf("email and password must not be empty")
	}

	return &PasswordAuth{
		Name:     "", // 名前はログイン時には必要ないので空文字
		Email:    req.GetEmail(),
		Password: req.GetPassword(), // 平文パスワードを保持
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

func EncryptPassword(password string) (string, error) {
	// パスワードの文字列をハッシュ化する
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword パスワードとハッシュを比較して検証する
func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
