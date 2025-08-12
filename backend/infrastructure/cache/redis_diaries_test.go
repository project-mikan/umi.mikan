package cache

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
)

// Mock Redis client for testing diary cache operations
type mockRedisClientForDiaries struct {
	data map[string]string
}

func (m *mockRedisClientForDiaries) SetDiaryCount(ctx context.Context, userID string, count uint32) error {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)
	m.data[key] = fmt.Sprintf("%d", count)
	return nil
}

func (m *mockRedisClientForDiaries) GetDiaryCount(ctx context.Context, userID string) (uint32, error) {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)
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

func (m *mockRedisClientForDiaries) UpdateDiaryCount(ctx context.Context, userID string, delta int) error {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)

	// Get current value or start with 0
	var currentCount int64 = 0
	if val, exists := m.data[key]; exists {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			currentCount = parsed
		}
	}

	// Update the count
	newCount := currentCount + int64(delta)
	if newCount < 0 {
		newCount = 0 // Ensure count doesn't go negative
	}

	m.data[key] = strconv.FormatInt(newCount, 10)
	return nil
}

func (m *mockRedisClientForDiaries) DeleteDiaryCount(ctx context.Context, userID string) error {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)
	delete(m.data, key)
	return nil
}

func (m *mockRedisClientForDiaries) Close() error {
	return nil
}

func setupMockRedisForDiaries() *mockRedisClientForDiaries {
	return &mockRedisClientForDiaries{
		data: make(map[string]string),
	}
}

func TestRedisClient_DiaryCount_SetAndGet(t *testing.T) {
	client := setupMockRedisForDiaries()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}
	}()

	ctx := context.Background()
	userID := uuid.New().String()
	expectedCount := uint32(5)

	// Test setting diary count
	err := client.SetDiaryCount(ctx, userID, expectedCount)
	if err != nil {
		t.Fatalf("Failed to set diary count: %v", err)
	}

	// Test getting diary count
	count, err := client.GetDiaryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get diary count: %v", err)
	}

	if count != expectedCount {
		t.Errorf("Expected count %d but got %d", expectedCount, count)
	}
}

func TestRedisClient_DiaryCount_CacheMiss(t *testing.T) {
	client := setupMockRedisForDiaries()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}
	}()

	ctx := context.Background()
	userID := uuid.New().String()

	// Test cache miss
	_, err := client.GetDiaryCount(ctx, userID)
	if err == nil {
		t.Error("Expected cache miss error but got nil")
	}
	if err.Error() != "cache miss" {
		t.Errorf("Expected 'cache miss' error but got: %v", err)
	}
}

func TestRedisClient_DiaryCount_Delete(t *testing.T) {
	client := setupMockRedisForDiaries()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}
	}()

	ctx := context.Background()
	userID := uuid.New().String()
	count := uint32(10)

	// Set diary count first
	err := client.SetDiaryCount(ctx, userID, count)
	if err != nil {
		t.Fatalf("Failed to set diary count: %v", err)
	}

	// Verify it exists
	retrievedCount, err := client.GetDiaryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get diary count: %v", err)
	}
	if retrievedCount != count {
		t.Errorf("Expected count %d but got %d", count, retrievedCount)
	}

	// Delete diary count
	err = client.DeleteDiaryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to delete diary count: %v", err)
	}

	// Verify it's deleted (should be cache miss)
	_, err = client.GetDiaryCount(ctx, userID)
	if err == nil {
		t.Error("Expected cache miss error after deletion but got nil")
	}
	if err.Error() != "cache miss" {
		t.Errorf("Expected 'cache miss' error after deletion but got: %v", err)
	}
}

func TestRedisClient_DiaryCount_MultipleUsers(t *testing.T) {
	client := setupMockRedisForDiaries()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}
	}()

	ctx := context.Background()
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()
	count1 := uint32(3)
	count2 := uint32(7)

	// Set counts for both users
	err := client.SetDiaryCount(ctx, userID1, count1)
	if err != nil {
		t.Fatalf("Failed to set diary count for user 1: %v", err)
	}

	err = client.SetDiaryCount(ctx, userID2, count2)
	if err != nil {
		t.Fatalf("Failed to set diary count for user 2: %v", err)
	}

	// Verify both users have correct counts
	retrievedCount1, err := client.GetDiaryCount(ctx, userID1)
	if err != nil {
		t.Fatalf("Failed to get diary count for user 1: %v", err)
	}
	if retrievedCount1 != count1 {
		t.Errorf("Expected count %d for user 1 but got %d", count1, retrievedCount1)
	}

	retrievedCount2, err := client.GetDiaryCount(ctx, userID2)
	if err != nil {
		t.Fatalf("Failed to get diary count for user 2: %v", err)
	}
	if retrievedCount2 != count2 {
		t.Errorf("Expected count %d for user 2 but got %d", count2, retrievedCount2)
	}

	// Delete one user's count and verify the other is unaffected
	err = client.DeleteDiaryCount(ctx, userID1)
	if err != nil {
		t.Fatalf("Failed to delete diary count for user 1: %v", err)
	}

	// User 1 should have cache miss
	_, err = client.GetDiaryCount(ctx, userID1)
	if err == nil || err.Error() != "cache miss" {
		t.Errorf("Expected cache miss for user 1 after deletion but got: %v", err)
	}

	// User 2 should still have their count
	retrievedCount2, err = client.GetDiaryCount(ctx, userID2)
	if err != nil {
		t.Fatalf("Failed to get diary count for user 2 after deleting user 1: %v", err)
	}
	if retrievedCount2 != count2 {
		t.Errorf("Expected count %d for user 2 but got %d", count2, retrievedCount2)
	}
}

