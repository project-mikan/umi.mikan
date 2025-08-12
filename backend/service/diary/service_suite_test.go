package diary

import (
	"testing"

	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)


func TestDiarySuite_CRUD(t *testing.T) {
	runner := testutil.NewTestRunner(t)
	runner.Run(func(suite *testutil.TestSuite) {
		mockRedis := createMockRedisClient()
		diaryService := &DiaryEntry{DB: suite.DB, Redis: mockRedis}
		ctx := suite.GetAuthenticatedContext()

		// Create a diary entry
		createReq := &g.CreateDiaryEntryRequest{
			Content: "Suite test diary entry",
			Date: &g.YMD{
				Year:  2025,
				Month: 1,
				Day:   15,
			},
		}
		createResp, err := diaryService.CreateDiaryEntry(ctx, createReq)
		if err != nil {
			t.Fatalf("Failed to create diary entry: %v", err)
		}
		if createResp.Entry == nil {
			t.Fatal("Expected diary entry in response")
		}

		entryID := createResp.Entry.Id

		// Get the diary entry
		getReq := &g.GetDiaryEntryRequest{
			Date: createReq.Date,
		}
		getResp, err := diaryService.GetDiaryEntry(ctx, getReq)
		if err != nil {
			t.Fatalf("Failed to get diary entry: %v", err)
		}
		if getResp.Entry.Content != createReq.Content {
			t.Errorf("Expected content '%s' but got '%s'", createReq.Content, getResp.Entry.Content)
		}

		// Update the diary entry
		updateReq := &g.UpdateDiaryEntryRequest{
			Id:      entryID,
			Content: "Updated suite test diary entry",
			Date:    createReq.Date,
		}
		updateResp, err := diaryService.UpdateDiaryEntry(ctx, updateReq)
		if err != nil {
			t.Fatalf("Failed to update diary entry: %v", err)
		}
		if updateResp.Entry.Content != updateReq.Content {
			t.Errorf("Expected updated content '%s' but got '%s'", updateReq.Content, updateResp.Entry.Content)
		}

		// Delete the diary entry
		deleteReq := &g.DeleteDiaryEntryRequest{
			Id: entryID,
		}
		deleteResp, err := diaryService.DeleteDiaryEntry(ctx, deleteReq)
		if err != nil {
			t.Fatalf("Failed to delete diary entry: %v", err)
		}
		if !deleteResp.Success {
			t.Error("Expected deletion to be successful")
		}

		// Verify deletion
		_, err = diaryService.GetDiaryEntry(ctx, getReq)
		if err == nil {
			t.Error("Expected error when getting deleted diary entry")
		}
	})
}

func TestDiarySuite_UserIsolation(t *testing.T) {
	runner := testutil.NewTestRunner(t)
	runner.RunWithData(func(suite *testutil.TestSuite, testData *testutil.TestData) {
		mockRedis := createMockRedisClient()
		diaryService := &DiaryEntry{DB: suite.DB, Redis: mockRedis}

		// Add another test user
		user2ID := testData.AddTestUser("user2-suite@example.com", "User Two", "password123")

		ctx1 := suite.GetAuthenticatedContext()
		ctx2 := testutil.CreateAuthenticatedContext(user2ID)

		// User 1 creates a diary entry
		createReq1 := &g.CreateDiaryEntryRequest{
			Content: "User 1's private diary",
			Date:    &g.YMD{Year: 2025, Month: 2, Day: 1},
		}
		createResp1, err := diaryService.CreateDiaryEntry(ctx1, createReq1)
		if err != nil {
			t.Fatalf("User 1 failed to create diary entry: %v", err)
		}

		// User 2 creates a diary entry
		createReq2 := &g.CreateDiaryEntryRequest{
			Content: "User 2's private diary",
			Date:    &g.YMD{Year: 2025, Month: 2, Day: 2},
		}
		_, err = diaryService.CreateDiaryEntry(ctx2, createReq2)
		if err != nil {
			t.Fatalf("User 2 failed to create diary entry: %v", err)
		}

		// Verify User 1 can access their own diary
		getReq1 := &g.GetDiaryEntryRequest{Date: createReq1.Date}
		getResp1, err := diaryService.GetDiaryEntry(ctx1, getReq1)
		if err != nil {
			t.Fatalf("User 1 failed to get their own diary: %v", err)
		}
		if getResp1.Entry.Content != createReq1.Content {
			t.Errorf("User 1 got wrong content: expected '%s', got '%s'", createReq1.Content, getResp1.Entry.Content)
		}

		// Verify User 2 cannot access User 1's diary by trying to update it
		updateReq := &g.UpdateDiaryEntryRequest{
			Id:      createResp1.Entry.Id,
			Content: "User 2 trying to update User 1's diary",
			Date:    createReq1.Date,
		}
		_, err = diaryService.UpdateDiaryEntry(ctx2, updateReq)
		if err == nil {
			t.Error("Expected permission denied error when User 2 tries to update User 1's diary")
		}
	})
}
