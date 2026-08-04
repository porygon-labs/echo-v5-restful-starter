package provider

import (
	"context"
	"net/http"
	"time"

	"go-restful-api/internal/pkg/response"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes wires every module and mounts its routes onto the Echo router.
func RegisterRoutes(e *echo.Echo, deps *Deps) {
	// ─── 404 ─────────────────────────────────────────────────────────────
	e.RouteNotFound("/*", func(c *echo.Context) error {
		return response.Error(c, http.StatusNotFound, "Not Found")
	})

	// ─── health ──────────────────────────────────────────────────────────
	e.GET("/healthz", healthz)
	e.GET("/readyz", func(c *echo.Context) error { return readyz(c, deps) })

	v1 := e.Group("/api/v1")
	// TODO: remove this line and line below, and register your module here
	_ = v1
}

// ─── health handlers ────────────────────────────────────────────────────────

func healthz(c *echo.Context) error {
	return response.OK(c, map[string]string{"status": "ok"})
}

func readyz(c *echo.Context, d *Deps) error {
	checks := map[string]string{}
	failing := false

	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	// db
	if d != nil && d.DB != nil {
		if sqlDB, err := d.DB.DB(); err == nil {
			if err := sqlDB.PingContext(ctx); err != nil {
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
		if err := d.Redis.Ping(ctx).Err(); err != nil {
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
	return response.Success(c, status, checks)
}
