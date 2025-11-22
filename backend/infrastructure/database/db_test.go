package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

// TestRoTransaction は読み取り専用トランザクションをテスト
func TestRoTransaction(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	t.Run("正常系: トランザクションが正常にコミットされる", func(t *testing.T) {
		err := database.RoTransaction(ctx, db, func(tx *sql.Tx) error {
			// 単純な読み取りクエリを実行
			var count int
			err := tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			t.Errorf("RoTransaction() error = %v, want nil", err)
		}
	})

	t.Run("異常系: トランザクション内でエラーが発生した場合", func(t *testing.T) {
		testErr := errors.New("test error")
		err := database.RoTransaction(ctx, db, func(tx *sql.Tx) error {
			return testErr
		})

		if err == nil {
			t.Error("RoTransaction() error = nil, want error")
		}
	})

	t.Run("異常系: 読み取り専用トランザクションで書き込みを試みる", func(t *testing.T) {
		err := database.RoTransaction(ctx, db, func(tx *sql.Tx) error {
			// 読み取り専用トランザクションで書き込みを試みる（失敗する）
			_, err := tx.Exec("CREATE TABLE test_table (id INT)")
			return err
		})

		if err == nil {
			t.Error("RoTransaction() should fail on write operation in read-only transaction")
		}
	})
}

// TestRoTransaction_Panic はパニック時のロールバックをテスト
func TestRoTransaction_Panic(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RoTransaction() should panic")
		}
	}()

	_ = database.RoTransaction(ctx, db, func(tx *sql.Tx) error {
		panic("test panic")
	})
}

// TestRwTransaction は読み書きトランザクションをテスト
func TestRwTransaction(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	t.Run("正常系: トランザクションが正常にコミットされる", func(t *testing.T) {
		err := database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
			// 単純な読み取りクエリを実行
			var count int
			err := tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			t.Errorf("RwTransaction() error = %v, want nil", err)
		}
	})

	t.Run("異常系: トランザクション内でエラーが発生した場合", func(t *testing.T) {
		testErr := errors.New("test error")
		err := database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
			return testErr
		})

		if err == nil {
			t.Error("RwTransaction() error = nil, want error")
		}
	})

	t.Run("正常系: 読み書きトランザクションで書き込みができる", func(t *testing.T) {
		// トランザクション内で一時テーブルを作成して削除
		err := database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
			// 一時テーブルを作成
			_, err := tx.Exec("CREATE TEMP TABLE test_rw_table (id INT)")
			if err != nil {
				return err
			}
			// データを挿入
			_, err = tx.Exec("INSERT INTO test_rw_table (id) VALUES (1)")
			if err != nil {
				return err
			}
			// データを読み取り
			var id int
			err = tx.QueryRow("SELECT id FROM test_rw_table LIMIT 1").Scan(&id)
			if err != nil {
				return err
			}
			if id != 1 {
				t.Errorf("expected id = 1, got %d", id)
			}
			return nil
		})

		if err != nil {
			t.Errorf("RwTransaction() error = %v, want nil", err)
		}
	})
}

// TestRwTransaction_Panic はパニック時のロールバックをテスト
func TestRwTransaction_Panic(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()

	defer func() {
		if r := recover(); r == nil {
			t.Error("RwTransaction() should panic")
		}
	}()

	_ = database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
		panic("test panic")
	})
}
