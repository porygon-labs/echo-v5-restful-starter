package config_test

import (
	"testing"

	"go-restful-api/internal/config"
)

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_HOST", "127.0.0.1")
	t.Setenv("APP_PORT", "9090")

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
	t.Setenv("APP_PORT", "not-a-port")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load() succeeded with an invalid APP_PORT")
	}
}
