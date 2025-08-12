package cache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// Cache keys for diary-related operations
const (
	DiaryCountCacheKey = "diary_count:%s" // %s = userID
)

// Cache expiration times
const (
	DiaryCountCacheExpiration = 0 // 無期限キャッシュ
)

// SetDiaryCount sets the diary count for a user in cache
func (r *RedisClient) SetDiaryCount(ctx context.Context, userID string, count uint32) error {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)
	err := r.client.Set(ctx, key, count, DiaryCountCacheExpiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set diary count cache: %w", err)
	}
	return nil
}

// GetDiaryCount gets the diary count for a user from cache
func (r *RedisClient) GetDiaryCount(ctx context.Context, userID string) (uint32, error) {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("cache miss")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get diary count cache: %w", err)
	}

	count, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse cached count: %w", err)
	}

	return uint32(count), nil
}

// UpdateDiaryCount updates (increments or decrements) the diary count for a user in cache
func (r *RedisClient) UpdateDiaryCount(ctx context.Context, userID string, delta int) error {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)

	// Use INCRBY to atomically increment/decrement the count
	_, err := r.client.IncrBy(ctx, key, int64(delta)).Result()
	if err != nil {
		return fmt.Errorf("failed to update diary count cache: %w", err)
	}

	// Set expiration to 0 (no expiration) if the key doesn't have one
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check cache key existence: %w", err)
	}

	if exists > 0 {
		// Set expiration to 0 (no expiration)
		err := r.client.Persist(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("failed to set cache key persistence: %w", err)
		}
	}

	return nil
}

// DeleteDiaryCount deletes the diary count cache for a user
func (r *RedisClient) DeleteDiaryCount(ctx context.Context, userID string) error {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete diary count cache: %w", err)
	}
	return nil
}
