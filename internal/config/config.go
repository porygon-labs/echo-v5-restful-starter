// Package config loads application configuration from environment variables.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config contains the application's runtime configuration.
type Config struct {
	App AppConfig `envPrefix:"APP_"`
}

// AppConfig contains settings for the HTTP application.
type AppConfig struct {
	Environment string `env:"ENV" envDefault:"development"`
	Host        string `env:"HOST" envDefault:"0.0.0.0"`
	Port        uint16 `env:"PORT" envDefault:"8080"`
}

// Load parses configuration from the process environment.
func Load() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse environment variables: %w", err)
	}

	return cfg, nil
}
