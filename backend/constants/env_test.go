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

func TestLoadGRPCReflectionEnabled(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{
			name:     "正常系：未設定（デフォルト有効）",
			env:      "",
			expected: true,
		},
		{
			name:     "正常系：development",
			env:      "development",
			expected: true,
		},
		{
			name:     "正常系：prod（無効）",
			env:      "prod",
			expected: false,
		},
		{
			name:     "正常系：production（無効）",
			env:      "production",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable
			if tt.env != "" {
				_ = os.Setenv("BACKEND_ENV", tt.env)
			} else {
				_ = os.Unsetenv("BACKEND_ENV")
			}

			// Test the function
			result := LoadGRPCReflectionEnabled()

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoadRegisterKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "正常系：キーが設定されている場合",
			key:      "my-secret-key",
			expected: "my-secret-key",
		},
		{
			name:     "正常系：キーが設定されていない場合",
			key:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variable
			if tt.key != "" {
				_ = os.Setenv("REGISTER_KEY", tt.key)
			} else {
				_ = os.Unsetenv("REGISTER_KEY")
			}

			// Test the function
			result := LoadRegisterKey()

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoadRateLimitConfig(t *testing.T) {
	tests := []struct {
		name                    string
		loginMaxAttempts        string
		loginWindow             string
		registerMaxAttempts     string
		registerWindow          string
		expectedLoginAttempts   int
		expectedLoginWindow     time.Duration
		expectedRegisterAttempts int
		expectedRegisterWindow   time.Duration
		expectError             bool
	}{
		{
			name:                     "正常系：デフォルト値",
			loginMaxAttempts:         "",
			loginWindow:              "",
			registerMaxAttempts:      "",
			registerWindow:           "",
			expectedLoginAttempts:    5,
			expectedLoginWindow:      15 * time.Minute,
			expectedRegisterAttempts: 3,
			expectedRegisterWindow:   1 * time.Hour,
			expectError:              false,
		},
		{
			name:                     "正常系：カスタム値",
			loginMaxAttempts:         "10",
			loginWindow:              "30m",
			registerMaxAttempts:      "5",
			registerWindow:           "2h",
			expectedLoginAttempts:    10,
			expectedLoginWindow:      30 * time.Minute,
			expectedRegisterAttempts: 5,
			expectedRegisterWindow:   2 * time.Hour,
			expectError:              false,
		},
		{
			name:                 "異常系：無効なログイン試行回数",
			loginMaxAttempts:     "invalid",
			loginWindow:          "15m",
			registerMaxAttempts:  "3",
			registerWindow:       "1h",
			expectError:          true,
		},
		{
			name:                 "異常系：ゼロのログイン試行回数",
			loginMaxAttempts:     "0",
			loginWindow:          "15m",
			registerMaxAttempts:  "3",
			registerWindow:       "1h",
			expectError:          true,
		},
		{
			name:                 "異常系：無効なログインウィンドウ",
			loginMaxAttempts:     "5",
			loginWindow:          "invalid",
			registerMaxAttempts:  "3",
			registerWindow:       "1h",
			expectError:          true,
		},
		{
			name:                 "異常系：ゼロのログインウィンドウ",
			loginMaxAttempts:     "5",
			loginWindow:          "0s",
			registerMaxAttempts:  "3",
			registerWindow:       "1h",
			expectError:          true,
		},
		{
			name:                 "異常系：無効な登録試行回数",
			loginMaxAttempts:     "5",
			loginWindow:          "15m",
			registerMaxAttempts:  "invalid",
			registerWindow:       "1h",
			expectError:          true,
		},
		{
			name:                 "異常系：ゼロの登録試行回数",
			loginMaxAttempts:     "5",
			loginWindow:          "15m",
			registerMaxAttempts:  "0",
			registerWindow:       "1h",
			expectError:          true,
		},
		{
			name:                 "異常系：無効な登録ウィンドウ",
			loginMaxAttempts:     "5",
			loginWindow:          "15m",
			registerMaxAttempts:  "3",
			registerWindow:       "invalid",
			expectError:          true,
		},
		{
			name:                 "異常系：ゼロの登録ウィンドウ",
			loginMaxAttempts:     "5",
			loginWindow:          "15m",
			registerMaxAttempts:  "3",
			registerWindow:       "0s",
			expectError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			if tt.loginMaxAttempts != "" {
				_ = os.Setenv("LOGIN_MAX_ATTEMPTS", tt.loginMaxAttempts)
			} else {
				_ = os.Unsetenv("LOGIN_MAX_ATTEMPTS")
			}

			if tt.loginWindow != "" {
				_ = os.Setenv("LOGIN_WINDOW", tt.loginWindow)
			} else {
				_ = os.Unsetenv("LOGIN_WINDOW")
			}

			if tt.registerMaxAttempts != "" {
				_ = os.Setenv("REGISTER_MAX_ATTEMPTS", tt.registerMaxAttempts)
			} else {
				_ = os.Unsetenv("REGISTER_MAX_ATTEMPTS")
			}

			if tt.registerWindow != "" {
				_ = os.Setenv("REGISTER_WINDOW", tt.registerWindow)
			} else {
				_ = os.Unsetenv("REGISTER_WINDOW")
			}

			// Test the function
			config, err := LoadRateLimitConfig()

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.LoginMaxAttempts != tt.expectedLoginAttempts {
				t.Errorf("expected login max attempts %d, got %d", tt.expectedLoginAttempts, config.LoginMaxAttempts)
			}

			if config.LoginWindow != tt.expectedLoginWindow {
				t.Errorf("expected login window %v, got %v", tt.expectedLoginWindow, config.LoginWindow)
			}

			if config.RegisterMaxAttempts != tt.expectedRegisterAttempts {
				t.Errorf("expected register max attempts %d, got %d", tt.expectedRegisterAttempts, config.RegisterMaxAttempts)
			}

			if config.RegisterWindow != tt.expectedRegisterWindow {
				t.Errorf("expected register window %v, got %v", tt.expectedRegisterWindow, config.RegisterWindow)
			}
		})
	}
}
