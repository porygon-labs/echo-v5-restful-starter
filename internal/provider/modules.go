package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/porygon-labs/go-kit/response"
	"github.com/porygon-labs/go-kit/telemetry"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes wires every module and mounts its routes onto the Echo router.
func RegisterRoutes(e *echo.Echo, deps *Deps) {
	// ─── 404 ─────────────────────────────────────────────────────────────
	e.RouteNotFound("/*", func(c *echo.Context) error {
		return response.Error(c.Response(), http.StatusNotFound, "Not Found")
	})

	// ─── health ──────────────────────────────────────────────────────────
	e.GET("/healthz", healthz)
	e.GET("/readyz", func(c *echo.Context) error { return readyz(c, deps) })

	v1 := e.Group("/api/v1")
	_ = v1
}

// ─── health handlers ────────────────────────────────────────────────────────

func healthz(c *echo.Context) error {
	ctx, end := telemetry.StartSpan(c.Request().Context())
	defer end()
	_ = ctx

	return response.OK(c.Response(), map[string]string{"status": "ok"})
}

func readyz(c *echo.Context, d *Deps) error {
	ctx, end := telemetry.StartSpan(c.Request().Context())
	defer end()

	timedCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	checks := map[string]string{}
	failing := false

	// db
	if d != nil && d.DB != nil {
		if sqlDB, err := d.DB.DB(); err == nil {
			if err := sqlDB.PingContext(timedCtx); err != nil {
				checks["db"] = err.Error()
				failing = true
			} else {
				checks["db"] = "ok"
			}
		} else {
			checks["db"] = "unavailable"
			failing = true
		}
	} else {
		checks["db"] = "unavailable"
		failing = true
	}

	// redis
	if d != nil && d.Redis != nil {
		if err := d.Redis.Ping(timedCtx).Err(); err != nil {
			checks["redis"] = err.Error()
			failing = true
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "unavailable"
		failing = true
	}

	status := http.StatusOK
	if failing {
		status = http.StatusServiceUnavailable
	}
	return response.Success(c.Response(), status, checks)
}
