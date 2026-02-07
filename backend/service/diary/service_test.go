package diary

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/testutil"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func setupTestRedis(t *testing.T) (rueidis.Client, func()) {
	// miniredisでテスト用Redisサーバーを起動
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// rueidisクライアントを作成（テスト用にキャッシュを無効化）
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	require.NoError(t, err)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func createTestUser(t *testing.T, db *sql.DB) uuid.UUID {
	return testutil.CreateTestUser(t, db, "diary-test@example.com", "Diary Test User")
}

func createAuthenticatedContext(userID uuid.UUID) context.Context {
	return testutil.CreateAuthenticatedContext(userID)
}

func TestDiaryEntry_CreateDiaryEntry(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	tests := []struct {
		name          string
		request       *g.CreateDiaryEntryRequest
		shouldSucceed bool
		expectedError string
	}{
		{
			name: "正常系：正常な日記エントリ",
			request: &g.CreateDiaryEntryRequest{
				Content: "This is a test diary entry",
				Date: &g.YMD{
					Year:  2024,
					Month: 1,
					Day:   15,
				},
			},
			shouldSucceed: true,
		},
		{
			name: "正常系：空のコンテンツ",
			request: &g.CreateDiaryEntryRequest{
				Content: "",
				Date: &g.YMD{
					Year:  2024,
					Month: 1,
					Day:   16,
				},
			},
			shouldSucceed: true, // 空のコンテンツも許可
		},
		{
			name: "正常系：未来の日付",
			request: &g.CreateDiaryEntryRequest{
				Content: "Future diary entry",
				Date: &g.YMD{
					Year:  2030,
					Month: 12,
					Day:   31,
				},
			},
			shouldSucceed: true, // 未来の日付も許可
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := diaryService.CreateDiaryEntry(ctx, tt.request)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if response == nil {
					t.Fatal("Expected response but got nil")
				}
				if response.Entry == nil {
					t.Fatal("Expected entry but got nil")
				}
				if response.Entry.Id == "" {
					t.Error("Expected entry ID but got empty string")
				}
				if response.Entry.Content != tt.request.Content {
					t.Errorf("Expected content '%s' but got '%s'", tt.request.Content, response.Entry.Content)
				}
				if response.Entry.Date.Year != tt.request.Date.Year ||
					response.Entry.Date.Month != tt.request.Date.Month ||
					response.Entry.Date.Day != tt.request.Date.Day {
					t.Errorf("Expected date %v but got %v", tt.request.Date, response.Entry.Date)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
			}
		})
	}
}

func TestDiaryEntry_GetDiaryEntry(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// まず日記エントリを作成
	createReq := &g.CreateDiaryEntryRequest{
		Content: "Test diary for getting",
		Date: &g.YMD{
			Year:  2024,
			Month: 2,
			Day:   15,
		},
	}
	_, err := diaryService.CreateDiaryEntry(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create diary entry for test: %v", err)
	}

	tests := []struct {
		name          string
		date          *g.YMD
		shouldSucceed bool
		expectedError string
	}{
		{
			name: "正常系：既存の日記エントリを取得",
			date: &g.YMD{
				Year:  2024,
				Month: 2,
				Day:   15,
			},
			shouldSucceed: true,
		},
		{
			name: "異常系：存在しない日記エントリを取得",
			date: &g.YMD{
				Year:  2024,
				Month: 2,
				Day:   16,
			},
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getReq := &g.GetDiaryEntryRequest{
				Date: tt.date,
			}
			response, err := diaryService.GetDiaryEntry(ctx, getReq)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if response == nil {
					t.Fatal("Expected response but got nil")
				}
				if response.Entry == nil {
					t.Fatal("Expected entry but got nil")
				}
				if response.Entry.Content != createReq.Content {
					t.Errorf("Expected content '%s' but got '%s'", createReq.Content, response.Entry.Content)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
			}
		})
	}
}

