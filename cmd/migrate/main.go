// Command migrate runs database migrations using goose.
//
// Usage:
//
//	migrate up              run all pending migrations
//	migrate down            roll back the most recent migration
//	migrate down-to VERSION roll back to a specific version
//	migrate status          print migration status
//	migrate create NAME     create a new migration file
//
// The DSN is read from the DB_DSN environment variable (same as the server).
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"go-restful-api/internal/config"
	"github.com/porygon-labs/go-kit/logger"

	"github.com/pressly/goose/v3"

	_ "gorm.io/driver/postgres"
	_ "modernc.org/sqlite"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// "create" doesn't need a DB connection.
	if command == "create" {
		if len(args) < 2 {
			fmt.Println("Usage: migrate create <name>")
			os.Exit(1)
		}
		if err := goose.Create(nil, "migrations", args[1], "sql"); err != nil {
			slog.Error("failed to create migration", "error", err)
			os.Exit(1)
		}
		fmt.Printf("created migration: migrations/%s\n", args[1])
		return
	}

	// All other commands need the DB DSN.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := runGoose(cfg, command, args[1:]); err != nil {
		slog.Error("migration failed", "command", command, "error", err)
		os.Exit(1)
	}
}

// runGoose opens the database and executes a goose command.
func runGoose(cfg config.Config, command string, args []string) error {
	logger.Init(logger.New(cfg.Logger.Level, cfg.Logger.Format))
	log := logger.Default()

	db, err := goose.OpenDBWithDriver(cfg.DB.Driver, cfg.DB.DSN)
	if err != nil {
		log.Error("failed to open database", "error", err, "dsn", cfg.DB.DSN)
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := goose.Run(command, db, "migrations", args...); err != nil {
		return fmt.Errorf("goose %s: %w", command, err)
	}

	return nil
}

// openTestDB is used by tests to open a database with a specific driver.
func openTestDB(driver, dsn string) (*sql.DB, error) {
	return goose.OpenDBWithDriver(driver, dsn)
}

func printUsage() {
	fmt.Println(`migrate — database migration tool

Usage:
  migrate up              run all pending migrations
  migrate down            roll back the most recent migration
  migrate down-to VERSION roll back to a specific version
  migrate status          print migration status
  migrate create NAME     create a new .sql migration file
  migrate redo            redo the last migration (down → up)

The DSN is read from DB_DSN in the environment / .env file.`)
}
