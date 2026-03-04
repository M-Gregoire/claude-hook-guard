package internal

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	// Create logger based on debug environment variable
	level := slog.LevelInfo
	if os.Getenv("CLAUDE_HOOK_GUARD_DEBUG") != "" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger = slog.New(handler)
}

// Debug logs a debug-level message
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Info logs an info-level message
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Warn logs a warning-level message
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// Error logs an error-level message
func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}