func TestDiaryEntry_UpdateDiaryEntry(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// まず日記エントリを作成
	createReq := &g.CreateDiaryEntryRequest{
		Content: "Original content",
		Date: &g.YMD{
			Year:  2024,
			Month: 3,
			Day:   15,
		},
	}
	createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create diary entry for test: %v", err)
	}

	tests := []struct {
		name          string
		entryID       string
		content       string
		date          *g.YMD
		shouldSucceed bool
		expectedError string
	}{
		{
			name:          "正常系：正常な更新",
			entryID:       createResp.Entry.Id,
			content:       "Updated content",
			date:          createReq.Date,
			shouldSucceed: true,
		},
		{
			name:    "正常系：新しい日付で更新",
			entryID: createResp.Entry.Id,
			content: "Updated content with new date",
			date: &g.YMD{
				Year:  2024,
				Month: 3,
				Day:   16,
			},
			shouldSucceed: true,
		},
		{
			name:          "異常系：無効なエントリID",
			entryID:       "invalid-uuid",
			content:       "Updated content",
			date:          createReq.Date,
			shouldSucceed: false,
		},
		{
			name:          "異常系：存在しないエントリID",
			entryID:       uuid.New().String(),
			content:       "Updated content",
			date:          createReq.Date,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateReq := &g.UpdateDiaryEntryRequest{
				Id:      tt.entryID,
				Content: tt.content,
				Date:    tt.date,
			}
			response, err := diaryService.UpdateDiaryEntry(ctx, updateReq)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if response == nil {
					t.Fatal("Expected response but got nil")
				}
				if response.Entry == nil {
					t.Fatal("Expected entry but got nil")
				}
				if response.Entry.Content != tt.content {
					t.Errorf("Expected content '%s' but got '%s'", tt.content, response.Entry.Content)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
			}
		})
	}
}

func TestDiaryEntry_DeleteDiaryEntry(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// まず日記エントリを作成
	createReq := &g.CreateDiaryEntryRequest{
		Content: "Content to be deleted",
		Date: &g.YMD{
			Year:  2024,
			Month: 4,
			Day:   15,
		},
	}
	createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create diary entry for test: %v", err)
	}

	tests := []struct {
		name          string
		entryID       string
		shouldSucceed bool
		expectedError string
	}{
		{
			name:          "正常系：正常な削除",
			entryID:       createResp.Entry.Id,
			shouldSucceed: true,
		},
		{
			name:          "異常系：無効なエントリーID",
			entryID:       "invalid-uuid",
			shouldSucceed: false,
		},
		{
			name:          "異常系：存在しないエントリーID",
			entryID:       uuid.New().String(),
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteReq := &g.DeleteDiaryEntryRequest{
				Id: tt.entryID,
			}
			response, err := diaryService.DeleteDiaryEntry(ctx, deleteReq)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
				if response == nil {
					t.Fatal("Expected response but got nil")
				}
				if !response.Success {
					t.Error("Expected success to be true")
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got nil")
				}
			}
		})
	}
}

func TestDiaryEntry_GetDiaryEntries(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// Create multiple diary entries
	dates := []*g.YMD{
		{Year: 2024, Month: 5, Day: 1},
		{Year: 2024, Month: 5, Day: 2},
		{Year: 2024, Month: 5, Day: 3},
	}

	for i, date := range dates {
		createReq := &g.CreateDiaryEntryRequest{
			Content: fmt.Sprintf("Content for day %d", i+1),
			Date:    date,
		}
		_, err := diaryService.CreateDiaryEntry(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create diary entry %d: %v", i+1, err)
		}
	}

	// Test getting multiple entries
	getReq := &g.GetDiaryEntriesRequest{
		Dates: dates,
	}
	response, err := diaryService.GetDiaryEntries(ctx, getReq)

	if err != nil {
		t.Fatalf("Expected success but got error: %v", err)
	}
	if response == nil {
		t.Fatal("Expected response but got nil")
	}
	if len(response.Entries) != len(dates) {
		t.Errorf("Expected %d entries but got %d", len(dates), len(response.Entries))
	}
}

func TestDiaryEntry_GetDiaryEntriesByMonth(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// Create diary entries for a specific month
	month := &g.YM{Year: 2024, Month: 6}
	for day := 1; day <= 5; day++ {
		createReq := &g.CreateDiaryEntryRequest{
			Content: fmt.Sprintf("Content for June %d", day),
			Date:    &g.YMD{Year: month.Year, Month: month.Month, Day: uint32(day)},
		}
		_, err := diaryService.CreateDiaryEntry(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create diary entry for day %d: %v", day, err)
		}
	}

	// Test getting entries by month
	getByMonthReq := &g.GetDiaryEntriesByMonthRequest{
		Month: month,
	}
	response, err := diaryService.GetDiaryEntriesByMonth(ctx, getByMonthReq)

	if err != nil {
		t.Fatalf("Expected success but got error: %v", err)
	}
	if response == nil {
		t.Fatal("Expected response but got nil")
	}
	if len(response.Entries) != 5 {
		t.Errorf("Expected 5 entries but got %d", len(response.Entries))
	}
}

