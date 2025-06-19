package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"golang.org/x/crypto/bcrypt"
)

// CreateTestUser creates a test user in the database and returns the user ID
func CreateTestUser(t *testing.T, db *sql.DB, email, name string) uuid.UUID {
	userID := uuid.New()
	currentTime := time.Now().Unix()
	
	// Make email unique by adding test information
	uniqueEmail := generateUniqueEmail(t, email)
	
	// Create test user
	_, err := db.Exec(
		"INSERT INTO users (id, email, name, auth_type, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		userID, uniqueEmail, name, model.AuthTypeEmailPassword.Int16(), currentTime, currentTime,
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	return userID
}

// CreateTestUserWithPassword creates a test user with password authentication
func CreateTestUserWithPassword(t *testing.T, db *sql.DB, email, name, password string) uuid.UUID {
	userID := uuid.New()
	currentTime := time.Now().Unix()
	
	// Make email unique by adding test information
	uniqueEmail := generateUniqueEmail(t, email)
	
	// Hash the password
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	hashedPassword := string(hashedPasswordBytes)
	
	// Use transaction to ensure data consistency
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()
	
	// Create test user
	_, err = tx.Exec(
		"INSERT INTO users (id, email, name, auth_type, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		userID, uniqueEmail, name, model.AuthTypeEmailPassword.Int16(), currentTime, currentTime,
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Create password auth record
	_, err = tx.Exec(
		"INSERT INTO user_password_authes (user_id, password_hashed, created_at, updated_at) VALUES ($1, $2, $3, $4)",
		userID, hashedPassword, currentTime, currentTime,
	)
	if err != nil {
		t.Fatalf("Failed to create password auth: %v", err)
	}
	
	// Commit transaction
	if err = tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
	
	return userID
}

// CreateAuthenticatedContext creates a context with user authentication
func CreateAuthenticatedContext(userID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), "userID", userID.String())
}

// CreateUnauthenticatedContext creates a context without authentication
func CreateUnauthenticatedContext() context.Context {
	return context.Background()
}

// GenerateTestTokens generates access and refresh tokens for a user
func GenerateTestTokens(t *testing.T, userID uuid.UUID) *model.TokenDetails {
	tokens, err := model.GenerateAuthTokens(userID.String())
	if err != nil {
		t.Fatalf("Failed to generate test tokens: %v", err)
	}
	return tokens
}

// generateUniqueEmail creates a unique email address for testing
func generateUniqueEmail(t *testing.T, baseEmail string) string {
	testID := fmt.Sprintf("%s-%d-%d", t.Name(), os.Getpid(), time.Now().UnixNano())
	testID = strings.ReplaceAll(testID, "/", "-")
	testID = strings.ReplaceAll(testID, " ", "-")
	
	// Split email into local and domain parts
	parts := strings.Split(baseEmail, "@")
	if len(parts) == 2 {
		return fmt.Sprintf("%s-%s@%s", parts[0], testID, parts[1])
	}
	// If no @ symbol, just append the test ID
	return fmt.Sprintf("%s-%s", baseEmail, testID)
}