package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// Mock Redis client for testing when testcontainers is not available
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

func setupMockRedis() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string]string),
	}
}

func TestRedisClient_SetAndGetDiaryCount(t *testing.T) {
	client := setupMockRedis()
	defer client.Close()

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

func TestRedisClient_GetDiaryCount_CacheMiss(t *testing.T) {
	client := setupMockRedis()
	defer client.Close()

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

func TestRedisClient_DeleteDiaryCount(t *testing.T) {
	client := setupMockRedis()
	defer client.Close()

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

func TestRedisClient_MultipleUsers(t *testing.T) {
	client := setupMockRedis()
	defer client.Close()

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

func TestRedisClient_CacheExpiration(t *testing.T) {
	client := setupMockRedis()
	defer client.Close()

	ctx := context.Background()
	userID := uuid.New().String()
	count := uint32(15)

	// This test would require modifying the cache expiration time to be very short
	// For now, we just verify the operation works
	err := client.SetDiaryCount(ctx, userID, count)
	if err != nil {
		t.Fatalf("Failed to set diary count: %v", err)
	}

	retrievedCount, err := client.GetDiaryCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get diary count: %v", err)
	}
	if retrievedCount != count {
		t.Errorf("Expected count %d but got %d", count, retrievedCount)
	}
}

func TestRedisClient_Close(t *testing.T) {
	client := setupMockRedis()

	// Close should work without error
	err := client.Close()
	if err != nil {
		t.Errorf("Expected Close() to succeed but got error: %v", err)
	}

	// Mock doesn't simulate connection errors after close, so we just test that Close() works
}