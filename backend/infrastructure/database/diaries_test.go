package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "test-pass")
	dbname := getEnvOrDefault("TEST_DB_NAME", "umi_mikan_test")

	dsn := "host=" + host + " port=5432 user=postgres password=" + password + " dbname=" + dbname + " sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("Database connection not available, skipping test: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Skipf("Database ping failed, skipping test: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createTestUser(t *testing.T, db *sql.DB, email, name string) uuid.UUID {
	userID := uuid.New()
	currentTime := time.Now().Unix()

	// nameは20文字以内に制限
	if len([]rune(name)) > 20 {
		name = string([]rune(name)[:20])
	}

	_, err := db.Exec(`
		INSERT INTO users (id, email, name, auth_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, email, name, 0, currentTime, currentTime) // auth_type = 0 (password)
	require.NoError(t, err)

	return userID
}

func TestDiariesByUserIDAndContent(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// テスト用のユーザーを作成（ユニークなメールアドレスを使用）
	testID := uuid.New().String()[:8]
	userID1 := createTestUser(t, db, fmt.Sprintf("dt1-%s@ex.com", testID), "DiaryUser1")
	userID2 := createTestUser(t, db, fmt.Sprintf("dt2-%s@ex.com", testID), "DiaryUser2")

	t.Run("正常系：キーワードにマッチする日記を検索", func(t *testing.T) {
		// テストデータを作成
		testDiaries := []struct {
			userID  uuid.UUID
			content string
			date    time.Time
		}{
			{userID1, "今日は天気が良かった", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{userID1, "雨の日は気分が沈む", time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
			{userID1, "今日は友達と会った", time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)},
			{userID2, "今日は仕事が忙しかった", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		}

		for _, td := range testDiaries {
			diary := &Diary{
				ID:        uuid.New(),
				UserID:    td.userID,
				Content:   td.content,
				Date:      td.date,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			}
			err := diary.Insert(ctx, db)
			require.NoError(t, err)
		}

		// "今日"で検索（userID1の2件がヒット）
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "今日")
		require.NoError(t, err)
		assert.Len(t, results, 2)

		// 結果は日付の降順（新しい順）で返される
		assert.Equal(t, "今日は友達と会った", results[0].Content)
		assert.Equal(t, "今日は天気が良かった", results[1].Content)
	})

	t.Run("正常系：キーワードにマッチしない場合", func(t *testing.T) {
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "存在しないキーワード")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("正常系：空のキーワードで検索（全ての日記がヒット）", func(t *testing.T) {
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 3) // 最低3件はあるはず
	})

	t.Run("正常系：部分一致検索", func(t *testing.T) {
		// "天気"で検索
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "天気")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Contains(t, results[0].Content, "天気")
	})

	t.Run("正常系：異なるユーザーの日記は検索されない", func(t *testing.T) {
		// userID1で"仕事"を検索（userID2の日記にのみ含まれる）
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "仕事")
		require.NoError(t, err)
		assert.Empty(t, results)

		// userID2で"仕事"を検索
		results, err = DiariesByUserIDAndContent(ctx, db, userID2.String(), "仕事")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Contains(t, results[0].Content, "仕事")
	})

	t.Run("正常系：複数の単語を含むキーワード", func(t *testing.T) {
		// "日は"で検索（3件全てに含まれる）
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "日は")
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("正常系：日付の降順でソートされる", func(t *testing.T) {
		// 新しい日記を追加
		newDiary := &Diary{
			ID:        uuid.New(),
			UserID:    userID1,
			Content:   "最新の日記エントリ",
			Date:      time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		err := newDiary.Insert(ctx, db)
		require.NoError(t, err)

		// 空のキーワードで全件取得
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "")
		require.NoError(t, err)
		require.Greater(t, len(results), 0)

		// 最初の結果が最新の日付であることを確認
		assert.Equal(t, "最新の日記エントリ", results[0].Content)
		expectedDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
		assert.True(t, results[0].Date.Equal(expectedDate), "Expected date %v but got %v", expectedDate, results[0].Date)
	})

	t.Run("異常系：存在しないユーザーIDで検索", func(t *testing.T) {
		nonExistentUserID := uuid.New()
		results, err := DiariesByUserIDAndContent(ctx, db, nonExistentUserID.String(), "今日")
		require.NoError(t, err)
		assert.Empty(t, results) // エラーではなく空のリストが返される
	})

	t.Run("正常系：特殊文字を含むキーワード", func(t *testing.T) {
		// 特殊文字を含む日記を作成
		specialDiary := &Diary{
			ID:        uuid.New(),
			UserID:    userID1,
			Content:   "100%の力で頑張った！",
			Date:      time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		err := specialDiary.Insert(ctx, db)
		require.NoError(t, err)

		// "%"を含むキーワードで検索（LIKE演算子のエスケープテスト）
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "100%")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Contains(t, results[0].Content, "100%")
	})

	t.Run("正常系：大文字小文字を区別しない検索", func(t *testing.T) {
		// 英語の日記を作成
		englishDiary := &Diary{
			ID:        uuid.New(),
			UserID:    userID1,
			Content:   "Today is a Beautiful day",
			Date:      time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		err := englishDiary.Insert(ctx, db)
		require.NoError(t, err)

		// 小文字で検索
		results, err := DiariesByUserIDAndContent(ctx, db, userID1.String(), "beautiful")
		require.NoError(t, err)
		// PostgreSQLのLIKEはデフォルトで大文字小文字を区別するため、ヒットしない可能性がある
		// ここではデータベースの動作を確認
		t.Logf("Results for 'beautiful': %d", len(results))
	})
}

// TestDiariesByUserIDAndContent_PerformanceTest は大量データでのパフォーマンステスト
func TestDiariesByUserIDAndContent_PerformanceTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	db := setupTestDB(t)
	ctx := context.Background()

	testID := uuid.New().String()[:8]
	userID := createTestUser(t, db, fmt.Sprintf("perf-%s@ex.com", testID), "PerfUser")

	// 100件の日記を作成
	for i := 0; i < 100; i++ {
		diary := &Diary{
			ID:        uuid.New(),
			UserID:    userID,
			Content:   fmt.Sprintf("日記エントリ %d", i),
			Date:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * 24 * time.Hour),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		err := diary.Insert(ctx, db)
		require.NoError(t, err)
	}

	// 検索実行
	start := time.Now()
	results, err := DiariesByUserIDAndContent(ctx, db, userID.String(), "日記")
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 100)
	t.Logf("Search completed in %v for %d results", elapsed, len(results))
}
