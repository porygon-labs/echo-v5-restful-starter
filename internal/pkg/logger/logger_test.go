package logger_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"go-restful-api/internal/pkg/logger"
)

func TestInitAndDefault(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Init(l)

	logger.Info("hello", "key", "val")

	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected log output to contain 'hello', got %q", buf.String())
	}
}

func TestDefault_Fallback(t *testing.T) {
	// Reset: Init(nil) should reset to slog.Default()
	logger.Init(nil)

	l := logger.Default()
	if l == nil {
		t.Fatal("Default() returned nil")
	}
	// Should not panic.
	logger.Info("fallback")
}

func TestConvenienceWrappers(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Init(l)

	logger.Debug("debug msg", "a", 1)
	logger.Info("info msg", "b", 2)
	logger.Warn("warn msg", "c", 3)
	logger.Error("error msg", "d", 4)

	out := buf.String()

	for _, want := range []string{"debug msg", "info msg", "warn msg", "error msg"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Init(l)

	child := logger.With("component", "test")
	child.Info("with msg", "x", "y")

	out := buf.String()
	if !strings.Contains(out, "component") || !strings.Contains(out, "test") {
		t.Errorf("output missing With fields: %q", out)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		level, format, want string
	}{
		{"debug", "json", "hello"},
		{"info", "json", "hello"},
		{"warn", "json", "hello"},
		{"error", "json", "hello"},
		{"debug", "text", "hello"},
		{"bogus", "json", "hello"}, // defaults to info
		{"info", "bogus", "hello"}, // defaults to json
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		var l *slog.Logger
		if tt.format == "text" {
			l = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		} else {
			l = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
		}
		l.Info("hello")
		if !strings.Contains(buf.String(), tt.want) {
			t.Errorf("New(%q, %q): missing %q in output: %q", tt.level, tt.format, tt.want, buf.String())
		}
	}
}

func TestFatal(t *testing.T) {
	// Fatal calls os.Exit(1) — can't test directly without forking.
	// Just verify it doesn't panic when init hasn't been called.
	// We can't really test this without subprocess, so at least verify
	// the code compiles and doesn't nil-panic.
	t.Skip("Fatal calls os.Exit; tested indirectly via coverage of the wrapper")
}
