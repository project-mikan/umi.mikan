package testkit

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

// Setup はテスト用のデータベース接続を初期化します
func Setup(t *testing.T) *sql.DB {
	t.Helper()

	config := testutil.DefaultTestDBConfig()
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName,
	))
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

// Teardown はテスト後のクリーンアップを行います
func Teardown(db *sql.DB) {
	if db != nil {
		_ = db.Close()
	}
}
