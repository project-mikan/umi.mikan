package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() (*RedisClient, error) {
	config, err := constants.LoadRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Redis config: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
	})

	// 接続テスト
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// 日記の総数をキャッシュに設定
func (r *RedisClient) SetDiaryCount(ctx context.Context, userID string, count uint32) error {
	key := fmt.Sprintf("diary_count:%s", userID)
	err := r.client.Set(ctx, key, count, 24*time.Hour).Err() // 24時間キャッシュ
	if err != nil {
		return fmt.Errorf("failed to set diary count cache: %w", err)
	}
	return nil
}

// 日記の総数をキャッシュから取得
func (r *RedisClient) GetDiaryCount(ctx context.Context, userID string) (uint32, error) {
	key := fmt.Sprintf("diary_count:%s", userID)
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

// 日記の総数キャッシュを削除（日記の作成・削除時に使用）
func (r *RedisClient) DeleteDiaryCount(ctx context.Context, userID string) error {
	key := fmt.Sprintf("diary_count:%s", userID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete diary count cache: %w", err)
	}
	return nil
}
