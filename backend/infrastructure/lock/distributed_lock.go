package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/rueidis"
)

// DistributedLockInterface defines the interface for distributed locks
type DistributedLockInterface interface {
	TryLock(ctx context.Context) (bool, error)
	Unlock(ctx context.Context) error
	Extend(ctx context.Context, newDuration time.Duration) error
	IsLocked(ctx context.Context) (bool, error)
	IsOwnedByMe(ctx context.Context) (bool, error)
}

// DistributedLock represents a Redis-based distributed lock
type DistributedLock struct {
	client   rueidis.Client
	key      string
	value    string
	duration time.Duration
}

// NewDistributedLock creates a new distributed lock
func NewDistributedLock(client rueidis.Client, key string, duration time.Duration) *DistributedLock {
	// Generate a unique value for this lock instance
	value := generateUniqueValue()

	return &DistributedLock{
		client:   client,
		key:      key,
		value:    value,
		duration: duration,
	}
}

// TryLock attempts to acquire the lock
// Returns true if lock is acquired, false if already locked by another process
func (dl *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	// Use SET key value NX EX seconds command
	// NX: only set if key does not exist
	// EX: set expiration time in seconds
	cmd := dl.client.B().Set().Key(dl.key).Value(dl.value).Nx().Ex(dl.duration).Build()
	result := dl.client.Do(ctx, cmd)

	if result.Error() != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", result.Error())
	}

	// If the result is "OK", lock was acquired
	response, err := result.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to parse lock result: %w", err)
	}

	return response == "OK", nil
}

// Unlock releases the lock if it's held by this instance
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	// Lua script to ensure we only delete the lock if we own it
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	cmd := dl.client.B().Eval().Script(script).Numkeys(1).Key(dl.key).Arg(dl.value).Build()
	result := dl.client.Do(ctx, cmd)

	if result.Error() != nil {
		return fmt.Errorf("failed to unlock: %w", result.Error())
	}

	return nil
}

// Extend extends the lock duration if it's held by this instance
func (dl *DistributedLock) Extend(ctx context.Context, newDuration time.Duration) error {
	// Lua script to extend expiration only if we own the lock
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	cmd := dl.client.B().Eval().Script(script).Numkeys(1).Key(dl.key).Arg(dl.value).Arg(fmt.Sprintf("%d", int(newDuration.Seconds()))).Build()
	result := dl.client.Do(ctx, cmd)

	if result.Error() != nil {
		return fmt.Errorf("failed to extend lock: %w", result.Error())
	}

	return nil
}

// IsLocked checks if the lock exists (regardless of owner)
func (dl *DistributedLock) IsLocked(ctx context.Context) (bool, error) {
	cmd := dl.client.B().Exists().Key(dl.key).Build()
	result := dl.client.Do(ctx, cmd)

	if result.Error() != nil {
		return false, fmt.Errorf("failed to check lock existence: %w", result.Error())
	}

	exists, err := result.AsInt64()
	if err != nil {
		return false, fmt.Errorf("failed to parse exists result: %w", err)
	}

	return exists > 0, nil
}

// IsOwnedByMe checks if the lock is owned by this instance
func (dl *DistributedLock) IsOwnedByMe(ctx context.Context) (bool, error) {
	cmd := dl.client.B().Get().Key(dl.key).Build()
	result := dl.client.Do(ctx, cmd)

	if result.Error() != nil {
		// If key doesn't exist, it's not owned by anyone
		if rueidis.IsRedisNil(result.Error()) {
			return false, nil
		}
		return false, fmt.Errorf("failed to get lock value: %w", result.Error())
	}

	value, err := result.ToString()
	if err != nil {
		return false, fmt.Errorf("failed to parse lock value: %w", err)
	}

	return value == dl.value, nil
}

// generateUniqueValue generates a unique value for the lock
func generateUniqueValue() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if random generation fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// Helper functions for creating lock keys

// DailySummaryLockKey creates a lock key for daily summary generation
func DailySummaryLockKey(userID, date string) string {
	return fmt.Sprintf("summary_lock:daily:%s:%s", userID, date)
}

// MonthlySummaryLockKey creates a lock key for monthly summary generation
func MonthlySummaryLockKey(userID string, year, month int) string {
	return fmt.Sprintf("summary_lock:monthly:%s:%d:%d", userID, year, month)
}