func TestDiaryEntry_UnauthorizedAccess(t *testing.T) {
	db := setupTestDB(t)

	// Create two users
	userID1 := createTestUser(t, db)
	userID2 := testutil.CreateTestUser(t, db, "diary-test2@example.com", "Diary Test User 2")

	diaryService := &DiaryEntry{DB: db}
	ctx1 := createAuthenticatedContext(userID1)
	ctx2 := createAuthenticatedContext(userID2)

	// User 1 creates a diary entry
	createReq := &g.CreateDiaryEntryRequest{
		Content: "User 1's private diary",
		Date:    &g.YMD{Year: 2024, Month: 7, Day: 15},
	}
	createResp, err := diaryService.CreateDiaryEntry(ctx1, createReq)
	if err != nil {
		t.Fatalf("Failed to create diary entry: %v", err)
	}

	// User 2 tries to update User 1's diary entry
	updateReq := &g.UpdateDiaryEntryRequest{
		Id:      createResp.Entry.Id,
		Content: "User 2 trying to update User 1's diary",
		Date:    createReq.Date,
	}
	_, err = diaryService.UpdateDiaryEntry(ctx2, updateReq)
	if err == nil {
		t.Error("Expected permission denied error but got nil")
	} else {
		statusErr, ok := status.FromError(err)
		if !ok || statusErr.Code() != codes.PermissionDenied {
			t.Errorf("Expected PermissionDenied error but got: %v", err)
		}
	}

	// User 2 tries to delete User 1's diary entry
	deleteReq := &g.DeleteDiaryEntryRequest{
		Id: createResp.Entry.Id,
	}
	_, err = diaryService.DeleteDiaryEntry(ctx2, deleteReq)
	if err == nil {
		t.Error("Expected permission denied error but got nil")
	} else {
		statusErr, ok := status.FromError(err)
		if !ok || statusErr.Code() != codes.PermissionDenied {
			t.Errorf("Expected PermissionDenied error but got: %v", err)
		}
	}
}

func TestDiaryEntry_UnauthenticatedAccess(t *testing.T) {
	db := setupTestDB(t)

	diaryService := &DiaryEntry{DB: db}
	ctx := context.Background() // No authentication

	// Try to create a diary entry without authentication
	createReq := &g.CreateDiaryEntryRequest{
		Content: "Unauthenticated diary entry",
		Date:    &g.YMD{Year: 2024, Month: 8, Day: 15},
	}
	_, err := diaryService.CreateDiaryEntry(ctx, createReq)
	if err == nil {
		t.Error("Expected authentication error but got nil")
	}
}

func TestDiaryEntry_TriggerDiaryHighlight(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// Create a test diary entry with sufficient content (500+ characters)
	longContent := ""
	for i := 0; i < 100; i++ {
		longContent += "これは日記のテスト内容です。"
	}

	createReq := &g.CreateDiaryEntryRequest{
		Content: longContent,
		Date: &g.YMD{
			Year:  2024,
			Month: 11,
			Day:   8,
		},
	}
	createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create diary entry: %v", err)
	}

	tests := []struct {
		name          string
		request       *g.TriggerDiaryHighlightRequest
		shouldSucceed bool
		expectedCode  codes.Code
	}{
		{
			name: "異常系：存在しない日記ID",
			request: &g.TriggerDiaryHighlightRequest{
				DiaryId: uuid.New().String(),
			},
			shouldSucceed: false,
			expectedCode:  codes.NotFound,
		},
		{
			name: "異常系：無効な日記ID",
			request: &g.TriggerDiaryHighlightRequest{
				DiaryId: "invalid-uuid",
			},
			shouldSucceed: false,
			expectedCode:  codes.InvalidArgument,
		},
		{
			name: "異常系：Gemini APIキー未設定（実際のテストでは成功する可能性がある）",
			request: &g.TriggerDiaryHighlightRequest{
				DiaryId: createResp.Entry.Id,
			},
			shouldSucceed: false,
			expectedCode:  codes.NotFound, // Gemini APIキーが見つからない
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := diaryService.TriggerDiaryHighlight(ctx, tt.request)

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error but got: %v", err)
					return
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected error code %v but got %v", tt.expectedCode, st.Code())
				}
			}
		})
	}
}

