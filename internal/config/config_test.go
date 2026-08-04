package config_test

import (
	"os"
	"testing"

	"go-restful-api/internal/config"
)

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_HOST", "127.0.0.1")
	t.Setenv("PORT", "9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.App.Environment != "test" {
		t.Errorf("Environment = %q, want %q", cfg.App.Environment, "test")
	}
	if cfg.App.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want %q", cfg.App.Host, "127.0.0.1")
	}
	if cfg.App.Port != 9090 {
		t.Errorf("Port = %d, want %d", cfg.App.Port, 9090)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-port")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load() succeeded with an invalid PORT")
	}
}

func TestLoadReadsDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte("PORT=8082\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	unsetEnv(t, "PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.App.Port != 8082 {
		t.Errorf("Port = %d, want %d from .env", cfg.App.Port, 8082)
	}
}

func TestLoadPrefersProcessEnvironment(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".env", []byte("PORT=8082\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Setenv("PORT", "9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.App.Port != 9090 {
		t.Errorf("Port = %d, want %d from process environment", cfg.App.Port, 9090)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	value, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if exists {
			if err := os.Setenv(key, value); err != nil {
				t.Errorf("restore %s: %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Errorf("unset %s during cleanup: %v", key, err)
		}
	})
}
