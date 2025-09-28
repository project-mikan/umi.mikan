package ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (rueidis.Client, func()) {
	// miniredisでテスト用Redisサーバーを起動
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// rueidisクライアントを作成（テスト用にキャッシュを無効化）
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	require.NoError(t, err)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestRedisRateLimiter_IsAllowed(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	rateLimiter := NewRedisRateLimiter(redisClient)
	ctx := context.Background()

	tests := []struct {
		name        string
		key         string
		limit       int
		window      time.Duration
		attempts    int
		expectError bool
	}{
		{
			name:        "正常系：制限内のリクエスト",
			key:         "test_key_1",
			limit:       5,
			window:      time.Minute,
			attempts:    3,
			expectError: false,
		},
		{
			name:        "異常系：制限を超えるリクエスト",
			key:         "test_key_2",
			limit:       2,
			window:      time.Minute,
			attempts:    5,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lastAllowed bool
			var lastRemaining int

			for i := 0; i < tt.attempts; i++ {
				allowed, remaining, resetTime, err := rateLimiter.IsAllowed(ctx, tt.key, tt.limit, tt.window)
				require.NoError(t, err)

				lastAllowed = allowed
				lastRemaining = remaining

				if !allowed {
					assert.True(t, resetTime > 0, "リセット時間が設定されているべき")
					break
				}
			}

			if tt.expectError {
				assert.False(t, lastAllowed, "制限を超えた場合は許可されないべき")
				assert.Equal(t, 0, lastRemaining, "制限を超えた場合は残り回数は0であるべき")
			} else {
				assert.True(t, lastAllowed, "制限内の場合は許可されるべき")
				assert.True(t, lastRemaining >= 0, "残り回数は0以上であるべき")
			}
		})
	}
}

func TestRedisRateLimiter_Reset(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	rateLimiter := NewRedisRateLimiter(redisClient)
	ctx := context.Background()

	key := "test_reset_key"
	limit := 2
	window := time.Minute

	// 制限まで使い切る
	for i := 0; i < limit; i++ {
		allowed, _, _, err := rateLimiter.IsAllowed(ctx, key, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed, "制限内であれば許可されるべき")
	}

	// 制限を超えることを確認
	allowed, _, _, err := rateLimiter.IsAllowed(ctx, key, limit, window)
	require.NoError(t, err)
	assert.False(t, allowed, "制限を超えた場合は許可されないべき")

	// リセットを実行
	err = rateLimiter.Reset(ctx, key)
	require.NoError(t, err)

	// リセット後は再度許可されることを確認
	allowed, remaining, _, err := rateLimiter.IsAllowed(ctx, key, limit, window)
	require.NoError(t, err)
	assert.True(t, allowed, "リセット後は許可されるべき")
	assert.Equal(t, limit-1, remaining, "リセット後の残り回数が正しいべき")
}

func TestLoginAttemptLimiter_CheckAttempt(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	rateLimiter := NewRedisRateLimiter(redisClient)
	loginLimiter := NewLoginAttemptLimiter(rateLimiter, 3, time.Minute)
	ctx := context.Background()

	identifier := "test_user"

	// 制限内のログイン試行
	for i := 0; i < 3; i++ {
		allowed, remaining, _, err := loginLimiter.CheckAttempt(ctx, identifier)
		require.NoError(t, err)
		assert.True(t, allowed, "制限内であれば許可されるべき")
		assert.Equal(t, 3-i-1, remaining, "残り回数が正しく計算されるべき")
	}

	// 制限を超えるログイン試行
	allowed, remaining, resetTime, err := loginLimiter.CheckAttempt(ctx, identifier)
	require.NoError(t, err)
	assert.False(t, allowed, "制限を超えた場合は許可されないべき")
	assert.Equal(t, 0, remaining, "制限を超えた場合は残り回数は0であるべき")
	assert.True(t, resetTime > 0, "リセット時間が設定されているべき")
}

func TestLoginAttemptLimiter_ResetAttempts(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	rateLimiter := NewRedisRateLimiter(redisClient)
	loginLimiter := NewLoginAttemptLimiter(rateLimiter, 2, time.Minute)
	ctx := context.Background()

	identifier := "test_user_reset"

	// 制限まで使い切る
	for i := 0; i < 2; i++ {
		allowed, _, _, err := loginLimiter.CheckAttempt(ctx, identifier)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// 制限を超えることを確認
	allowed, _, _, err := loginLimiter.CheckAttempt(ctx, identifier)
	require.NoError(t, err)
	assert.False(t, allowed)

	// リセットを実行
	err = loginLimiter.ResetAttempts(ctx, identifier)
	require.NoError(t, err)

	// リセット後は再度許可されることを確認
	allowed, remaining, _, err := loginLimiter.CheckAttempt(ctx, identifier)
	require.NoError(t, err)
	assert.True(t, allowed, "リセット後は許可されるべき")
	assert.Equal(t, 1, remaining, "リセット後の残り回数が正しいべき")
}

func TestRedisRateLimiter_SlidingWindow(t *testing.T) {
	redisClient, cleanup := setupTestRedis(t)
	defer cleanup()

	rateLimiter := NewRedisRateLimiter(redisClient)
	ctx := context.Background()

	key := "test_sliding_window"
	limit := 3
	window := 2 * time.Second

	// 制限まで使い切る
	for i := 0; i < limit; i++ {
		allowed, _, _, err := rateLimiter.IsAllowed(ctx, key, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// 制限を超えることを確認
	allowed, _, _, err := rateLimiter.IsAllowed(ctx, key, limit, window)
	require.NoError(t, err)
	assert.False(t, allowed)

	// ウィンドウ期間を超えて待機
	time.Sleep(2100 * time.Millisecond)

	// 古いエントリが削除され、再度許可されることを確認
	allowed, remaining, _, err := rateLimiter.IsAllowed(ctx, key, limit, window)
	require.NoError(t, err)
	assert.True(t, allowed, "ウィンドウ期間後は再度許可されるべき")
	assert.Equal(t, limit-1, remaining, "ウィンドウ期間後の残り回数が正しいべき")
}