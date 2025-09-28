package constants

import (
	"os"
	"testing"
	"time"
)

func TestLoadSchedulerConfig(t *testing.T) {
	tests := []struct {
		name            string
		dailyInterval   string
		monthlyInterval string
		expectedDaily   time.Duration
		expectedMonthly time.Duration
		expectError     bool
	}{
		{
			name:            "正常系：デフォルト値",
			dailyInterval:   "",
			monthlyInterval: "",
			expectedDaily:   5 * time.Minute,
			expectedMonthly: 5 * time.Minute,
			expectError:     false,
		},
		{
			name:            "正常系：カスタム値",
			dailyInterval:   "10m",
			monthlyInterval: "1h",
			expectedDaily:   10 * time.Minute,
			expectedMonthly: 1 * time.Hour,
			expectError:     false,
		},
		{
			name:            "異常系：無効な日次インターバル",
			dailyInterval:   "invalid",
			monthlyInterval: "5m",
			expectedDaily:   0,
			expectedMonthly: 0,
			expectError:     true,
		},
		{
			name:            "異常系：無効な月次インターバル",
			dailyInterval:   "5m",
			monthlyInterval: "invalid",
			expectedDaily:   0,
			expectedMonthly: 0,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			if tt.dailyInterval != "" {
				_ = os.Setenv("SCHEDULER_DAILY_INTERVAL", tt.dailyInterval)
			} else {
				_ = os.Unsetenv("SCHEDULER_DAILY_INTERVAL")
			}

			if tt.monthlyInterval != "" {
				_ = os.Setenv("SCHEDULER_MONTHLY_INTERVAL", tt.monthlyInterval)
			} else {
				_ = os.Unsetenv("SCHEDULER_MONTHLY_INTERVAL")
			}

			// Test the function
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

			if config.DailySummaryInterval != tt.expectedDaily {
				t.Errorf("expected daily interval %v, got %v", tt.expectedDaily, config.DailySummaryInterval)
			}

			if config.MonthlySummaryInterval != tt.expectedMonthly {
				t.Errorf("expected monthly interval %v, got %v", tt.expectedMonthly, config.MonthlySummaryInterval)
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
