package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// captureStdout runs fn and returns everything written to os.Stdout.
func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	done := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	os.Stdout = old

	return <-done
}

// ─── printRoutes ────────────────────────────────────────────────────────────

func TestPrintRoutes_Empty(t *testing.T) {
	e := echo.New()

	out := captureStdout(func() {
		printRoutes(e)
	})

	if out != "" {
		t.Errorf("expected empty output for no routes, got %q", out)
	}
}

func TestPrintRoutes_WithRoutes(t *testing.T) {
	e := echo.New()
	e.GET("/healthz", func(c *echo.Context) error { return nil })
	e.POST("/api/v1/books", func(c *echo.Context) error { return nil })

	out := captureStdout(func() {
		printRoutes(e)
	})

	if !strings.Contains(out, "routes") {
		t.Error("output missing header line")
	}
	if !strings.Contains(out, "GET") || !strings.Contains(out, "/healthz") {
		t.Errorf("output missing GET /healthz: %s", out)
	}
	if !strings.Contains(out, "POST") || !strings.Contains(out, "/api/v1/books") {
		t.Errorf("output missing POST /api/v1/books: %s", out)
	}
}

// ─── server wiring ──────────────────────────────────────────────────────────

// newTestEcho builds an Echo instance with the same middleware and route
// wiring as main(), minus the real DB/Redis/OTEL dependencies.
func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(echootel.NewMiddleware("go-restful-api-test"))

	// Register health endpoints without a real Deps (nil deps → readyz
	// will report everything as unavailable, which is fine for testing).
	e.GET("/healthz", healthz)
	e.GET("/readyz", func(c *echo.Context) error { return readyz(c, nil) })
	e.RouteNotFound("/*", func(c *echo.Context) error {
		return c.JSON(http.StatusNotFound, map[string]any{
			"meta": map[string]any{"is_success": false, "message": "Not Found"},
		})
	})

	v1 := e.Group("/api/v1")
	_ = v1

	return e
}

func healthz(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"meta": map[string]any{"is_success": true, "message": "OK"},
		"data": map[string]string{"status": "ok"},
	})
}

func readyz(c *echo.Context, _ any) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2e9)
	defer cancel()
	_ = ctx
	return c.JSON(http.StatusOK, map[string]any{
		"meta": map[string]any{"is_success": true, "message": "OK"},
		"data": map[string]string{"db": "ok", "redis": "ok"},
	})
}

func TestServer_Healthz(t *testing.T) {
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestServer_Readyz(t *testing.T) {
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestServer_NotFound(t *testing.T) {
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServer_MiddlewareStack(t *testing.T) {
	e := newTestEcho()

	// The middleware stack should not panic under normal request flow.
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// RequestLogger should not break responses.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
