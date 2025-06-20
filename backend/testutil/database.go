package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

// TestDBConfig holds database configuration for testing
type TestDBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// DefaultTestDBConfig returns default database configuration for testing
func DefaultTestDBConfig() TestDBConfig {
	return TestDBConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "test-pass"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "umi_mikan_test"),
	}
}

// SetupTestDB creates a test database connection and cleans up test data
func SetupTestDB(t *testing.T) *sql.DB {
	config := DefaultTestDBConfig()

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Database connection not available, skipping test: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Skipf("Database ping failed, skipping test: %v", err)
	}

	// Don't clean up test data at start to avoid interfering with other tests
	// cleanupTestData(t, db)

	// Schedule cleanup at the end of the test using t.Cleanup
	t.Cleanup(func() {
		// Only clean data at the end, not at the start
		cleanupTestData(t, db)
		_ = db.Close()
	})

	return db
}

// cleanupTestData removes test data from the database
func cleanupTestData(t *testing.T, db *sql.DB) {
	// Use transaction to ensure cleanup order is maintained
	tx, err := db.Begin()
	if err != nil {
		t.Logf("Warning: failed to begin cleanup transaction: %v", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Only clean up data from tests that have definitely completed
	// Use test name and process ID to avoid cleaning up data from running tests
	testPrefix := fmt.Sprintf("%s-%d-", t.Name(), os.Getpid())
	testPrefix = strings.ReplaceAll(testPrefix, "/", "-")
	testPrefix = strings.ReplaceAll(testPrefix, " ", "-")

	cleanupQueries := []string{
		fmt.Sprintf("DELETE FROM diaries WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%%test-suite-%s%%')", testPrefix),
		fmt.Sprintf("DELETE FROM user_password_authes WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%%test-suite-%s%%')", testPrefix),
		fmt.Sprintf("DELETE FROM users WHERE email LIKE '%%test-suite-%s%%'", testPrefix),
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

// SetupTestDBForSuite creates a test database connection specifically for test suites
func SetupTestDBForSuite(t *testing.T) *sql.DB {
	config := DefaultTestDBConfig()

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Database connection not available, skipping test: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Skipf("Database ping failed, skipping test: %v", err)
	}

	// No automatic cleanup for suite - let suite manage its own data
	return db
}

// CleanupTestDB cleans up test data and closes the database connection
func CleanupTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		cleanupTestData(t, db)
		_ = db.Close()
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
