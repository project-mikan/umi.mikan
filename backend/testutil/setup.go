package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"golang.org/x/crypto/bcrypt"
)

// TestSuite manages test setup and teardown
type TestSuite struct {
	DB     *sql.DB
	UserID uuid.UUID
	t      *testing.T
}

// SetupTestSuite creates a complete test environment with test data
func SetupTestSuite(t *testing.T) *TestSuite {
	db := SetupTestDBForSuite(t)
	
	// Don't clean all test data - only clean conflicting data if needed
	// cleanupAllTestData(t, db)
	
	// Create a test user for the suite
	userID := createTestUserForSuite(t, db)
	
	suite := &TestSuite{
		DB:     db,
		UserID: userID,
		t:      t,
	}
	
	// Schedule cleanup when test finishes
	t.Cleanup(func() {
		cleanupTestSuiteData(t, db, userID)
		db.Close()
	})
	
	return suite
}


// GetAuthenticatedContext returns a context with the test user authenticated
func (ts *TestSuite) GetAuthenticatedContext() context.Context {
	return CreateAuthenticatedContext(ts.UserID)
}

// CreateTestData creates additional test data for specific tests
func (ts *TestSuite) CreateTestData() *TestData {
	return &TestData{
		suite: ts,
		Users: []uuid.UUID{ts.UserID},
	}
}

// TestData holds test data created for specific tests
type TestData struct {
	suite *TestSuite
	Users []uuid.UUID
}

// AddTestUser adds another test user to the test data
func (td *TestData) AddTestUser(email, name, password string) uuid.UUID {
	userID := CreateTestUserWithPassword(td.suite.t, td.suite.DB, email, name, password)
	td.Users = append(td.Users, userID)
	return userID
}

// Cleanup removes all test data created by this TestData instance
func (td *TestData) Cleanup() {
	for _, userID := range td.Users[1:] { // Skip the first user (suite user)
		cleanupTestSuiteData(td.suite.t, td.suite.DB, userID)
	}
}

// createTestUserForSuite creates a test user for the entire test suite
func createTestUserForSuite(t *testing.T, db *sql.DB) uuid.UUID {
	userID := uuid.New()
	currentTime := time.Now().Unix()
	
	// Generate unique email for this test run using test name and timestamp
	testRunID := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())
	email := fmt.Sprintf("test-suite-%s@example.com", testRunID)
	
	// Hash password first
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte("testPassword123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password for test suite user: %v", err)
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
		userID, email, "Test Suite User", model.AuthTypeEmailPassword.Int16(), currentTime, currentTime,
	)
	if err != nil {
		t.Fatalf("Failed to create test suite user: %v", err)
	}
	
	// Create password auth
	_, err = tx.Exec(
		"INSERT INTO user_password_authes (user_id, password_hashed, created_at, updated_at) VALUES ($1, $2, $3, $4)",
		userID, hashedPassword, currentTime, currentTime,
	)
	if err != nil {
		t.Fatalf("Failed to create password auth for test suite user: %v", err)
	}
	
	// Commit transaction
	if err = tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
	
	return userID
}

// cleanupTestSuiteData removes all data for a specific user
func cleanupTestSuiteData(t *testing.T, db *sql.DB, userID uuid.UUID) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Warning: failed to begin cleanup transaction: %v", err)
		return
	}
	defer tx.Rollback()
	
	cleanupQueries := []string{
		"DELETE FROM diaries WHERE user_id = $1",
		"DELETE FROM user_password_authes WHERE user_id = $1",
		"DELETE FROM users WHERE id = $1",
	}
	
	for _, query := range cleanupQueries {
		if _, err := tx.Exec(query, userID); err != nil {
			log.Printf("Warning: cleanup query failed: %v", err)
		}
	}
	
	if err := tx.Commit(); err != nil {
		log.Printf("Warning: failed to commit cleanup transaction: %v", err)
	}
}

// cleanupAllTestData removes all test data from the database
func cleanupAllTestData(t *testing.T, db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		t.Logf("Warning: failed to begin cleanup transaction: %v", err)
		return
	}
	defer tx.Rollback()
	
	cleanupQueries := []string{
		"DELETE FROM diaries WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%test%' OR email LIKE '%suite%')",
		"DELETE FROM user_password_authes WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%test%' OR email LIKE '%suite%')",
		"DELETE FROM users WHERE email LIKE '%test%' OR email LIKE '%suite%'",
	}
	
	for _, query := range cleanupQueries {
		if _, err := tx.Exec(query); err != nil {
			t.Logf("Warning: cleanup query failed: %v", err)
		}
	}
	
	if err := tx.Commit(); err != nil {
		t.Logf("Warning: failed to commit cleanup transaction: %v", err)
	}
}

// TestRunner provides utilities for running tests with proper setup/teardown
type TestRunner struct {
	suite *TestSuite
}

// NewTestRunner creates a new test runner with setup
func NewTestRunner(t *testing.T) *TestRunner {
	return &TestRunner{
		suite: SetupTestSuite(t),
	}
}

// Run executes a test function with automatic cleanup
func (tr *TestRunner) Run(testFunc func(*TestSuite)) {
	testFunc(tr.suite)
}

// RunWithData executes a test function with additional test data and automatic cleanup
func (tr *TestRunner) RunWithData(testFunc func(*TestSuite, *TestData)) {
	testData := tr.suite.CreateTestData()
	// Schedule cleanup of additional test data
	tr.suite.t.Cleanup(func() {
		testData.Cleanup()
	})
	testFunc(tr.suite, testData)
}