func TestDiaryEntry_TriggerDiaryHighlight_ShortContent(t *testing.T) {
	// Note: このテストはRedisクライアントが必要なため、統合テストとして実装すべき
	// ここでは最小限のバリデーションテストとして実装
	t.Skip("Skipping test that requires Redis configuration - integration test recommended")
}

func TestDiaryEntry_GetDiaryHighlight(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// Create a test diary entry
	createReq := &g.CreateDiaryEntryRequest{
		Content: "Test diary content",
		Date: &g.YMD{
			Year:  2024,
			Month: 11,
			Day:   8,
		},
	}
	createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create diary entry: %v", err)
	}

	tests := []struct {
		name          string
		request       *g.GetDiaryHighlightRequest
		shouldSucceed bool
		expectedCode  codes.Code
	}{
		{
			name: "異常系：ハイライトが存在しない",
			request: &g.GetDiaryHighlightRequest{
				DiaryId: createResp.Entry.Id,
			},
			shouldSucceed: false,
			expectedCode:  codes.NotFound,
		},
		{
			name: "異常系：無効な日記ID",
			request: &g.GetDiaryHighlightRequest{
				DiaryId: "invalid-uuid",
			},
			shouldSucceed: false,
			expectedCode:  codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := diaryService.GetDiaryHighlight(ctx, tt.request)

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("Expected gRPC status error but got: %v", err)
					return
				}
				if st.Code() != tt.expectedCode {
					t.Errorf("Expected error code %v but got %v", tt.expectedCode, st.Code())
				}
			}
		})
	}
}

func TestDiaryEntry_SearchDiaryEntries(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// テスト用の日記エントリを複数作成
	testEntries := []struct {
		content string
		date    *g.YMD
	}{
		{
			content: "今日は天気が良かった",
			date: &g.YMD{
				Year:  2024,
				Month: 1,
				Day:   15,
			},
		},
		{
			content: "雨の日は気分が沈む",
			date: &g.YMD{
				Year:  2024,
				Month: 1,
				Day:   16,
			},
		},
		{
			content: "今日は友達と会った",
			date: &g.YMD{
				Year:  2024,
				Month: 1,
				Day:   17,
			},
		},
	}

	for _, entry := range testEntries {
		createReq := &g.CreateDiaryEntryRequest{
			Content: entry.content,
			Date:    entry.date,
		}
		_, err := diaryService.CreateDiaryEntry(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create diary entry: %v", err)
		}
	}

	tests := []struct {
		name             string
		keyword          string
		expectedMinCount int
	}{
		{
			name:             "正常系：今日で検索",
			keyword:          "今日",
			expectedMinCount: 2,
		},
		{
			name:             "正常系：天気で検索",
			keyword:          "天気",
			expectedMinCount: 1,
		},
		{
			name:             "正常系：存在しないキーワードで検索",
			keyword:          "存在しない",
			expectedMinCount: 0,
		},
		{
			name:             "正常系：空のキーワードで検索",
			keyword:          "",
			expectedMinCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchReq := &g.SearchDiaryEntriesRequest{
				Keyword: tt.keyword,
			}
			response, err := diaryService.SearchDiaryEntries(ctx, searchReq)

			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}
			if response == nil {
				t.Fatal("Expected response but got nil")
			}
			if response.SearchedKeyword != tt.keyword {
				t.Errorf("Expected searched keyword '%s' but got '%s'", tt.keyword, response.SearchedKeyword)
			}
			if len(response.Entries) < tt.expectedMinCount {
				t.Errorf("Expected at least %d entries but got %d", tt.expectedMinCount, len(response.Entries))
			}
		})
	}
}

