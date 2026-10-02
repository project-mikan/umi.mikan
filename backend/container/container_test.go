package container

import (
	"context"
	"testing"
	"time"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/llm"
)

func TestNewContainer(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if container == nil {
		t.Fatal("expected container to be non-nil")
		return
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

// TestContainerDependencyResolution すべての依存関係が正しく解決されることをテスト
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

// TestApplicationBundles アプリケーションバンドルが必要なすべての依存関係を含んでいることをテスト
func TestApplicationBundles(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error creating container, got %v", err)
	}

	// ServerAppの依存関係をテスト
	t.Run("ServerApp", func(t *testing.T) {
		err := container.Invoke(func(app *ServerApp) {
			if app == nil {
				t.Fatal("ServerApp should not be nil")
				return
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

	// SchedulerAppの依存関係をテスト
	t.Run("SchedulerApp", func(t *testing.T) {
		err := container.Invoke(func(app *SchedulerApp) {
			if app == nil {
				t.Fatal("SchedulerApp should not be nil")
				return
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

	// SubscriberAppの依存関係をテスト
	t.Run("SubscriberApp", func(t *testing.T) {
		err := container.Invoke(func(app *SubscriberApp) {
			if app == nil {
				t.Fatal("SubscriberApp should not be nil")
				return
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

// TestCleanupFunction クリーンアップ機能をテスト
func TestCleanupFunction(t *testing.T) {
	container, err := NewContainer()
	if err != nil {
		t.Fatalf("expected no error creating container, got %v", err)
	}

	err = container.Invoke(func(cleanup *Cleanup) {
		if cleanup == nil {
			t.Fatal("Cleanup should not be nil")
			return
		}
		if cleanup.db == nil {
			t.Error("Cleanup.db should not be nil")
		}
		if cleanup.redis == nil {
			t.Error("Cleanup.redis should not be nil")
		}

		// クリーンアップがパニックしないことをテスト（実際のクリーンアップでは接続を閉じる）
		// 実際のテスト環境では、これを適切にテストしたい
		// 今のところは、メソッドが存在して呼び出し可能であることを確認
		err := cleanup.Close()
		// テスト環境では実際の接続がないため、ここでエラーが期待される
		// しかしメソッドは存在するべき
		_ = err // テスト環境では期待されるエラーなので無視
	})
	if err != nil {
		t.Errorf("failed to resolve Cleanup: %v", err)
	}
}

// TestLLMClientFactoryFunctionality LLMクライアントファクトリをテスト
func TestLLMClientFactoryFunctionality(t *testing.T) {
	tests := []struct {
		name   string
		config constants.VertexAIConfig
	}{
		{
			name:   "異常系: GOOGLE_CLOUD_PROJECTが未設定だとVertex AIの接続先が決まらないのでエラーになる",
			config: constants.VertexAIConfig{Project: "", Location: "global"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &geminiClientFactory{config: tt.config}
			ctx := context.Background()

			client, err := factory.CreateGeminiClient(ctx)
			if err == nil {
				t.Fatal("エラーが返ることを期待したが nil だった")
			}
			if client != nil {
				t.Error("エラー時はクライアントが nil であることを期待")
			}

			// diary.LLMFactory 用アダプタも同じ設定を使うためエラーになる
			adapter := &diaryLLMFactory{factory: factory}
			embedder, err := adapter.CreateGeminiClient(ctx)
			if err == nil {
				t.Fatal("アダプタでもエラーが返ることを期待したが nil だった")
			}
			if embedder != nil {
				t.Error("エラー時は embedder が nil であることを期待")
			}
		})
	}
}

// TestLLMClientFactoryReusesClient 生成済みのクライアントを使い回すことをテスト
func TestLLMClientFactoryReusesClient(t *testing.T) {
	cached := &llm.GeminiClient{}
	// Projectが空でも、生成済みクライアントがあれば再生成せずにそれを返す
	factory := &geminiClientFactory{client: cached}

	got, err := factory.CreateGeminiClient(context.Background())
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if got != cached {
		t.Error("生成済みのクライアントが返ることを期待")
	}
}

// TestLockServiceFunctionality ロックサービスをテスト
func TestLockServiceFunctionality(t *testing.T) {
	// 注意: このテストが意味を成すには実際のRedis接続が必要
	// 今のところは、メソッドが存在して呼び出し可能であることだけをテスト
	lockSvc := &lockService{}

	// 分散ロックを作成できることをテスト（Redisなしでは動作しない）
	lock := lockSvc.NewDistributedLock("test-key", 5*time.Minute)
	if lock == nil {
		t.Error("NewDistributedLock should return a non-nil lock")
	}
}

// TestConfigurationValues 設定値が合理的であることをテスト
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
		// DB設定をテスト
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

		// Redis設定をテスト
		if redisConfig.Host == "" {
			t.Error("Redis host should not be empty")
		}
		if redisConfig.Port <= 0 || redisConfig.Port > 65535 {
			t.Errorf("Redis port should be valid, got %d", redisConfig.Port)
		}

		// Scheduler設定をテスト
		if schedulerConfig.MonthlySummaryInterval <= 0 {
			t.Error("Monthly summary interval should be positive")
		}

		// Subscriber設定をテスト
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
