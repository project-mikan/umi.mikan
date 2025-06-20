package diary

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	diaryService := &DiaryEntry{DB: db}
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
	diaryService := &DiaryEntry{DB: db}
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
	diaryService := &DiaryEntry{DB: db}
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
	diaryService := &DiaryEntry{DB: db}
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
