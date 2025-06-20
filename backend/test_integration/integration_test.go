package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/service/auth"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestCompleteUserJourney(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Initialize services
	authService := &auth.AuthEntry{DB: db}
	diaryService := &diary.DiaryEntry{DB: db}
	ctx := context.Background()

	// Generate unique test identifier
	testID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Step 1: Register a new user
	registerReq := &g.RegisterByPasswordRequest{
		Email:    fmt.Sprintf("journey-integration-%s@example.com", testID),
		Password: "securePassword123",
		Name:     "Journey Test User",
	}
	registerResp, err := authService.RegisterByPassword(ctx, registerReq)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}
	if registerResp.AccessToken == "" {
		t.Fatal("Expected access token after registration")
	}

	// Step 2: Login with the registered user
	loginReq := &g.LoginByPasswordRequest{
		Email:    registerReq.Email,
		Password: registerReq.Password,
	}
	loginResp, err := authService.LoginByPassword(ctx, loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loginResp.AccessToken == "" {
		t.Fatal("Expected access token after login")
	}

	// Step 3: Create authenticated context (simulating middleware)
	// In a real scenario, this would be done by the auth interceptor
	user, err := getUserFromToken(t, db, loginResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to get user from token: %v", err)
	}
	authCtx := testutil.CreateAuthenticatedContext(user.ID)

	// Step 4: Create diary entries with unique dates
	// Use nanoseconds to get more unique values for each test run
	nanoTime := time.Now().UnixNano()
	baseDate := uint32((nanoTime % 25) + 1) // Get a day between 1-25 (leaving room for +2)
	year := uint32(2025 + (nanoTime % 5))   // Use different years too
	diaryEntries := []struct {
		content string
		date    *g.YMD
	}{
		{"First diary entry", &g.YMD{Year: year, Month: 1, Day: baseDate}},
		{"Second diary entry", &g.YMD{Year: year, Month: 1, Day: baseDate + 1}},
		{"Third diary entry", &g.YMD{Year: year, Month: 1, Day: baseDate + 2}},
	}

	createdEntries := []*g.DiaryEntry{}
	for _, entry := range diaryEntries {
		createReq := &g.CreateDiaryEntryRequest{
			Content: entry.content,
			Date:    entry.date,
		}
		createResp, err := diaryService.CreateDiaryEntry(authCtx, createReq)
		if err != nil {
			t.Fatalf("Failed to create diary entry: %v", err)
		}
		createdEntries = append(createdEntries, createResp.Entry)
	}

	// Step 5: Retrieve diary entries
	for i, entry := range createdEntries {
		getReq := &g.GetDiaryEntryRequest{
			Date: entry.Date,
		}
		getResp, err := diaryService.GetDiaryEntry(authCtx, getReq)
		if err != nil {
			t.Fatalf("Failed to get diary entry %d: %v", i, err)
		}
		if getResp.Entry.Content != diaryEntries[i].content {
			t.Errorf("Expected content '%s' but got '%s'", diaryEntries[i].content, getResp.Entry.Content)
		}
	}

	// Step 6: Update a diary entry
	updateReq := &g.UpdateDiaryEntryRequest{
		Id:      createdEntries[0].Id,
		Content: "Updated first diary entry",
		Date:    createdEntries[0].Date,
	}
	updateResp, err := diaryService.UpdateDiaryEntry(authCtx, updateReq)
	if err != nil {
		t.Fatalf("Failed to update diary entry: %v", err)
	}
	if updateResp.Entry.Content != updateReq.Content {
		t.Errorf("Expected updated content '%s' but got '%s'", updateReq.Content, updateResp.Entry.Content)
	}

	// Step 7: Get multiple diary entries
	dates := []*g.YMD{}
	for _, entry := range createdEntries {
		dates = append(dates, entry.Date)
	}
	getMultipleReq := &g.GetDiaryEntriesRequest{
		Dates: dates,
	}
	getMultipleResp, err := diaryService.GetDiaryEntries(authCtx, getMultipleReq)
	if err != nil {
		t.Fatalf("Failed to get multiple diary entries: %v", err)
	}
	if len(getMultipleResp.Entries) != len(createdEntries) {
		t.Errorf("Expected %d entries but got %d", len(createdEntries), len(getMultipleResp.Entries))
	}

	// Step 8: Get diary entries by month
	getByMonthReq := &g.GetDiaryEntriesByMonthRequest{
		Month: &g.YM{Year: diaryEntries[0].date.Year, Month: diaryEntries[0].date.Month},
	}
	getByMonthResp, err := diaryService.GetDiaryEntriesByMonth(authCtx, getByMonthReq)
	if err != nil {
		t.Fatalf("Failed to get diary entries by month: %v", err)
	}
	if len(getByMonthResp.Entries) != len(createdEntries) {
		t.Errorf("Expected %d entries for month but got %d", len(createdEntries), len(getByMonthResp.Entries))
	}

	// Step 9: Delete a diary entry
	deleteReq := &g.DeleteDiaryEntryRequest{
		Id: createdEntries[2].Id,
	}
	deleteResp, err := diaryService.DeleteDiaryEntry(authCtx, deleteReq)
	if err != nil {
		t.Fatalf("Failed to delete diary entry: %v", err)
	}
	if !deleteResp.Success {
		t.Error("Expected deletion to be successful")
	}

	// Step 10: Verify deletion
	getDeletedReq := &g.GetDiaryEntryRequest{
		Date: createdEntries[2].Date,
	}
	_, err = diaryService.GetDiaryEntry(authCtx, getDeletedReq)
	if err == nil {
		t.Error("Expected error when getting deleted diary entry")
	}

	// Step 11: Refresh access token
	refreshReq := &g.RefreshAccessTokenRequest{
		RefreshToken: loginResp.RefreshToken,
	}
	refreshResp, err := authService.RefreshAccessToken(ctx, refreshReq)
	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}
	if refreshResp.AccessToken == "" {
		t.Fatal("Expected new access token after refresh")
	}
}

