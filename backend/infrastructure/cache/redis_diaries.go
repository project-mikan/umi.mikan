package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache keys for diary-related operations
const (
	DiaryCountCacheKey = "diary_count:%s" // %s = userID
)

// Cache expiration times
const (
	DiaryCountCacheExpiration = 24 * time.Hour // 24時間キャッシュ
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

// DeleteDiaryCount deletes the diary count cache for a user
func (r *RedisClient) DeleteDiaryCount(ctx context.Context, userID string) error {
	key := fmt.Sprintf(DiaryCountCacheKey, userID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete diary count cache: %w", err)
	}
	return nil
}
