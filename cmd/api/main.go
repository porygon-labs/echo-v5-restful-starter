package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go-restful-api/internal/config"
	"github.com/porygon-labs/go-kit/logger"
	"github.com/porygon-labs/go-kit/telemetry"
	"go-restful-api/internal/provider"

	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	// Must be the first thing: load config so we can set up the logger.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Init the global structured logger.
	logger.Init(logger.New(cfg.Logger.Level, cfg.Logger.Format))
	log := logger.Default()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Init OpenTelemetry.
	shutdown, err := telemetry.Init(ctx, cfg.Otel.ServiceName, cfg.Otel.ExporterEndpoint)
	if err != nil {
		log.Error("Failed to initialize telemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Error("Failed to shutdown telemetry", "error", err)
		}
	}()

	// Init shared dependencies (DB, Redis, …).
	deps, err := provider.NewDeps(cfg)
	if err != nil {
		log.Error("Failed to initialize dependencies", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := deps.Close(); err != nil {
			log.Error("Failed to close dependencies", "error", err)
		}
	}()

	e := echo.New()

	// Echo request logger → slog.
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:    true,
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogHost:      true,
		LogUserAgent: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			log.Info("request",
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("host", v.Host),
				slog.String("user_agent", v.UserAgent),
			)
			return nil
		},
	}))

	e.Use(middleware.Recover())
	e.Use(echootel.NewMiddleware(cfg.Otel.ServiceName))

	// Wire and mount all module routes.
	provider.RegisterRoutes(e, deps)

	// Print registered routes.
	printRoutes(e)

	app := echo.StartConfig{
		Address:         fmt.Sprintf(":%d", cfg.App.Port),
		HideBanner:      true,
		GracefulTimeout: 10 * time.Second,
	}

	if err := app.Start(ctx, e); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}

// printRoutes logs all registered routes in method-path order.
func printRoutes(e *echo.Echo) {
	routes := e.Router().Routes()
	if len(routes) == 0 {
		return
	}

	fmt.Println("\n── routes ──────────────────────────────────────────────")
	for _, r := range routes {
		fmt.Printf("  %-7s %s\n", r.Method, r.Path)
	}
	fmt.Println(strings.Repeat("─", 56))
}
