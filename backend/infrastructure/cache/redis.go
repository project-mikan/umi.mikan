package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client instance
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

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetClient returns the underlying Redis client for specific operations
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}
