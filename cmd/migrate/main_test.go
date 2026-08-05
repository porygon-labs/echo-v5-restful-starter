package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-restful-api/internal/config"

	"github.com/pressly/goose/v3"
)

// tempMigrationsDir creates a temp directory and writes a minimal .env file
// so that config.Load() works. It returns the path to the migrations dir and
// a cleanup function.
func tempMigrationsDir(t *testing.T) (migrationsDir string, cleanup func()) {
	t.Helper()

	root := t.TempDir()
	migrationsDir = filepath.Join(root, "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}

	// Write a .env so config.Load doesn't complain.
	envContent := "DB_DSN=test.db\nREDIS_URL=redis://localhost:6379/0\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(envContent), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	prev, _ := os.Getwd()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cleanup = func() {
		_ = os.Chdir(prev)
	}

	return
}

// ─── create ─────────────────────────────────────────────────────────────────

func TestCreateMigration(t *testing.T) {
	migrationsDir, cleanup := tempMigrationsDir(t)
	defer cleanup()

	err := goose.Create(nil, migrationsDir, "add_users", "sql")
	if err != nil {
		t.Fatalf("goose.Create() error = %v", err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*_add_users.sql"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 migration file, got %d", len(files))
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "+goose Up") {
		t.Error("migration missing Up section")
	}
	if !strings.Contains(s, "+goose Down") {
		t.Error("migration missing Down section")
	}
}

// ─── up / down / status ─────────────────────────────────────────────────────

func TestRunGoose_UpDownStatus(t *testing.T) {
	_, cleanup := tempMigrationsDir(t)
	defer cleanup()

	// Create a test migration that creates and drops a table.
	err := goose.Create(nil, "migrations", "create_test_table", "sql")
	if err != nil {
		t.Fatalf("goose.Create() error = %v", err)
	}

	files, _ := filepath.Glob("migrations/*_create_test_table.sql")
	if len(files) != 1 {
		t.Fatalf("expected 1 migration file, got %d", len(files))
	}

	upSQL := `
-- +goose Up
CREATE TABLE test_items (id INTEGER PRIMARY KEY, name TEXT);

-- +goose Down
DROP TABLE test_items;
`
	if err := os.WriteFile(files[0], []byte(upSQL), 0o600); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	dsn := "file:" + dbPath + "?_pragma=busy_timeout=5000"

	cfg := config.Config{
		DB:     config.DBConfig{Driver: "sqlite3", DSN: dsn},
		Logger: config.LoggerConfig{Level: "info", Format: "json"},
	}

	// ── up ──
	if err := runGoose(cfg, "up", nil); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	// Verify the table exists.
	db, err := openTestDB("sqlite3", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM test_items").Scan(&count); err != nil {
		t.Fatalf("query test_items: %v", err)
	}
	db.Close()

	// ── status ──
	if err := runGoose(cfg, "status", nil); err != nil {
		t.Fatalf("goose status: %v", err)
	}

	// ── down ──
	if err := runGoose(cfg, "down", nil); err != nil {
		t.Fatalf("goose down: %v", err)
	}

	// Verify the table is gone.
	db, err = openTestDB("sqlite3", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM test_items").Scan(&count)
	if err == nil {
		t.Error("expected table to be dropped, but query succeeded")
	}
	db.Close()
}

func TestRunGoose_EmptyMigrations(t *testing.T) {
	_, cleanup := tempMigrationsDir(t)
	defer cleanup()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	cfg := config.Config{
		DB:     config.DBConfig{Driver: "sqlite3", DSN: "file:" + dbPath},
		Logger: config.LoggerConfig{Level: "info", Format: "json"},
	}

	// Up with an empty migrations directory should return a clear error.
	err := runGoose(cfg, "up", nil)
	if err == nil {
		t.Fatal("expected error for empty migrations directory")
	}
	if !strings.Contains(err.Error(), "no migration") {
		t.Errorf("error = %v, want 'no migration'", err)
	}
}

func TestRunGoose_BadCommand(t *testing.T) {
	_, cleanup := tempMigrationsDir(t)
	defer cleanup()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	cfg := config.Config{
		DB:     config.DBConfig{Driver: "sqlite3", DSN: "file:" + dbPath},
		Logger: config.LoggerConfig{Level: "info", Format: "json"},
	}

	err := runGoose(cfg, "bogus-command", nil)
	if err == nil {
		t.Fatal("expected error for bogus command")
	}
	if !strings.Contains(err.Error(), "goose") {
		t.Errorf("error = %v, want 'goose'", err)
	}
}