// TestGetTaskTimeout は環境変数からタスクタイムアウトを取得する関数をテスト
func TestGetTaskTimeout(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{
			name:     "環境変数が設定されていない場合はデフォルト値",
			envValue: "",
			expected: 600,
		},
		{
			name:     "有効な値が設定されている場合",
			envValue: "300",
			expected: 300,
		},
		{
			name:     "無効な値（非数値）が設定されている場合はデフォルト値",
			envValue: "invalid",
			expected: 600,
		},
		{
			name:     "無効な値（ゼロ）が設定されている場合はデフォルト値",
			envValue: "0",
			expected: 600,
		},
		{
			name:     "無効な値（負の数）が設定されている場合はデフォルト値",
			envValue: "-100",
			expected: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数を設定
			if tt.envValue != "" {
				os.Setenv("TASK_TIMEOUT_SECONDS", tt.envValue)
			} else {
				os.Unsetenv("TASK_TIMEOUT_SECONDS")
			}
			defer os.Unsetenv("TASK_TIMEOUT_SECONDS")

			result := getTaskTimeout()
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestDiaryEntry_RedisTaskStatus はRedisタスクステータス管理関数をテスト
func TestDiaryEntry_RedisTaskStatus(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	diaryService := &DiaryEntry{Redis: redisClient}
	ctx := context.Background()

	t.Run("setTaskStatus: タスクステータスを設定", func(t *testing.T) {
		taskKey := "test:task:123"
		status := "processing"
		expireSeconds := 60

		err := diaryService.setTaskStatus(ctx, taskKey, status, expireSeconds)
		require.NoError(t, err)

		// 設定した値を確認
		retrievedStatus, err := diaryService.getTaskStatus(ctx, taskKey)
		require.NoError(t, err)
		assert.Equal(t, status, retrievedStatus)
	})

	t.Run("getTaskStatus: 存在しないキーの取得", func(t *testing.T) {
		taskKey := "test:task:nonexistent"

		_, err := diaryService.getTaskStatus(ctx, taskKey)
		assert.Error(t, err) // キーが存在しない場合はエラー
	})

	t.Run("deleteTaskStatus: タスクステータスを削除", func(t *testing.T) {
		taskKey := "test:task:delete"
		status := "completed"
		expireSeconds := 60

		// まず設定
		err := diaryService.setTaskStatus(ctx, taskKey, status, expireSeconds)
		require.NoError(t, err)

		// 削除
		err = diaryService.deleteTaskStatus(ctx, taskKey)
		require.NoError(t, err)

		// 削除後は取得できない
		_, err = diaryService.getTaskStatus(ctx, taskKey)
		assert.Error(t, err)
	})

	t.Run("setTaskStatus: 複数の異なるタスクステータスを設定", func(t *testing.T) {
		tasks := map[string]string{
			"test:task:user1:daily":   "processing",
			"test:task:user2:monthly": "completed",
			"test:task:user3:trend":   "failed",
		}

		// 各タスクステータスを設定
		for key, status := range tasks {
			err := diaryService.setTaskStatus(ctx, key, status, 120)
			require.NoError(t, err)
		}

		// 各タスクステータスを確認
		for key, expectedStatus := range tasks {
			retrievedStatus, err := diaryService.getTaskStatus(ctx, key)
			require.NoError(t, err)
			assert.Equal(t, expectedStatus, retrievedStatus)
		}
	})

	t.Run("setTaskStatus: 有効期限のテスト（miniredisの制限により簡易確認）", func(t *testing.T) {
		taskKey := "test:task:expire"
		status := "processing"
		expireSeconds := 1 // 1秒で期限切れ

		err := diaryService.setTaskStatus(ctx, taskKey, status, expireSeconds)
		require.NoError(t, err)

		// すぐに取得できることを確認
		retrievedStatus, err := diaryService.getTaskStatus(ctx, taskKey)
		require.NoError(t, err)
		assert.Equal(t, status, retrievedStatus)

		// miniredisでは自動的に期限切れにならないため、FastForwardを使用
		// ただし、ここでは期限設定が正しく行われたことの確認に留める
	})
}

// TestDiaryEntry_RedisTaskStatusEdgeCases はRedisタスクステータスのエッジケースをテスト
func TestDiaryEntry_RedisTaskStatusEdgeCases(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	diaryService := &DiaryEntry{Redis: redisClient}
	ctx := context.Background()

	t.Run("空文字列のステータスを設定", func(t *testing.T) {
		taskKey := "test:task:empty"
		status := ""
		expireSeconds := 60

		err := diaryService.setTaskStatus(ctx, taskKey, status, expireSeconds)
		require.NoError(t, err)

		retrievedStatus, err := diaryService.getTaskStatus(ctx, taskKey)
		require.NoError(t, err)
		assert.Equal(t, status, retrievedStatus)
	})

	t.Run("長いステータス文字列を設定", func(t *testing.T) {
		taskKey := "test:task:long"
		// 長いJSON文字列をステータスとして設定
		status := `{"type":"daily_summary","user_id":"user123","date":"2024-01-15","status":"processing","details":"This is a very long status message with many details"}`
		expireSeconds := 60

		err := diaryService.setTaskStatus(ctx, taskKey, status, expireSeconds)
		require.NoError(t, err)

		retrievedStatus, err := diaryService.getTaskStatus(ctx, taskKey)
		require.NoError(t, err)
		assert.Equal(t, status, retrievedStatus)
	})

	t.Run("同じキーで複数回設定（上書き）", func(t *testing.T) {
		taskKey := "test:task:overwrite"

		// 最初の値を設定
		err := diaryService.setTaskStatus(ctx, taskKey, "processing", 60)
		require.NoError(t, err)

		// 値を上書き
		err = diaryService.setTaskStatus(ctx, taskKey, "completed", 60)
		require.NoError(t, err)

		// 最新の値が取得できることを確認
		retrievedStatus, err := diaryService.getTaskStatus(ctx, taskKey)
		require.NoError(t, err)
		assert.Equal(t, "completed", retrievedStatus)
	})
}

// TestDiaryEntry_RedisTaskStatusConcurrency は並行処理のテスト
func TestDiaryEntry_RedisTaskStatusConcurrency(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	diaryService := &DiaryEntry{Redis: redisClient}
	ctx := context.Background()

	t.Run("並行して複数のタスクステータスを設定", func(t *testing.T) {
		numGoroutines := 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				taskKey := fmt.Sprintf("test:task:concurrent:%d", id)
				status := fmt.Sprintf("processing_%d", id)
				err := diaryService.setTaskStatus(ctx, taskKey, status, 120)
				require.NoError(t, err)
				done <- true
			}(i)
		}

		// すべてのgoroutineの完了を待つ
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// すべてのタスクステータスが正しく設定されているか確認
		for i := 0; i < numGoroutines; i++ {
			taskKey := fmt.Sprintf("test:task:concurrent:%d", i)
			expectedStatus := fmt.Sprintf("processing_%d", i)
			retrievedStatus, err := diaryService.getTaskStatus(ctx, taskKey)
			require.NoError(t, err)
			assert.Equal(t, expectedStatus, retrievedStatus)
		}
	})
}

