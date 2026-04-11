package constants

import (
	"os"
	"testing"
	"time"
)

func TestLoadSchedulerConfig(t *testing.T) {
	tests := []struct {
		name                       string
		monthlyInterval            string
		diaryEmbeddingHour         string
		diaryEmbeddingMinute       string
		expectedMonthly            time.Duration
		expectedDiaryEmbeddingHour int
		expectedDiaryEmbeddingMin  int
		expectError                bool
	}{
		{
			name:                       "正常系：デフォルト値",
			monthlyInterval:            "",
			diaryEmbeddingHour:         "",
			diaryEmbeddingMinute:       "",
			expectedMonthly:            5 * time.Minute,
			expectedDiaryEmbeddingHour: 4,
			expectedDiaryEmbeddingMin:  30,
			expectError:                false,
		},
		{
			name:                       "正常系：カスタム値",
			monthlyInterval:            "1h",
			diaryEmbeddingHour:         "5",
			diaryEmbeddingMinute:       "15",
			expectedMonthly:            1 * time.Hour,
			expectedDiaryEmbeddingHour: 5,
			expectedDiaryEmbeddingMin:  15,
			expectError:                false,
		},
		{
			name:            "異常系：無効な月次インターバル",
			monthlyInterval: "invalid",
			expectError:     true,
		},
		{
			name:                 "異常系：無効なdiaryEmbeddingHour",
			monthlyInterval:      "5m",
			diaryEmbeddingHour:   "25",
			diaryEmbeddingMinute: "0",
			expectError:          true,
		},
		{
			name:                 "異常系：無効なdiaryEmbeddingMinute",
			monthlyInterval:      "5m",
			diaryEmbeddingHour:   "4",
			diaryEmbeddingMinute: "60",
			expectError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をセットアップ
			_ = os.Unsetenv("SCHEDULER_DAILY_INTERVAL")

			if tt.monthlyInterval != "" {
				_ = os.Setenv("SCHEDULER_MONTHLY_INTERVAL", tt.monthlyInterval)
			} else {
				_ = os.Unsetenv("SCHEDULER_MONTHLY_INTERVAL")
			}

			if tt.diaryEmbeddingHour != "" {
				_ = os.Setenv("SCHEDULER_DIARY_EMBEDDING_HOUR", tt.diaryEmbeddingHour)
			} else {
				_ = os.Unsetenv("SCHEDULER_DIARY_EMBEDDING_HOUR")
			}

			if tt.diaryEmbeddingMinute != "" {
				_ = os.Setenv("SCHEDULER_DIARY_EMBEDDING_MINUTE", tt.diaryEmbeddingMinute)
			} else {
				_ = os.Unsetenv("SCHEDULER_DIARY_EMBEDDING_MINUTE")
			}

			// 関数をテスト
			config, err := LoadSchedulerConfig()

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.MonthlySummaryInterval != tt.expectedMonthly {
				t.Errorf("expected monthly interval %v, got %v", tt.expectedMonthly, config.MonthlySummaryInterval)
			}

			if config.DiaryEmbeddingTargetHour != tt.expectedDiaryEmbeddingHour {
				t.Errorf("expected DiaryEmbeddingTargetHour %d, got %d", tt.expectedDiaryEmbeddingHour, config.DiaryEmbeddingTargetHour)
			}

			if config.DiaryEmbeddingTargetMinute != tt.expectedDiaryEmbeddingMin {
				t.Errorf("expected DiaryEmbeddingTargetMinute %d, got %d", tt.expectedDiaryEmbeddingMin, config.DiaryEmbeddingTargetMinute)
			}
		})
	}
}

func TestLoadSubscriberConfig(t *testing.T) {
	tests := []struct {
		name              string
		maxConcurrentJobs string
		expectedMaxJobs   int
		expectError       bool
	}{
		{
			name:              "正常系：デフォルト値",
			maxConcurrentJobs: "",
			expectedMaxJobs:   10,
			expectError:       false,
		},
		{
			name:              "正常系：カスタム値",
			maxConcurrentJobs: "5",
			expectedMaxJobs:   5,
			expectError:       false,
		},
		{
			name:              "異常系：無効な値（非数値）",
			maxConcurrentJobs: "invalid",
			expectedMaxJobs:   0,
			expectError:       true,
		},
		{
			name:              "異常系：無効な値（ゼロ）",
			maxConcurrentJobs: "0",
			expectedMaxJobs:   0,
			expectError:       true,
		},
		{
			name:              "異常系：無効な値（負の値）",
			maxConcurrentJobs: "-1",
			expectedMaxJobs:   0,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable
			if tt.maxConcurrentJobs != "" {
				_ = os.Setenv("SUBSCRIBER_MAX_CONCURRENT_JOBS", tt.maxConcurrentJobs)
			} else {
				_ = os.Unsetenv("SUBSCRIBER_MAX_CONCURRENT_JOBS")
			}

			// Test the function
			config, err := LoadSubscriberConfig()

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.MaxConcurrentJobs != tt.expectedMaxJobs {
				t.Errorf("expected max concurrent jobs %d, got %d", tt.expectedMaxJobs, config.MaxConcurrentJobs)
			}
		})
	}
}
