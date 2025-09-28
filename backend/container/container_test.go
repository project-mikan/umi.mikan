package container

import (
	"context"
	"testing"
	"time"
)

func TestNewContainer(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if container == nil {
		t.Fatal("expected container to be non-nil")
	}

	if container.container == nil {
		t.Fatal("expected inner container to be non-nil")
	}
}

func TestContainerProviders(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 各アプリケーションタイプが解決できることをテスト
	var serverApp *ServerApp
	err = container.Invoke(func(app *ServerApp) {
		serverApp = app
	})
	if err != nil {
		t.Errorf("failed to resolve ServerApp: %v", err)
	}
	if serverApp == nil {
		t.Error("expected ServerApp to be non-nil")
	}

	var schedulerApp *SchedulerApp
	err = container.Invoke(func(app *SchedulerApp) {
		schedulerApp = app
	})
	if err != nil {
		t.Errorf("failed to resolve SchedulerApp: %v", err)
	}
	if schedulerApp == nil {
		t.Error("expected SchedulerApp to be non-nil")
	}

	var subscriberApp *SubscriberApp
	err = container.Invoke(func(app *SubscriberApp) {
		subscriberApp = app
	})
	if err != nil {
		t.Errorf("failed to resolve SubscriberApp: %v", err)
	}
	if subscriberApp == nil {
		t.Error("expected SubscriberApp to be non-nil")
	}

	var cleanup *Cleanup
	err = container.Invoke(func(c *Cleanup) {
		cleanup = c
	})
	if err != nil {
		t.Errorf("failed to resolve Cleanup: %v", err)
	}
	if cleanup == nil {
		t.Error("expected Cleanup to be non-nil")
	}
}

func TestLLMClientFactory(t *testing.T) {
	factory := &geminiClientFactory{}
	// ファクトリがインターフェースを実装していることをテスト
	var _ LLMClientFactory = factory
}

func TestLockService(t *testing.T) {
	// テストでは実際のRedis接続がないため、最小限のテスト
	lockSvc := &lockService{}
	// ロックサービスがインターフェースを実装していることをテスト
	var _ LockService = lockSvc
}

// TestContainerDependencyResolution tests that all dependencies are correctly resolved
func TestContainerDependencyResolution(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error creating container, got %v", err)
	}

	// Test individual components can be resolved
	testCases := []struct {
		name string
		fn   any
	}{
		{"DBConfig", func(config *DBConfig) {
			if config == nil {
				t.Error("DBConfig should not be nil")
			}
		}},
		{"RedisConfig", func(config *RedisConfig) {
			if config == nil {
				t.Error("RedisConfig should not be nil")
			}
		}},
		{"SchedulerConfig", func(config *SchedulerConfig) {
			if config == nil {
				t.Error("SchedulerConfig should not be nil")
				return
			}
			if config.DailySummaryInterval <= 0 {
				t.Error("DailySummaryInterval should be positive")
			}
			if config.MonthlySummaryInterval <= 0 {
				t.Error("MonthlySummaryInterval should be positive")
			}
		}},
		{"SubscriberConfig", func(config *SubscriberConfig) {
			if config == nil {
				t.Error("SubscriberConfig should not be nil")
				return
			}
			if config.MaxConcurrentJobs <= 0 {
				t.Error("MaxConcurrentJobs should be positive")
			}
		}},
		{"LLMClientFactory", func(factory LLMClientFactory) {
			if factory == nil {
				t.Error("LLMClientFactory should not be nil")
			}
		}},
		{"LockService", func(lockSvc LockService) {
			if lockSvc == nil {
				t.Error("LockService should not be nil")
			}
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := container.Invoke(tc.fn)
			if err != nil {
				t.Errorf("failed to resolve %s: %v", tc.name, err)
			}
		})
	}
}

// TestApplicationBundles tests that application bundles contain all necessary dependencies
func TestApplicationBundles(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error creating container, got %v", err)
	}

	// Test ServerApp dependencies
	t.Run("ServerApp", func(t *testing.T) {
		err := container.Invoke(func(app *ServerApp) {
			if app == nil {
				t.Fatal("ServerApp should not be nil")
			}
			if app.DB == nil {
				t.Error("ServerApp.DB should not be nil")
			}
			if app.Redis == nil {
				t.Error("ServerApp.Redis should not be nil")
			}
			if app.AuthService == nil {
				t.Error("ServerApp.AuthService should not be nil")
			}
			if app.DiaryService == nil {
				t.Error("ServerApp.DiaryService should not be nil")
			}
			if app.UserService == nil {
				t.Error("ServerApp.UserService should not be nil")
			}
		})
		if err != nil {
			t.Errorf("failed to resolve ServerApp: %v", err)
		}
	})

	// Test SchedulerApp dependencies
	t.Run("SchedulerApp", func(t *testing.T) {
		err := container.Invoke(func(app *SchedulerApp) {
			if app == nil {
				t.Fatal("SchedulerApp should not be nil")
			}
			if app.DB == nil {
				t.Error("SchedulerApp.DB should not be nil")
			}
			if app.Redis == nil {
				t.Error("SchedulerApp.Redis should not be nil")
			}
			if app.SchedulerConfig == nil {
				t.Error("SchedulerApp.SchedulerConfig should not be nil")
			}
		})
		if err != nil {
			t.Errorf("failed to resolve SchedulerApp: %v", err)
		}
	})

	// Test SubscriberApp dependencies
	t.Run("SubscriberApp", func(t *testing.T) {
		err := container.Invoke(func(app *SubscriberApp) {
			if app == nil {
				t.Fatal("SubscriberApp should not be nil")
			}
			if app.DB == nil {
				t.Error("SubscriberApp.DB should not be nil")
			}
			if app.Redis == nil {
				t.Error("SubscriberApp.Redis should not be nil")
			}
			if app.LLMFactory == nil {
				t.Error("SubscriberApp.LLMFactory should not be nil")
			}
			if app.LockService == nil {
				t.Error("SubscriberApp.LockService should not be nil")
			}
			if app.SubscriberConfig == nil {
				t.Error("SubscriberApp.SubscriberConfig should not be nil")
			}
		})
		if err != nil {
			t.Errorf("failed to resolve SubscriberApp: %v", err)
		}
	})
}