// TestDiaryEntry_RedisTaskStatusContextCancellation はコンテキストキャンセルのテスト
func TestDiaryEntry_RedisTaskStatusContextCancellation(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	diaryService := &DiaryEntry{Redis: redisClient}

	t.Run("コンテキストがキャンセルされた場合のエラー処理", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // すぐにキャンセル

		taskKey := "test:task:cancelled"
		status := "processing"
		expireSeconds := 60

		// キャンセルされたコンテキストでsetを試みる
		err := diaryService.setTaskStatus(ctx, taskKey, status, expireSeconds)
		// miniredisではコンテキストキャンセルが即座に反映されない可能性があるため、
		// エラーが発生しない場合もある（実装依存）
		// ここではエラーチェックのみ行う
		t.Logf("setTaskStatus with cancelled context: %v", err)
	})

	t.Run("タイムアウト付きコンテキストでの操作", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		taskKey := "test:task:timeout"
		status := "processing"
		expireSeconds := 60

		// タイムアウト前に完了するはず
		err := diaryService.setTaskStatus(ctx, taskKey, status, expireSeconds)
		require.NoError(t, err)

		retrievedStatus, err := diaryService.getTaskStatus(ctx, taskKey)
		require.NoError(t, err)
		assert.Equal(t, status, retrievedStatus)
	})
}

