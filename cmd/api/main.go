package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go-restful-api/internal/config"
	"go-restful-api/internal/provider"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Init shared dependencies (DB, Redis, …).
	deps, err := provider.NewDeps(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}
	defer func() {
		if err := deps.Close(); err != nil {
			log.Printf("Failed to close dependencies: %v", err)
		}
	}()

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

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
		log.Fatalf("Failed to start server: %v", err)
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
