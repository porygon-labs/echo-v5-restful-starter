package provider_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-restful-api/internal/config"
	"go-restful-api/internal/pkg/response"
	"go-restful-api/internal/provider"

	"github.com/alicebob/miniredis/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// ─── helpers ─────────────────────────────────────────────────────────────────

func newEcho() *echo.Echo { return echo.New() }

func decodeMeta(t *testing.T, rec *httptest.ResponseRecorder) response.Meta {
	t.Helper()
	var env response.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return env.Meta
}

func newSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	return db
}

func newMiniRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func serve(t *testing.T, e *echo.Echo, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ─── NewDeps ─────────────────────────────────────────────────────────────────

func TestNewDepsValidatesConfigBeforeConnecting(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		message string
	}{
		{
			name: "missing database DSN",
			cfg: config.Config{
				Redis: config.RedisConfig{URL: "redis://localhost:6379"},
			},
			message: "DB_DSN",
		},
		{
			name: "missing Redis URL",
			cfg: config.Config{
				DB: config.DBConfig{DSN: "://"},
			},
			message: "REDIS_URL",
		},
		{
			name: "invalid Redis URL",
			cfg: config.Config{
				DB:    config.DBConfig{DSN: "://"},
				Redis: config.RedisConfig{URL: "://"},
			},
			message: "redis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.NewDeps(tt.cfg)
			if err == nil {
				t.Fatal("NewDeps() succeeded")
			}
			if !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("NewDeps() error = %q, want it to contain %q", err, tt.message)
			}
		})
	}
}

func TestNewDepsReportsDatabaseConnectionError(t *testing.T) {
	cfg := config.Config{
		DB:    config.DBConfig{DSN: "://"},
		Redis: config.RedisConfig{URL: "redis://localhost:6379"},
	}

	_, err := provider.NewDeps(cfg)
	if err == nil {
		t.Fatal("NewDeps() succeeded")
	}
	if !strings.Contains(err.Error(), "database") {
		t.Fatalf("NewDeps() error = %q, want a database error", err)
	}
}

// ─── healthz ─────────────────────────────────────────────────────────────────

func TestHealthz(t *testing.T) {
	e := newEcho()
	deps := &provider.Deps{DB: newSQLiteDB(t), Redis: newMiniRedis(t)}
	provider.RegisterRoutes(e, deps)

	rec := serve(t, e, http.MethodGet, "/healthz")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	meta := decodeMeta(t, rec)
	if !meta.IsSuccess {
		t.Error("IsSuccess = false")
	}
}

// ─── readyz ──────────────────────────────────────────────────────────────────

func TestReadyz_AllHealthy(t *testing.T) {
	e := newEcho()
	deps := &provider.Deps{DB: newSQLiteDB(t), Redis: newMiniRedis(t)}
	provider.RegisterRoutes(e, deps)

	rec := serve(t, e, http.MethodGet, "/readyz")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestReadyz_DBUnavailable(t *testing.T) {
	e := newEcho()
	provider.RegisterRoutes(e, &provider.Deps{})

	rec := serve(t, e, http.MethodGet, "/readyz")

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	meta := decodeMeta(t, rec)
	if meta.IsSuccess {
		t.Error("IsSuccess = true, want false")
	}
}

func TestReadyz_RedisUnavailable(t *testing.T) {
	e := newEcho()
	deps := &provider.Deps{DB: newSQLiteDB(t)}
	provider.RegisterRoutes(e, deps)

	rec := serve(t, e, http.MethodGet, "/readyz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

// ─── 404 ─────────────────────────────────────────────────────────────────────

func TestRouteNotFound(t *testing.T) {
	e := newEcho()
	provider.RegisterRoutes(e, &provider.Deps{})

	rec := serve(t, e, http.MethodGet, "/nonexistent")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if decodeMeta(t, rec).IsSuccess {
		t.Error("IsSuccess = true, want false")
	}
}

// ─── Close ───────────────────────────────────────────────────────────────────

func TestClose(t *testing.T) {
	t.Run("nil dependencies", func(t *testing.T) {
		var deps *provider.Deps
		if err := deps.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})

	t.Run("closes database and Redis", func(t *testing.T) {
		db := newSQLiteDB(t)
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("DB() error = %v", err)
		}
		rdb := newMiniRedis(t)

		deps := &provider.Deps{DB: db, Redis: rdb}
		if err := deps.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if err := sqlDB.Ping(); err == nil {
			t.Error("database still accepts connections after Close()")
		}
		if err := rdb.Ping(context.Background()).Err(); !errors.Is(err, redis.ErrClosed) {
			t.Errorf("Redis Ping() error = %v, want %v", err, redis.ErrClosed)
		}
	})

	t.Run("continues after a close error", func(t *testing.T) {
		db := newSQLiteDB(t)
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("DB() error = %v", err)
		}
		rdb := newMiniRedis(t)
		if err := rdb.Close(); err != nil {
			t.Fatalf("Redis Close() error = %v", err)
		}

		err = (&provider.Deps{DB: db, Redis: rdb}).Close()
		if err == nil || !strings.Contains(err.Error(), "redis") {
			t.Fatalf("Close() error = %v, want a Redis close error", err)
		}
		if err := sqlDB.Ping(); err == nil {
			t.Error("database still accepts connections after Close()")
		}
	})
}

// ─── routes registration ────────────────────────────────────────────────────

func TestRegisterRoutes_AddsHealthEndpoints(t *testing.T) {
	e := newEcho()
	deps := &provider.Deps{DB: newSQLiteDB(t), Redis: newMiniRedis(t)}
	provider.RegisterRoutes(e, deps)

	for _, path := range []string{"/healthz", "/readyz"} {
		rec := serve(t, e, http.MethodGet, path)
		if rec.Code == http.StatusNotFound {
			t.Errorf("%s returned 404, route not registered", path)
		}
	}
}

func TestRegisterRoutes_NoPanic_WithNilDeps(t *testing.T) {
	e := newEcho()
	provider.RegisterRoutes(e, &provider.Deps{})
}
