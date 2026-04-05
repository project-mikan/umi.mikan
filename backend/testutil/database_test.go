package testutil

import (
	"os"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("環境変数が設定されている場合はその値を返す", func(t *testing.T) {
		if err := os.Setenv("TEST_GETENV_VAR_UNIQUE_12345", "test-value"); err != nil {
			t.Fatalf("Setenv失敗: %v", err)
		}
		defer func() { _ = os.Unsetenv("TEST_GETENV_VAR_UNIQUE_12345") }()

		result := getEnvOrDefault("TEST_GETENV_VAR_UNIQUE_12345", "default")
		if result != "test-value" {
			t.Errorf("got %q, want %q", result, "test-value")
		}
	})

	t.Run("環境変数が設定されていない場合はデフォルト値を返す", func(t *testing.T) {
		result := getEnvOrDefault("NON_EXISTENT_ENV_VAR_UNIQUE_12345", "default-value")
		if result != "default-value" {
			t.Errorf("got %q, want %q", result, "default-value")
		}
	})
}

func TestDefaultTestDBConfig(t *testing.T) {
	config := DefaultTestDBConfig()

	if config.Host == "" {
		t.Error("Hostはデフォルトでlocalhostであるべき")
	}
	if config.Port == 0 {
		t.Error("Portは0でないはず")
	}
	if config.User == "" {
		t.Error("Userは空であってはならない")
	}
	if config.DBName == "" {
		t.Error("DBNameは空であってはならない")
	}
}

func TestDefaultTestDBConfig_WithEnv(t *testing.T) {
	envVars := map[string]string{
		"TEST_DB_HOST":     "custom-host",
		"TEST_DB_USER":     "custom-user",
		"TEST_DB_PASSWORD": "custom-pass",
		"TEST_DB_NAME":     "custom-db",
	}
	for k, v := range envVars {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("Setenv(%q)失敗: %v", k, err)
		}
	}
	defer func() {
		for k := range envVars {
			_ = os.Unsetenv(k)
		}
	}()

	config := DefaultTestDBConfig()

	if config.Host != "custom-host" {
		t.Errorf("Host: got %q, want %q", config.Host, "custom-host")
	}
	if config.User != "custom-user" {
		t.Errorf("User: got %q, want %q", config.User, "custom-user")
	}
	if config.DBName != "custom-db" {
		t.Errorf("DBName: got %q, want %q", config.DBName, "custom-db")
	}
}