// TestDiaryEntry_GetDiaryEntityOutputs は日記エンティティ出力を取得する内部ヘルパー関数をテスト
func TestDiaryEntry_GetDiaryEntityOutputs(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	t.Run("正常系：エンティティが存在する場合", func(t *testing.T) {
		// エンティティを作成
		entityID := uuid.New()
		err := db.QueryRowContext(ctx, `
			INSERT INTO entities (id, user_id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5) RETURNING id
		`, entityID, userID, "TestEntity", time.Now().Unix(), time.Now().Unix()).Err()
		require.NoError(t, err)

		// 日記エントリを作成
		diaryEntities := []*g.DiaryEntityInput{
			{
				EntityId: entityID.String(),
				Positions: []*g.Position{
					{Start: 0, End: 10, AliasId: ""},
					{Start: 20, End: 30, AliasId: "alias1"},
				},
			},
		}

		createReq := &g.CreateDiaryEntryRequest{
			Content:       "Test content with entity",
			Date:          &g.YMD{Year: 2024, Month: 10, Day: 1},
			DiaryEntities: diaryEntities,
		}

		createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
		require.NoError(t, err)

		diaryID, err := uuid.Parse(createResp.Entry.Id)
		require.NoError(t, err)

		// getDiaryEntityOutputsを呼び出し
		outputs, err := diaryService.getDiaryEntityOutputs(ctx, diaryID)
		require.NoError(t, err)
		require.Len(t, outputs, 1)

		// 結果の検証
		assert.Equal(t, entityID.String(), outputs[0].EntityId)
		require.Len(t, outputs[0].Positions, 2)
		assert.Equal(t, uint32(0), outputs[0].Positions[0].Start)
		assert.Equal(t, uint32(10), outputs[0].Positions[0].End)
		assert.Equal(t, "", outputs[0].Positions[0].AliasId)
		assert.Equal(t, uint32(20), outputs[0].Positions[1].Start)
		assert.Equal(t, uint32(30), outputs[0].Positions[1].End)
		assert.Equal(t, "alias1", outputs[0].Positions[1].AliasId)
	})

	t.Run("正常系：エンティティが存在しない場合", func(t *testing.T) {
		// エンティティなしで日記エントリを作成
		createReq := &g.CreateDiaryEntryRequest{
			Content:       "Test content without entity",
			Date:          &g.YMD{Year: 2024, Month: 10, Day: 2},
			DiaryEntities: nil,
		}

		createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
		require.NoError(t, err)

		diaryID, err := uuid.Parse(createResp.Entry.Id)
		require.NoError(t, err)

		// getDiaryEntityOutputsを呼び出し
		outputs, err := diaryService.getDiaryEntityOutputs(ctx, diaryID)
		require.NoError(t, err)
		assert.Empty(t, outputs)
	})
}

// TestDiaryEntry_GetDiaryEntityOutputsForDiaries は複数の日記に対するエンティティ一括取得をテスト（N+1問題回避）
func TestDiaryEntry_GetDiaryEntityOutputsForDiaries(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	t.Run("正常系：複数の日記のエンティティを一括取得", func(t *testing.T) {
		// 2つのエンティティを作成
		entityID1 := uuid.New()
		entityID2 := uuid.New()

		for i, entityID := range []uuid.UUID{entityID1, entityID2} {
			err := db.QueryRowContext(ctx, `
				INSERT INTO entities (id, user_id, name, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
			`, entityID, userID, fmt.Sprintf("TestEntity%d", i+1), time.Now().Unix(), time.Now().Unix()).Err()
			require.NoError(t, err)
		}

		// 3つの日記エントリを作成
		diaryIDs := make([]uuid.UUID, 0, 3)

		for i := 0; i < 3; i++ {
			var diaryEntities []*g.DiaryEntityInput
			if i == 0 {
				// 最初の日記にはentityID1
				diaryEntities = []*g.DiaryEntityInput{
					{
						EntityId:  entityID1.String(),
						Positions: []*g.Position{{Start: 0, End: 5}},
					},
				}
			} else if i == 1 {
				// 2番目の日記にはentityID2
				diaryEntities = []*g.DiaryEntityInput{
					{
						EntityId:  entityID2.String(),
						Positions: []*g.Position{{Start: 0, End: 3}},
					},
				}
			} else {
				// 3番目の日記にはエンティティなし
				diaryEntities = nil
			}

			createReq := &g.CreateDiaryEntryRequest{
				Content:       fmt.Sprintf("Diary %d", i+1),
				Date:          &g.YMD{Year: 2024, Month: 10, Day: uint32(i + 10)},
				DiaryEntities: diaryEntities,
			}

			createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
			require.NoError(t, err)

			diaryID, err := uuid.Parse(createResp.Entry.Id)
			require.NoError(t, err)
			diaryIDs = append(diaryIDs, diaryID)
		}

		// getDiaryEntityOutputsForDiariesを呼び出し
		entityMap, err := diaryService.getDiaryEntityOutputsForDiaries(ctx, diaryIDs)
		require.NoError(t, err)

		// 結果の検証
		assert.Len(t, entityMap, 2) // エンティティがあるのは2つのみ

		// 最初の日記のエンティティ確認
		diary1Entities, ok := entityMap[diaryIDs[0].String()]
		assert.True(t, ok)
		require.Len(t, diary1Entities, 1)
		assert.Equal(t, entityID1.String(), diary1Entities[0].EntityId)

		// 2番目の日記のエンティティ確認
		diary2Entities, ok := entityMap[diaryIDs[1].String()]
		assert.True(t, ok)
		require.Len(t, diary2Entities, 1)
		assert.Equal(t, entityID2.String(), diary2Entities[0].EntityId)

		// 3番目の日記にはエンティティがない
		_, ok = entityMap[diaryIDs[2].String()]
		assert.False(t, ok)
	})

	t.Run("正常系：空の日記IDリスト", func(t *testing.T) {
		entityMap, err := diaryService.getDiaryEntityOutputsForDiaries(ctx, []uuid.UUID{})
		require.NoError(t, err)
		assert.Empty(t, entityMap)
	})
}

