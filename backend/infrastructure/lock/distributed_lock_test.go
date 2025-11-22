package lock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

// setupRedisClient はテスト用のRedisクライアントをセットアップ
func setupRedisClient(t *testing.T) rueidis.Client {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"redis:6379"},
	})
	if err != nil {
		t.Fatalf("failed to create Redis client: %v", err)
	}
	return client
}

// cleanupRedisKey はテスト用のRedisキーをクリーンアップ
func cleanupRedisKey(ctx context.Context, client rueidis.Client, key string) {
	cmd := client.B().Del().Key(key).Build()
	client.Do(ctx, cmd)
}

// TestNewDistributedLock tests the NewDistributedLock function
func TestNewDistributedLock(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		duration time.Duration
	}{
		{
			name:     "正常なロックの作成",
			key:      "test-lock-1",
			duration: 10 * time.Second,
		},
		{
			name:     "異なる期間でのロック作成",
			key:      "test-lock-2",
			duration: 1 * time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := setupRedisClient(t)
			defer client.Close()

			lock := NewDistributedLock(client, tc.key, tc.duration)

			assert.NotNil(t, lock)
			assert.Equal(t, tc.key, lock.key)
			assert.Equal(t, tc.duration, lock.duration)
			assert.NotEmpty(t, lock.value)
		})
	}
}

// TestTryLock tests the TryLock function
func TestTryLock(t *testing.T) {
	testCases := []struct {
		name            string
		setupLock       func(ctx context.Context, client rueidis.Client, key string)
		expectedAcquire bool
	}{
		{
			name: "ロックが取得できる場合",
			setupLock: func(ctx context.Context, client rueidis.Client, key string) {
				// 何もしない（ロックは存在しない）
			},
			expectedAcquire: true,
		},
		{
			name: "既にロックが存在する場合",
			setupLock: func(ctx context.Context, client rueidis.Client, key string) {
				// 既存のロックを作成
				cmd := client.B().Set().Key(key).Value("existing-lock").Ex(10 * time.Second).Build()
				client.Do(ctx, cmd)
			},
			expectedAcquire: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client := setupRedisClient(t)
			defer client.Close()

			key := fmt.Sprintf("test-lock-%d", time.Now().UnixNano())
			defer cleanupRedisKey(ctx, client, key)

			tc.setupLock(ctx, client, key)

			lock := NewDistributedLock(client, key, 10*time.Second)
			acquired, err := lock.TryLock(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAcquire, acquired)
		})
	}
}

// TestUnlock tests the Unlock function
func TestUnlock(t *testing.T) {
	testCases := []struct {
		name      string
		setupLock func(ctx context.Context, lock *DistributedLock) error
	}{
		{
			name: "自分が所有するロックを解放",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				_, err := lock.TryLock(ctx)
				return err
			},
		},
		{
			name: "所有していないロックを解放しようとする",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				// 別のvalueでロックを取得
				cmd := lock.client.B().Set().Key(lock.key).Value("different-value").Ex(10 * time.Second).Build()
				result := lock.client.Do(ctx, cmd)
				return result.Error()
			},
		},
		{
			name: "存在しないロックを解放しようとする",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				// 何もしない
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client := setupRedisClient(t)
			defer client.Close()

			key := fmt.Sprintf("test-lock-%d", time.Now().UnixNano())
			defer cleanupRedisKey(ctx, client, key)

			lock := NewDistributedLock(client, key, 10*time.Second)
			err := tc.setupLock(ctx, lock)
			assert.NoError(t, err)

			err = lock.Unlock(ctx)
			assert.NoError(t, err)
		})
	}
}

// TestExtend tests the Extend function
func TestExtend(t *testing.T) {
	testCases := []struct {
		name        string
		setupLock   func(ctx context.Context, lock *DistributedLock) error
		newDuration time.Duration
	}{
		{
			name: "自分が所有するロックの期間を延長",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				_, err := lock.TryLock(ctx)
				return err
			},
			newDuration: 20 * time.Second,
		},
		{
			name: "所有していないロックの期間を延長しようとする",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				// 別のvalueでロックを取得
				cmd := lock.client.B().Set().Key(lock.key).Value("different-value").Ex(10 * time.Second).Build()
				result := lock.client.Do(ctx, cmd)
				return result.Error()
			},
			newDuration: 20 * time.Second,
		},
		{
			name: "存在しないロックの期間を延長しようとする",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				// 何もしない
				return nil
			},
			newDuration: 20 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client := setupRedisClient(t)
			defer client.Close()

			key := fmt.Sprintf("test-lock-%d", time.Now().UnixNano())
			defer cleanupRedisKey(ctx, client, key)

			lock := NewDistributedLock(client, key, 10*time.Second)
			err := tc.setupLock(ctx, lock)
			assert.NoError(t, err)

			err = lock.Extend(ctx, tc.newDuration)
			assert.NoError(t, err)
		})
	}
}

