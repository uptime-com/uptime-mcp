package app

import (
	"flag"
	"fmt"
	"log/slog"
)

// Config holds the application configuration from CLI flags.
type Config struct {
	// Transport mode: "http" or "stdio"
	Transport string

	// HTTP server settings (only used when Transport == "http")
	ListenAddr string

	// Uptime.com API base URL (default: https://uptime.com/api/v1/)
	APIBaseURL string

	// LogLevel for application logger (debug, info, warn, error)
	LogLevel *slog.LevelVar
}

// provideConfig parses CLI flags and returns application configuration.
func provideConfig() (Config, error) {
	cfg := Config{
		LogLevel: new(slog.LevelVar),
	}
	cfg.LogLevel.Set(slog.LevelError) // default

	flag.StringVar(&cfg.Transport, "transport", "stdio", "Transport mode: stdio or http")
	flag.StringVar(&cfg.ListenAddr, "listen", ":8080", "HTTP listen address (only used with -transport=http)")
	flag.StringVar(&cfg.APIBaseURL, "api-url", "", "Uptime.com API base URL (default: https://uptime.com/api/v1/)")
	flag.TextVar(cfg.LogLevel, "log-level", cfg.LogLevel, "Log level: debug, info, warn, error")
	flag.Parse()

	if cfg.Transport != "stdio" && cfg.Transport != "http" {
		return Config{}, fmt.Errorf("invalid transport: %s (must be stdio or http)", cfg.Transport)
	}

	return cfg, nil
}