func TestRedisClient_DiaryCount_Update(t *testing.T) {
	client := setupMockRedisForDiaries()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}
	}()

	ctx := context.Background()
	userID := uuid.New().String()

	// Test updating from 0 (non-existent key)
	err := client.UpdateDiaryCount(ctx, userID, 5)
	if err != nil {
		t.Fatalf("Failed to update diary count from 0: %v", err)
	}

	count, err := client.GetDiaryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get diary count after update: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected count 5 after update but got %d", count)
	}

	// Test incrementing
	err = client.UpdateDiaryCount(ctx, userID, 3)
	if err != nil {
		t.Fatalf("Failed to increment diary count: %v", err)
	}

	count, err = client.GetDiaryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get diary count after increment: %v", err)
	}
	if count != 8 {
		t.Errorf("Expected count 8 after increment but got %d", count)
	}

	// Test decrementing
	err = client.UpdateDiaryCount(ctx, userID, -3)
	if err != nil {
		t.Fatalf("Failed to decrement diary count: %v", err)
	}

	count, err = client.GetDiaryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get diary count after decrement: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected count 5 after decrement but got %d", count)
	}
}

func TestRedisClient_DiaryCount_UpdateToNegative(t *testing.T) {
	client := setupMockRedisForDiaries()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}
	}()

	ctx := context.Background()
	userID := uuid.New().String()

	// Set initial count
	err := client.SetDiaryCount(ctx, userID, 3)
	if err != nil {
		t.Fatalf("Failed to set initial diary count: %v", err)
	}

	// Try to decrement below 0
	err = client.UpdateDiaryCount(ctx, userID, -5)
	if err != nil {
		t.Fatalf("Failed to update diary count to negative: %v", err)
	}

	count, err := client.GetDiaryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get diary count after negative update: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 after negative update but got %d", count)
	}
}

func TestRedisClient_DiaryCount_UpdateMultipleUsers(t *testing.T) {
	client := setupMockRedisForDiaries()
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}
	}()

	ctx := context.Background()
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	// Set initial counts
	err := client.SetDiaryCount(ctx, userID1, 5)
	if err != nil {
		t.Fatalf("Failed to set diary count for user 1: %v", err)
	}

	err = client.SetDiaryCount(ctx, userID2, 10)
	if err != nil {
		t.Fatalf("Failed to set diary count for user 2: %v", err)
	}

	// Update user 1 count
	err = client.UpdateDiaryCount(ctx, userID1, 2)
	if err != nil {
		t.Fatalf("Failed to update diary count for user 1: %v", err)
	}

	// Update user 2 count
	err = client.UpdateDiaryCount(ctx, userID2, -3)
	if err != nil {
		t.Fatalf("Failed to update diary count for user 2: %v", err)
	}

	// Verify both users have correct updated counts
	count1, err := client.GetDiaryCount(ctx, userID1)
	if err != nil {
		t.Fatalf("Failed to get diary count for user 1: %v", err)
	}
	if count1 != 7 {
		t.Errorf("Expected count 7 for user 1 but got %d", count1)
	}

	count2, err := client.GetDiaryCount(ctx, userID2)
	if err != nil {
		t.Fatalf("Failed to get diary count for user 2: %v", err)
	}
	if count2 != 7 {
		t.Errorf("Expected count 7 for user 2 but got %d", count2)
	}
}

func TestDiaryCountCacheKey_Format(t *testing.T) {
	userID := "test-user-123"
	expectedKey := "diary_count:test-user-123"
	actualKey := fmt.Sprintf(DiaryCountCacheKey, userID)

	if actualKey != expectedKey {
		t.Errorf("Expected key '%s' but got '%s'", expectedKey, actualKey)
	}
}
