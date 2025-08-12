package diary

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock Redis client for testing
type mockRedisClient struct {
	data map[string]string
}

func (m *mockRedisClient) SetDiaryCount(ctx context.Context, userID string, count uint32) error {
	key := fmt.Sprintf("diary_count:%s", userID)
	m.data[key] = fmt.Sprintf("%d", count)
	return nil
}

func (m *mockRedisClient) GetDiaryCount(ctx context.Context, userID string) (uint32, error) {
	key := fmt.Sprintf("diary_count:%s", userID)
	val, exists := m.data[key]
	if !exists {
		return 0, fmt.Errorf("cache miss")
	}
	
	var count uint32
	_, err := fmt.Sscanf(val, "%d", &count)
	if err != nil {
		return 0, fmt.Errorf("failed to parse cached count: %w", err)
	}
	
	return count, nil
}

func (m *mockRedisClient) DeleteDiaryCount(ctx context.Context, userID string) error {
	key := fmt.Sprintf("diary_count:%s", userID)
	delete(m.data, key)
	return nil
}

func (m *mockRedisClient) Close() error {
	return nil
}

func createMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string]string),
	}
}

// Error-inducing mock Redis client for testing error scenarios
type errorMockRedisClient struct {
	failDelete bool
	failSet    bool
}

func (m *errorMockRedisClient) SetDiaryCount(ctx context.Context, userID string, count uint32) error {
	if m.failSet {
		return fmt.Errorf("mock redis set error")
	}
	return nil
}

func (m *errorMockRedisClient) GetDiaryCount(ctx context.Context, userID string) (uint32, error) {
	return 0, fmt.Errorf("cache miss")
}

func (m *errorMockRedisClient) DeleteDiaryCount(ctx context.Context, userID string) error {
	if m.failDelete {
		return fmt.Errorf("mock redis delete error")
	}
	return nil
}

func (m *errorMockRedisClient) Close() error {
	return nil
}

func setupTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
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
	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
	ctx := createAuthenticatedContext(userID)

	tests := []struct {
		name          string
		request       *g.CreateDiaryEntryRequest
		shouldSucceed bool
		expectedError string
	}{
		{
			name: "Valid diary entry",
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
			name: "Empty content",
			request: &g.CreateDiaryEntryRequest{
				Content: "",
				Date: &g.YMD{
					Year:  2024,
					Month: 1,
					Day:   16,
				},
			},
			shouldSucceed: true, // Empty content should be allowed
		},
		{
			name: "Future date",
			request: &g.CreateDiaryEntryRequest{
				Content: "Future diary entry",
				Date: &g.YMD{
					Year:  2030,
					Month: 12,
					Day:   31,
				},
			},
			shouldSucceed: true, // Future dates should be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := diaryService.CreateDiaryEntry(ctx, tt.request)

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if response == nil {
					t.Error("Expected response but got nil")
					return
				}
				if response.Entry == nil {
					t.Error("Expected entry but got nil")
					return
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
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestDiaryEntry_GetDiaryEntry(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
	ctx := createAuthenticatedContext(userID)

	// Create a diary entry first
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
			name: "Get existing diary entry",
			date: &g.YMD{
				Year:  2024,
				Month: 2,
				Day:   15,
			},
			shouldSucceed: true,
		},
		{
			name: "Get non-existent diary entry",
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
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if response == nil {
					t.Error("Expected response but got nil")
					return
				}
				if response.Entry == nil {
					t.Error("Expected entry but got nil")
					return
				}
				if response.Entry.Content != createReq.Content {
					t.Errorf("Expected content '%s' but got '%s'", createReq.Content, response.Entry.Content)
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestDiaryEntry_UpdateDiaryEntry(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
	ctx := createAuthenticatedContext(userID)

	// Create a diary entry first
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
			name:          "Valid update",
			entryID:       createResp.Entry.Id,
			content:       "Updated content",
			date:          createReq.Date,
			shouldSucceed: true,
		},
		{
			name:    "Update with new date",
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
			name:          "Invalid entry ID",
			entryID:       "invalid-uuid",
			content:       "Updated content",
			date:          createReq.Date,
			shouldSucceed: false,
		},
		{
			name:          "Non-existent entry ID",
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
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if response == nil {
					t.Error("Expected response but got nil")
					return
				}
				if response.Entry == nil {
					t.Error("Expected entry but got nil")
					return
				}
				if response.Entry.Content != tt.content {
					t.Errorf("Expected content '%s' but got '%s'", tt.content, response.Entry.Content)
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestDiaryEntry_DeleteDiaryEntry(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
	ctx := createAuthenticatedContext(userID)

	// Create a diary entry first
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
			name:          "Valid deletion",
			entryID:       createResp.Entry.Id,
			shouldSucceed: true,
		},
		{
			name:          "Invalid entry ID",
			entryID:       "invalid-uuid",
			shouldSucceed: false,
		},
		{
			name:          "Non-existent entry ID",
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
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if response == nil {
					t.Error("Expected response but got nil")
					return
				}
				if !response.Success {
					t.Error("Expected success to be true")
				}
			} else {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			}
		})
	}
}

func TestDiaryEntry_GetDiaryEntries(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
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
	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
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

	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
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

func TestDiaryEntry_GetDiaryCount(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
	ctx := createAuthenticatedContext(userID)

	// Test with empty diary count
	getCountReq := &g.GetDiaryCountRequest{}
	response, err := diaryService.GetDiaryCount(ctx, getCountReq)
	if err != nil {
		t.Fatalf("Expected success but got error: %v", err)
	}
	if response == nil {
		t.Fatal("Expected response but got nil")
	}
	if response.Count != 0 {
		t.Errorf("Expected count to be 0 but got %d", response.Count)
	}

	// Create some diary entries
	dates := []*g.YMD{
		{Year: 2024, Month: 9, Day: 1},
		{Year: 2024, Month: 9, Day: 2},
		{Year: 2024, Month: 9, Day: 3},
	}

	for i, date := range dates {
		createReq := &g.CreateDiaryEntryRequest{
			Content: fmt.Sprintf("Test diary entry %d", i+1),
			Date:    date,
		}
		_, err := diaryService.CreateDiaryEntry(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create diary entry %d: %v", i+1, err)
		}
	}

	// Test diary count after creating entries
	response, err = diaryService.GetDiaryCount(ctx, getCountReq)
	if err != nil {
		t.Fatalf("Expected success but got error: %v", err)
	}
	if response == nil {
		t.Fatal("Expected response but got nil")
	}
	if response.Count != uint32(len(dates)) {
		t.Errorf("Expected count to be %d but got %d", len(dates), response.Count)
	}

	// Delete one entry and check count decreases
	// First, get one of the created entries
	getReq := &g.GetDiaryEntryRequest{Date: dates[0]}
	getResp, err := diaryService.GetDiaryEntry(ctx, getReq)
	if err != nil {
		t.Fatalf("Failed to get diary entry for deletion test: %v", err)
	}

	deleteReq := &g.DeleteDiaryEntryRequest{Id: getResp.Entry.Id}
	_, err = diaryService.DeleteDiaryEntry(ctx, deleteReq)
	if err != nil {
		t.Fatalf("Failed to delete diary entry: %v", err)
	}

	// Test diary count after deletion
	response, err = diaryService.GetDiaryCount(ctx, getCountReq)
	if err != nil {
		t.Fatalf("Expected success but got error: %v", err)
	}
	if response == nil {
		t.Fatal("Expected response but got nil")
	}
	expectedCount := uint32(len(dates) - 1)
	if response.Count != expectedCount {
		t.Errorf("Expected count to be %d but got %d", expectedCount, response.Count)
	}
}

func TestDiaryEntry_GetDiaryCount_MultipleUsers(t *testing.T) {
	db := setupTestDB(t)

	// Create two users
	userID1 := createTestUser(t, db)
	userID2 := testutil.CreateTestUser(t, db, "diary-count-test2@example.com", "Test User 2")

	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
	ctx1 := createAuthenticatedContext(userID1)
	ctx2 := createAuthenticatedContext(userID2)

	// User 1 creates diary entries
	for i := 1; i <= 3; i++ {
		createReq := &g.CreateDiaryEntryRequest{
			Content: fmt.Sprintf("User 1 diary entry %d", i),
			Date:    &g.YMD{Year: 2024, Month: 10, Day: uint32(i)},
		}
		_, err := diaryService.CreateDiaryEntry(ctx1, createReq)
		if err != nil {
			t.Fatalf("Failed to create diary entry for user 1: %v", err)
		}
	}

	// User 2 creates diary entries
	for i := 1; i <= 2; i++ {
		createReq := &g.CreateDiaryEntryRequest{
			Content: fmt.Sprintf("User 2 diary entry %d", i),
			Date:    &g.YMD{Year: 2024, Month: 10, Day: uint32(i + 10)},
		}
		_, err := diaryService.CreateDiaryEntry(ctx2, createReq)
		if err != nil {
			t.Fatalf("Failed to create diary entry for user 2: %v", err)
		}
	}

	// Check count for user 1
	getCountReq := &g.GetDiaryCountRequest{}
	response1, err := diaryService.GetDiaryCount(ctx1, getCountReq)
	if err != nil {
		t.Fatalf("Expected success for user 1 but got error: %v", err)
	}
	if response1.Count != 3 {
		t.Errorf("Expected user 1 count to be 3 but got %d", response1.Count)
	}

	// Check count for user 2
	response2, err := diaryService.GetDiaryCount(ctx2, getCountReq)
	if err != nil {
		t.Fatalf("Expected success for user 2 but got error: %v", err)
	}
	if response2.Count != 2 {
		t.Errorf("Expected user 2 count to be 2 but got %d", response2.Count)
	}
}

func TestDiaryEntry_UnauthenticatedAccess(t *testing.T) {
	db := setupTestDB(t)

	mockRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: mockRedis}
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

	// Try to get diary count without authentication
	getCountReq := &g.GetDiaryCountRequest{}
	_, err = diaryService.GetDiaryCount(ctx, getCountReq)
	if err == nil {
		t.Error("Expected authentication error but got nil")
	}
}

func TestDiaryEntry_CacheDeleteError_CreateDiary(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	errorRedis := &errorMockRedisClient{failDelete: true}
	diaryService := &DiaryEntry{DB: db, Redis: errorRedis}
	ctx := createAuthenticatedContext(userID)

	createReq := &g.CreateDiaryEntryRequest{
		Content: "Test diary with cache delete error",
		Date: &g.YMD{
			Year:  2024,
			Month: 11,
			Day:   1,
		},
	}

	_, err := diaryService.CreateDiaryEntry(ctx, createReq)
	if err == nil {
		t.Error("Expected cache deletion error but got nil")
	}
	if !strings.Contains(fmt.Sprintf("%v", err), "failed to invalidate diary count cache after creation") {
		t.Errorf("Expected cache invalidation error message but got: %v", err)
	}
}

func TestDiaryEntry_CacheDeleteError_DeleteDiary(t *testing.T) {
	db := setupTestDB(t)

	userID := createTestUser(t, db)
	// First create with normal redis to get an entry
	normalRedis := createMockRedisClient()
	diaryService := &DiaryEntry{DB: db, Redis: normalRedis}
	ctx := createAuthenticatedContext(userID)

	createReq := &g.CreateDiaryEntryRequest{
		Content: "Test diary for delete error test",
		Date: &g.YMD{
			Year:  2024,
			Month: 11,
			Day:   2,
		},
	}

	createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create diary entry: %v", err)
	}

	// Now switch to error redis for deletion
	errorRedis := &errorMockRedisClient{failDelete: true}
	diaryService.Redis = errorRedis

	deleteReq := &g.DeleteDiaryEntryRequest{
		Id: createResp.Entry.Id,
	}

	_, err = diaryService.DeleteDiaryEntry(ctx, deleteReq)
	if err == nil {
		t.Error("Expected cache deletion error but got nil")
	}
	if !strings.Contains(fmt.Sprintf("%v", err), "failed to invalidate diary count cache after deletion") {
		t.Errorf("Expected cache invalidation error message but got: %v", err)
	}
}
