// Package logger provides a single, globally-accessible slog.Logger.
//
// Call Init once during startup (typically from main), then import this
// package anywhere and call logger.Info, logger.Error, etc.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

var log *slog.Logger

// Init sets the global logger. Must be called once at startup before any
// logging calls. Passing nil resets to the default slog.Logger.
func Init(l *slog.Logger) {
	if l == nil {
		l = slog.Default()
	}
	log = l
}

// Default returns the current global logger. Returns slog.Default() if
// Init has not been called.
func Default() *slog.Logger {
	if log == nil {
		return slog.Default()
	}
	return log
}

// ─── convenience wrappers ───────────────────────────────────────────────────

// Info logs at LevelInfo.
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// Error logs at LevelError.
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// Warn logs at LevelWarn.
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// Debug logs at LevelDebug.
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// Fatal logs at LevelError then exits with code 1.
func Fatal(msg string, args ...any) {
	Default().Error(msg, args...)
	os.Exit(1)
}

// With returns a Logger that includes the given key-value pairs.
func With(args ...any) *slog.Logger {
	return Default().With(args...)
}

// ─── factory ────────────────────────────────────────────────────────────────

// New creates a slog.Logger configured with the given level and format.
//
//	level  — "debug", "info", "warn", "error"
//	format — "json" or "text"
func New(level, format string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
