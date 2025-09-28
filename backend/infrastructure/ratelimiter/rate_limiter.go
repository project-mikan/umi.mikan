package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/rueidis"
)

// RateLimiter Redisベースのレート制限インターフェース
type RateLimiter interface {
	// IsAllowed 指定されたキーに対してアクションが許可されているかチェック
	IsAllowed(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error)
	// Reset 指定されたキーのレート制限をリセット
	Reset(ctx context.Context, key string) error
}

// RedisRateLimiter Redisを使用したレート制限実装
type RedisRateLimiter struct {
	client rueidis.Client
}

// NewRedisRateLimiter 新しいRedisRateLimiterを作成
func NewRedisRateLimiter(client rueidis.Client) RateLimiter {
	return &RedisRateLimiter{
		client: client,
	}
}

// IsAllowed スライディングウィンドウ方式でレート制限をチェック
// 戻り値: (許可されているか, 残り試行回数, リセットまでの時間, エラー)
func (r *RedisRateLimiter) IsAllowed(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error) {
	now := time.Now()
	windowStart := now.Add(-window)
	windowStartUnix := windowStart.UnixNano()
	nowUnix := now.UnixNano()

	// Luaスクリプトでアトミックにスライディングウィンドウレート制限を実行
	script := `
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local ttl = tonumber(ARGV[4])

		-- 古いエントリを削除（ウィンドウの開始時刻より前）
		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

		-- 現在のカウントを取得
		local current = redis.call('ZCARD', key)

		if current < limit then
			-- 制限内の場合、新しいエントリを追加
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, ttl)
			return {1, limit - current - 1, 0}
		else
			-- 制限を超えている場合、最も古いエントリの時刻を取得してリセット時間を計算
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local reset_time = 0
			if #oldest >= 2 then
				reset_time = oldest[2] + (ttl * 1000000000) - now
				if reset_time < 0 then
					reset_time = 0
				end
			end
			return {0, 0, reset_time}
		end
	`

	cmd := r.client.B().Eval().Script(script).Numkeys(1).Key(key).
		Arg(strconv.FormatInt(windowStartUnix, 10)).
		Arg(strconv.FormatInt(nowUnix, 10)).
		Arg(strconv.Itoa(limit)).
		Arg(strconv.Itoa(int(window.Seconds()))).
		Build()

	result, err := r.client.Do(ctx, cmd).AsIntSlice()
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to execute rate limit check: %w", err)
	}

	if len(result) != 3 {
		return false, 0, 0, fmt.Errorf("unexpected result length from Redis script: %d", len(result))
	}

	allowed := result[0] == 1
	remaining := int(result[1])
	resetNanos := result[2]
	resetDuration := time.Duration(resetNanos) * time.Nanosecond

	return allowed, remaining, resetDuration, nil
}

// Reset 指定されたキーのレート制限をリセット
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	cmd := r.client.B().Del().Key(key).Build()
	_, err := r.client.Do(ctx, cmd).AsInt64()
	if err != nil {
		return fmt.Errorf("failed to reset rate limit for key %s: %w", key, err)
	}
	return nil
}

// LoginAttemptLimiter ログイン試行専用のレート制限
type LoginAttemptLimiter struct {
	rateLimiter RateLimiter
	maxAttempts int
	window      time.Duration
}

// NewLoginAttemptLimiter 新しいLoginAttemptLimiterを作成
func NewLoginAttemptLimiter(rateLimiter RateLimiter, maxAttempts int, window time.Duration) *LoginAttemptLimiter {
	return &LoginAttemptLimiter{
		rateLimiter: rateLimiter,
		maxAttempts: maxAttempts,
		window:      window,
	}
}

// CheckAttempt ログイン試行が許可されているかチェック
func (l *LoginAttemptLimiter) CheckAttempt(ctx context.Context, identifier string) (bool, int, time.Duration, error) {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	return l.rateLimiter.IsAllowed(ctx, key, l.maxAttempts, l.window)
}

// ResetAttempts 指定された識別子のログイン試行回数をリセット
func (l *LoginAttemptLimiter) ResetAttempts(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	return l.rateLimiter.Reset(ctx, key)
}