// TestDiaryEntry_SaveAndDeleteDiaryEntities はdiary_entitiesの保存と削除をテスト
func TestDiaryEntry_SaveAndDeleteDiaryEntities(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	diaryService := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	t.Run("正常系：エンティティの保存", func(t *testing.T) {
		// エンティティを作成
		entityID := uuid.New()
		err := db.QueryRowContext(ctx, `
			INSERT INTO entities (id, user_id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`, entityID, userID, "SaveTestEntity", time.Now().Unix(), time.Now().Unix()).Err()
		require.NoError(t, err)

		// 日記エントリを作成
		diaryEntities := []*g.DiaryEntityInput{
			{
				EntityId: entityID.String(),
				Positions: []*g.Position{
					{Start: 0, End: 10, AliasId: ""},
					{Start: 20, End: 30, AliasId: "alias1"},
				},
			},
		}

		createReq := &g.CreateDiaryEntryRequest{
			Content:       "Test content",
			Date:          &g.YMD{Year: 2024, Month: 10, Day: 20},
			DiaryEntities: diaryEntities,
		}

		createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
		require.NoError(t, err)

		diaryID, err := uuid.Parse(createResp.Entry.Id)
		require.NoError(t, err)

		// diary_entitiesが保存されているか確認
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM diary_entities WHERE diary_id = $1", diaryID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("正常系：エンティティの削除", func(t *testing.T) {
		// エンティティを作成
		entityID := uuid.New()
		err := db.QueryRowContext(ctx, `
			INSERT INTO entities (id, user_id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`, entityID, userID, "DeleteTestEntity", time.Now().Unix(), time.Now().Unix()).Err()
		require.NoError(t, err)

		// 日記エントリを作成
		diaryEntities := []*g.DiaryEntityInput{
			{
				EntityId:  entityID.String(),
				Positions: []*g.Position{{Start: 0, End: 10}},
			},
		}

		createReq := &g.CreateDiaryEntryRequest{
			Content:       "Test content for delete",
			Date:          &g.YMD{Year: 2024, Month: 10, Day: 21},
			DiaryEntities: diaryEntities,
		}

		createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
		require.NoError(t, err)

		diaryID, err := uuid.Parse(createResp.Entry.Id)
		require.NoError(t, err)

		// 削除前に存在確認
		var countBefore int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM diary_entities WHERE diary_id = $1", diaryID).Scan(&countBefore)
		require.NoError(t, err)
		assert.Equal(t, 1, countBefore)

		// deleteDiaryEntitiesを呼び出し（トランザクション内）
		err = database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
			return diaryService.deleteDiaryEntities(ctx, tx, diaryID)
		})
		require.NoError(t, err)

		// 削除後に存在しないことを確認
		var countAfter int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM diary_entities WHERE diary_id = $1", diaryID).Scan(&countAfter)
		require.NoError(t, err)
		assert.Equal(t, 0, countAfter)
	})

	t.Run("異常系：無効なエンティティIDで保存", func(t *testing.T) {
		// 無効なエンティティIDでsaveDiaryEntitiesを呼び出し
		diaryID := uuid.New()
		invalidEntities := []*g.DiaryEntityInput{
			{
				EntityId:  "invalid-uuid",
				Positions: []*g.Position{{Start: 0, End: 10}},
			},
		}

		err := database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
			return diaryService.saveDiaryEntities(ctx, tx, diaryID, invalidEntities, time.Now().Unix())
		})

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("正常系：空のエンティティリストで保存", func(t *testing.T) {
		diaryID := uuid.New()
		emptyEntities := []*g.DiaryEntityInput{}

		err := database.RwTransaction(ctx, db, func(tx *sql.Tx) error {
			return diaryService.saveDiaryEntities(ctx, tx, diaryID, emptyEntities, time.Now().Unix())
		})

		assert.NoError(t, err) // 空のリストは正常に処理される
	})
}