// TestCleanupFunction tests the cleanup functionality
func TestCleanupFunction(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error creating container, got %v", err)
	}

	err = container.Invoke(func(cleanup *Cleanup) {
		if cleanup == nil {
			t.Fatal("Cleanup should not be nil")
		}
		if cleanup.db == nil {
			t.Error("Cleanup.db should not be nil")
		}
		if cleanup.redis == nil {
			t.Error("Cleanup.redis should not be nil")
		}

		// Test that cleanup doesn't panic (actual cleanup would close connections)
		// In a real test environment, we'd want to test this properly
		// For now, just ensure the method exists and can be called
		err := cleanup.Close()
		// We expect an error here because we're in a test environment
		// without real connections, but the method should exist
		_ = err // Ignore the error as it's expected in test environment
	})
	if err != nil {
		t.Errorf("failed to resolve Cleanup: %v", err)
	}
}

// TestLLMClientFactoryFunctionality tests the LLM client factory
func TestLLMClientFactoryFunctionality(t *testing.T) {
	factory := &geminiClientFactory{}

	// Test with empty API key (should return error)
	ctx := context.Background()
	client, err := factory.CreateGeminiClient(ctx, "")

	// We expect this to fail with empty key
	if err == nil {
		t.Error("Expected error with empty API key, but got success")
	}
	if client != nil {
		t.Error("Expected nil client with empty API key")
	}

	// Test with non-empty API key (may or may not fail, but should not panic)
	// The actual validation might happen during API calls, not during client creation
	client2, err2 := factory.CreateGeminiClient(ctx, "test-key")
	// We don't assert on the result here since the behavior may vary
	// Just ensure it doesn't panic and returns something reasonable
	_ = client2
	_ = err2
}

// TestLockServiceFunctionality tests the lock service
func TestLockServiceFunctionality(t *testing.T) {
	// Note: This test would require a real Redis connection to be meaningful
	// For now, we just test that the methods exist and can be called
	lockSvc := &lockService{}

	// Test that we can create a distributed lock (won't work without Redis)
	lock := lockSvc.NewDistributedLock("test-key", 5*time.Minute)
	if lock == nil {
		t.Error("NewDistributedLock should return a non-nil lock")
	}
}

// TestConfigurationValues tests that configuration values are reasonable
func TestConfigurationValues(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error creating container, got %v", err)
	}

	err = container.Invoke(func(
		dbConfig *DBConfig,
		redisConfig *RedisConfig,
		schedulerConfig *SchedulerConfig,
		subscriberConfig *SubscriberConfig,
	) {
		// Test DB config
		if dbConfig.Host == "" {
			t.Error("DB host should not be empty")
		}
		if dbConfig.Port <= 0 || dbConfig.Port > 65535 {
			t.Errorf("DB port should be valid, got %d", dbConfig.Port)
		}
		if dbConfig.User == "" {
			t.Error("DB user should not be empty")
		}
		if dbConfig.DBName == "" {
			t.Error("DB name should not be empty")
		}

		// Test Redis config
		if redisConfig.Host == "" {
			t.Error("Redis host should not be empty")
		}
		if redisConfig.Port <= 0 || redisConfig.Port > 65535 {
			t.Errorf("Redis port should be valid, got %d", redisConfig.Port)
		}

		// Test Scheduler config
		if schedulerConfig.DailySummaryInterval <= 0 {
			t.Error("Daily summary interval should be positive")
		}
		if schedulerConfig.MonthlySummaryInterval <= 0 {
			t.Error("Monthly summary interval should be positive")
		}

		// Test Subscriber config
		if subscriberConfig.MaxConcurrentJobs <= 0 {
			t.Error("Max concurrent jobs should be positive")
		}
		if subscriberConfig.MaxConcurrentJobs > 1000 {
			t.Error("Max concurrent jobs seems unreasonably high")
		}
	})
	if err != nil {
		t.Errorf("failed to resolve configs: %v", err)
	}
}