// TestIsLocked tests the IsLocked function
func TestIsLocked(t *testing.T) {
	testCases := []struct {
		name           string
		setupLock      func(ctx context.Context, client rueidis.Client, key string)
		expectedLocked bool
	}{
		{
			name: "ロックが存在する場合",
			setupLock: func(ctx context.Context, client rueidis.Client, key string) {
				cmd := client.B().Set().Key(key).Value("some-value").Ex(10 * time.Second).Build()
				client.Do(ctx, cmd)
			},
			expectedLocked: true,
		},
		{
			name: "ロックが存在しない場合",
			setupLock: func(ctx context.Context, client rueidis.Client, key string) {
				// 何もしない
			},
			expectedLocked: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client := setupRedisClient(t)
			defer client.Close()

			key := fmt.Sprintf("test-lock-%d", time.Now().UnixNano())
			defer cleanupRedisKey(ctx, client, key)

			tc.setupLock(ctx, client, key)

			lock := NewDistributedLock(client, key, 10*time.Second)
			locked, err := lock.IsLocked(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedLocked, locked)
		})
	}
}

// TestIsOwnedByMe tests the IsOwnedByMe function
func TestIsOwnedByMe(t *testing.T) {
	testCases := []struct {
		name          string
		setupLock     func(ctx context.Context, lock *DistributedLock) error
		expectedOwned bool
	}{
		{
			name: "自分が所有するロック",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				_, err := lock.TryLock(ctx)
				return err
			},
			expectedOwned: true,
		},
		{
			name: "他者が所有するロック",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				cmd := lock.client.B().Set().Key(lock.key).Value("different-value").Ex(10 * time.Second).Build()
				result := lock.client.Do(ctx, cmd)
				return result.Error()
			},
			expectedOwned: false,
		},
		{
			name: "存在しないロック",
			setupLock: func(ctx context.Context, lock *DistributedLock) error {
				// 何もしない
				return nil
			},
			expectedOwned: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client := setupRedisClient(t)
			defer client.Close()

			key := fmt.Sprintf("test-lock-%d", time.Now().UnixNano())
			defer cleanupRedisKey(ctx, client, key)

			lock := NewDistributedLock(client, key, 10*time.Second)
			err := tc.setupLock(ctx, lock)
			assert.NoError(t, err)

			owned, err := lock.IsOwnedByMe(ctx)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOwned, owned)
		})
	}
}

// TestDailySummaryLockKey tests the DailySummaryLockKey function
func TestDailySummaryLockKey(t *testing.T) {
	testCases := []struct {
		name     string
		userID   string
		date     string
		expected string
	}{
		{
			name:     "正常なキーの生成",
			userID:   "user-123",
			date:     "2025-01-15",
			expected: "summary_lock:daily:user-123:2025-01-15",
		},
		{
			name:     "異なるユーザーIDとdate",
			userID:   "user-456",
			date:     "2025-02-20",
			expected: "summary_lock:daily:user-456:2025-02-20",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DailySummaryLockKey(tc.userID, tc.date)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestMonthlySummaryLockKey tests the MonthlySummaryLockKey function
func TestMonthlySummaryLockKey(t *testing.T) {
	testCases := []struct {
		name     string
		userID   string
		year     int
		month    int
		expected string
	}{
		{
			name:     "正常なキーの生成",
			userID:   "user-123",
			year:     2025,
			month:    1,
			expected: "summary_lock:monthly:user-123:2025:1",
		},
		{
			name:     "異なるユーザーIDと年月",
			userID:   "user-456",
			year:     2025,
			month:    12,
			expected: "summary_lock:monthly:user-456:2025:12",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MonthlySummaryLockKey(tc.userID, tc.year, tc.month)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGenerateUniqueValue tests the generateUniqueValue function
func TestGenerateUniqueValue(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "ユニークな値の生成",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value1 := generateUniqueValue()
			value2 := generateUniqueValue()

			assert.NotEmpty(t, value1)
			assert.NotEmpty(t, value2)
			// 2つの値は異なるべき
			assert.NotEqual(t, value1, value2)
		})
	}
}
