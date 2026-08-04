// Package provider wires the application's shared dependencies.
package provider

import (
	"errors"
	"fmt"

	"go-restful-api/internal/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Deps contains the shared dependencies used by the application modules.
type Deps struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// NewDeps connects to the application's database and creates its Redis client.
func NewDeps(cfg config.Config) (*Deps, error) {
	if cfg.DB.DSN == "" {
		return nil, errors.New("database: DB_DSN is required")
	}
	if cfg.Redis.URL == "" {
		return nil, errors.New("redis: REDIS_URL is required")
	}

	redisOptions, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("redis: parse URL: %w", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database: connect: %w", err)
	}

	return &Deps{
		DB:    db,
		Redis: redis.NewClient(redisOptions),
	}, nil
}

// Close releases all shared dependencies.
func (d *Deps) Close() error {
	if d == nil {
		return nil
	}

	var errs []error
	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close redis: %w", err))
		}
	}
	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("get database connection: %w", err))
		} else if err := sqlDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close database: %w", err))
		}
	}

	return errors.Join(errs...)
}
