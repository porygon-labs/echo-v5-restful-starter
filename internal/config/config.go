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
	App    AppConfig
	DB     DBConfig
	Redis  RedisConfig
	Otel   OtelConfig
	Logger LoggerConfig
}

// AppConfig contains settings for the HTTP application.
type AppConfig struct {
	Environment string `env:"APP_ENV" envDefault:"development"`
	Host        string `env:"APP_HOST" envDefault:"0.0.0.0"`
	Port        uint16 `env:"PORT" envDefault:"8080"`
}

// DBConfig contains database connection settings.
type DBConfig struct {
	DSN string `env:"DB_DSN,required,notEmpty"`
}

// RedisConfig contains Redis connection settings.
type RedisConfig struct {
	URL string `env:"REDIS_URL,required,notEmpty"`
}

// OtelConfig contains OpenTelemetry settings.
type OtelConfig struct {
	ExporterEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"localhost:4317"`
	ServiceName      string `env:"OTEL_SERVICE_NAME" envDefault:"go-restful-api"`
}

// LoggerConfig contains structured logging settings.
type LoggerConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
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