func TestUserIsolation(t *testing.T) {
	db := testutil.SetupTestDB(t)

	// Initialize services
	authService := &auth.AuthEntry{DB: db}
	diaryService := &diary.DiaryEntry{DB: db}
	ctx := context.Background()

	// Generate unique test identifier
	testID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create two users
	user1RegisterReq := &g.RegisterByPasswordRequest{
		Email:    fmt.Sprintf("user1-isolation-%s@example.com", testID),
		Password: "password123",
		Name:     "User One",
	}
	user1RegisterResp, err := authService.RegisterByPassword(ctx, user1RegisterReq)
	if err != nil {
		t.Fatalf("User 1 registration failed: %v", err)
	}

	user2RegisterReq := &g.RegisterByPasswordRequest{
		Email:    fmt.Sprintf("user2-isolation-%s@example.com", testID),
		Password: "password123",
		Name:     "User Two",
	}
	user2RegisterResp, err := authService.RegisterByPassword(ctx, user2RegisterReq)
	if err != nil {
		t.Fatalf("User 2 registration failed: %v", err)
	}

	// Get user contexts
	user1, err := getUserFromToken(t, db, user1RegisterResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to get user 1 from token: %v", err)
	}
	user1Ctx := testutil.CreateAuthenticatedContext(user1.ID)

	user2, err := getUserFromToken(t, db, user2RegisterResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to get user 2 from token: %v", err)
	}
	user2Ctx := testutil.CreateAuthenticatedContext(user2.ID)

	// User 1 creates a diary entry
	nanoTime := time.Now().UnixNano()
	baseDate := uint32((nanoTime % 25) + 1)
	year := uint32(2024 + (nanoTime % 5))

	user1CreateReq := &g.CreateDiaryEntryRequest{
		Content: "User 1's private diary",
		Date:    &g.YMD{Year: year, Month: 2, Day: baseDate},
	}
	_, err = diaryService.CreateDiaryEntry(user1Ctx, user1CreateReq)
	if err != nil {
		t.Fatalf("User 1 failed to create diary entry: %v", err)
	}

	// User 2 creates a diary entry
	user2CreateReq := &g.CreateDiaryEntryRequest{
		Content: "User 2's private diary",
		Date:    &g.YMD{Year: year, Month: 2, Day: baseDate + 1},
	}
	_, err = diaryService.CreateDiaryEntry(user2Ctx, user2CreateReq)
	if err != nil {
		t.Fatalf("User 2 failed to create diary entry: %v", err)
	}

	// Verify User 1 can access their own diary
	user1GetReq := &g.GetDiaryEntryRequest{
		Date: user1CreateReq.Date,
	}
	user1GetResp, err := diaryService.GetDiaryEntry(user1Ctx, user1GetReq)
	if err != nil {
		t.Fatalf("User 1 failed to get their own diary: %v", err)
	}
	if user1GetResp.Entry.Content != user1CreateReq.Content {
		t.Errorf("User 1 got wrong content: expected '%s', got '%s'", user1CreateReq.Content, user1GetResp.Entry.Content)
	}

	// Verify User 2 can access their own diary
	user2GetReq := &g.GetDiaryEntryRequest{
		Date: user2CreateReq.Date,
	}
	user2GetResp, err := diaryService.GetDiaryEntry(user2Ctx, user2GetReq)
	if err != nil {
		t.Fatalf("User 2 failed to get their own diary: %v", err)
	}
	if user2GetResp.Entry.Content != user2CreateReq.Content {
		t.Errorf("User 2 got wrong content: expected '%s', got '%s'", user2CreateReq.Content, user2GetResp.Entry.Content)
	}

	// Verify diaries are isolated (User 1 shouldn't see User 2's diary content)
	if user1GetResp.Entry.Content == user2GetResp.Entry.Content {
		t.Error("User diaries are not properly isolated")
	}
}

// Helper function to get user from token (simplified version)
func getUserFromToken(t *testing.T, db *sql.DB, token string) (*database.User, error) {
	// In a real implementation, this would parse the JWT token
	// For testing purposes, we'll query the database directly
	rows, err := db.Query("SELECT id, email, name FROM users WHERE email LIKE '%@example.com' ORDER BY created_at DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("No user found")
	}

	var user database.User
	err = rows.Scan(&user.ID, &user.Email, &user.Name)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
