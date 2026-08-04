// Package config loads application configuration from an optional .env file and process environment variables.
package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config contains the application's runtime configuration.
type Config struct {
	App AppConfig
}

// AppConfig contains settings for the HTTP application.
type AppConfig struct {
	Environment string `env:"APP_ENV" envDefault:"development"`
	Host        string `env:"APP_HOST" envDefault:"0.0.0.0"`
	Port        uint16 `env:"PORT" envDefault:"8080"`
}

// Load parses configuration from .env and the process environment.
// Process environment variables take precedence over values in .env.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse environment variables: %w", err)
	}

	return cfg, nil
}
