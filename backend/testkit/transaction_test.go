package testkit

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestTransactionFunctions(t *testing.T) {
	config := testutil.DefaultTestDBConfig()
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.DBName))
	if err != nil {
		t.Skip("Database connection not available, skipping test")
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Failed to close database connection: %v", err)
		}
	}()

	// Test connection
	if err := db.Ping(); err != nil {
		t.Skip("Database ping failed, skipping test")
	}

	ctx := context.Background()

	// RwTransactionのテスト
	t.Run("RwTransaction", func(t *testing.T) {
		err := database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
			// トランザクション内でのテスト処理
			_, err := tx.ExecContext(ctx, "SELECT 1")
			return err
		})
		if err != nil {
			t.Errorf("RwTransaction failed: %v", err)
		}
	})

	// RoTransactionのテスト
	t.Run("RoTransaction", func(t *testing.T) {
		err := database.RoTransaction(ctx, db, func(tx *sql.Tx) error {
			// 読み取り専用トランザクション内でのテスト処理
			_, err := tx.QueryContext(ctx, "SELECT 1")
			return err
		})
		if err != nil {
			t.Errorf("RoTransaction failed: %v", err)
		}
	})

	// エラーハンドリングのテスト
	t.Run("TransactionRollback", func(t *testing.T) {
		err := database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
			// 意図的にエラーを発生させる
			return sql.ErrNoRows
		})
		if err != sql.ErrNoRows {
			t.Errorf("Expected error to be returned, got: %v", err)
		}
	})
}
