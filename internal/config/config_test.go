package config_test

import (
	"os"
	"strings"
	"testing"

	"go-restful-api/internal/config"
)

func TestLoad(t *testing.T) {
	setRequiredEnv(t)
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

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnv(t)
	unsetEnv(t, "APP_ENV", "APP_HOST", "PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.App.Environment != "development" {
		t.Errorf("default Environment = %q, want %q", cfg.App.Environment, "development")
	}
	if cfg.App.Host != "0.0.0.0" {
		t.Errorf("default Host = %q, want %q", cfg.App.Host, "0.0.0.0")
	}
	if cfg.App.Port != 8080 {
		t.Errorf("default Port = %d, want %d", cfg.App.Port, 8080)
	}
}

func TestLoad_DBConfig(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("DB_DSN", "host=db user=app")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.DB.DSN != "host=db user=app" {
		t.Errorf("DB.DSN = %q, want %q", cfg.DB.DSN, "host=db user=app")
	}
}

func TestLoad_RedisConfig(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("REDIS_URL", "redis://cache:6380")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Redis.URL != "redis://cache:6380" {
		t.Errorf("Redis.URL = %q, want %q", cfg.Redis.URL, "redis://cache:6380")
	}
}

func TestLoadRejectsMissingDatabaseDSN(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnv(t)
	unsetEnv(t, "DB_DSN")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "DB_DSN") {
		t.Fatalf("Load() error = %v, want missing DB_DSN error", err)
	}
}

func TestLoadRejectsMissingRedisURL(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnv(t)
	unsetEnv(t, "REDIS_URL")

	_, err := config.Load()
	if err == nil || !strings.Contains(err.Error(), "REDIS_URL") {
		t.Fatalf("Load() error = %v, want missing REDIS_URL error", err)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PORT", "not-a-port")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load() succeeded with an invalid PORT")
	}
}

func TestLoadReadsDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	setRequiredEnv(t)
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
	setRequiredEnv(t)
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

func TestLoad_UnreadableDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir(".env", 0o700); err != nil {
		t.Fatalf("create .env dir: %v", err)
	}

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when .env is a directory")
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_DSN", "postgres://app:secret@localhost:5432/app?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()

	saved := map[string]*string{}
	for _, key := range keys {
		value, exists := os.LookupEnv(key)
		if exists {
			v := value
			saved[key] = &v
		} else {
			saved[key] = nil
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for key, val := range saved {
			if val != nil {
				os.Setenv(key, *val)
			} else {
				os.Unsetenv(key)
			}
		}
	})
}
