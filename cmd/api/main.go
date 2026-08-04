package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-restful-api/internal/config"

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
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	app := echo.StartConfig{
		Address:         fmt.Sprintf(":%d", cfg.App.Port),
		HideBanner:      true,
		GracefulTimeout: 10 * time.Second,
	}

	if err := app.Start(ctx, e); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
