package app

import (
	"log/slog"
	"os"
)

// provideLogger creates the application logger based on configuration.
func provideLogger(cfg Config) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
}
