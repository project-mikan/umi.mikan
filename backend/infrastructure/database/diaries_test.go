package database

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Use in-memory SQLite for testing to avoid circular import
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS diaries (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			content TEXT,
			date TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, date)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create diaries table: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	})

	return db
}

func createTestUser(t *testing.T, db *sql.DB) uuid.UUID {
	userID := uuid.New()
	_, err := db.Exec("INSERT INTO users (id, email, name) VALUES (?, ?, ?)",
		userID.String(), "diaries-test@example.com", "Diaries Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

func TestCountDiariesByUserID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Test with no diaries
	count, err := CountDiariesByUserID(ctx, db, userID.String())
	if err != nil {
		t.Fatalf("Failed to count diaries for empty user: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for user with no diaries but got %d", count)
	}

	// Create some diary entries
	diaryEntries := []struct {
		content string
		date    string
	}{
		{"First diary entry", "2024-01-01"},
		{"Second diary entry", "2024-01-02"},
		{"Third diary entry", "2024-01-03"},
	}

	for _, entry := range diaryEntries {
		_, err := db.ExecContext(ctx,
			"INSERT INTO diaries (id, user_id, content, date) VALUES ($1, $2, $3, $4)",
			uuid.New(), userID, entry.content, entry.date,
		)
		if err != nil {
			t.Fatalf("Failed to create diary entry: %v", err)
		}
	}

	// Test count with diary entries
	count, err = CountDiariesByUserID(ctx, db, userID.String())
	if err != nil {
		t.Fatalf("Failed to count diaries: %v", err)
	}
	if count != len(diaryEntries) {
		t.Errorf("Expected count %d but got %d", len(diaryEntries), count)
	}
}

func TestCountDiariesByUserID_MultipleUsers(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create two users
	userID1 := createTestUser(t, db)
	userID2 := uuid.New()
	_, err := db.Exec("INSERT INTO users (id, email, name) VALUES (?, ?, ?)",
		userID2.String(), "diaries-test2@example.com", "Test User 2")
	if err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	// User 1 creates 3 diary entries
	for i := 1; i <= 3; i++ {
		_, err := db.ExecContext(ctx,
			"INSERT INTO diaries (id, user_id, content, date) VALUES ($1, $2, $3, $4)",
			uuid.New(), userID1, "User 1 diary", "2024-02-0"+string(rune(48+i)),
		)
		if err != nil {
			t.Fatalf("Failed to create diary entry for user 1: %v", err)
		}
	}

	// User 2 creates 2 diary entries
	for i := 1; i <= 2; i++ {
		_, err := db.ExecContext(ctx,
			"INSERT INTO diaries (id, user_id, content, date) VALUES ($1, $2, $3, $4)",
			uuid.New(), userID2, "User 2 diary", "2024-03-0"+string(rune(48+i)),
		)
		if err != nil {
			t.Fatalf("Failed to create diary entry for user 2: %v", err)
		}
	}

	// Check count for user 1
	count1, err := CountDiariesByUserID(ctx, db, userID1.String())
	if err != nil {
		t.Fatalf("Failed to count diaries for user 1: %v", err)
	}
	if count1 != 3 {
		t.Errorf("Expected count 3 for user 1 but got %d", count1)
	}

	// Check count for user 2
	count2, err := CountDiariesByUserID(ctx, db, userID2.String())
	if err != nil {
		t.Fatalf("Failed to count diaries for user 2: %v", err)
	}
	if count2 != 2 {
		t.Errorf("Expected count 2 for user 2 but got %d", count2)
	}
}

func TestCountDiariesByUserID_NonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Test with non-existent user
	nonExistentUserID := uuid.New().String()
	count, err := CountDiariesByUserID(ctx, db, nonExistentUserID)
	if err != nil {
		t.Fatalf("Failed to count diaries for non-existent user: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for non-existent user but got %d", count)
	}
}

func TestCountDiariesByUserID_InvalidUserID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Test with invalid user ID format
	invalidUserID := "invalid-user-id"
	count, err := CountDiariesByUserID(ctx, db, invalidUserID)
	if err != nil {
		t.Fatalf("Failed to count diaries for invalid user ID: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 for invalid user ID but got %d", count)
	}
}

func TestCountDiariesByUserID_DatabaseError(t *testing.T) {
	// Test with closed database connection
	db := setupTestDB(t)
	ctx := context.Background()

	userID := createTestUser(t, db)

	// Close the database connection to simulate an error
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	_, err := CountDiariesByUserID(ctx, db, userID.String())
	if err == nil {
		t.Error("Expected error with closed database but got nil")
	}
}
