package app

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
)

// Config holds the application configuration from CLI flags.
type Config struct {
	// Transport mode: "http" or "stdio"
	Transport string

	// HTTP server settings (only used when Transport == "http")
	ListenAddr string

	// UptimeURL is the Uptime.com instance URL (e.g., https://uptime.com).
	// Used to derive both the API base URL (UptimeURL + "/api/v1/") and the
	// OAuth2 authorization server endpoints.
	UptimeURL string

	// ResourceURL is this server's public URL for OAuth2 protected resource metadata.
	// Defaults to http://localhost{ListenAddr} when not set.
	ResourceURL string

	// ClientID is the OAuth2 client ID.
	ClientID string

	// ClientSecret is the OAuth2 client secret (confidential clients).
	ClientSecret string

	// LogLevel for application logger (debug, info, warn, error)
	LogLevel *slog.LevelVar
}

// APIBaseURL returns the Uptime.com API base URL derived from UptimeURL.
func (c Config) APIBaseURL() string {
	u := c.UptimeURL
	if u == "" {
		u = "https://uptime.com"
	}
	return strings.TrimRight(u, "/") + "/api/v1/"
}

// provideConfig parses CLI flags and returns application configuration.
func provideConfig() (Config, error) {
	cfg := Config{
		LogLevel: new(slog.LevelVar),
	}
	cfg.LogLevel.Set(slog.LevelError) // default

	flag.StringVar(&cfg.Transport, "transport", "stdio", "Transport mode: stdio or http")
	flag.StringVar(&cfg.ListenAddr, "listen", ":8080", "HTTP listen address (only used with -transport=http)")
	flag.StringVar(&cfg.UptimeURL, "uptime-url", "", "Uptime.com instance URL (e.g., https://uptime.com)")
	flag.StringVar(&cfg.ResourceURL, "resource-url", "", "Public URL of this server (for OAuth2 resource metadata, defaults to http://localhost:{listen})")
	flag.StringVar(&cfg.ClientID, "client-id", "", "OAuth2 client ID")
	flag.StringVar(&cfg.ClientSecret, "client-secret", "", "OAuth2 client secret (confidential clients)")
	flag.TextVar(cfg.LogLevel, "log-level", cfg.LogLevel, "Log level: debug, info, warn, error")
	flag.Parse()

	if cfg.Transport != "stdio" && cfg.Transport != "http" {
		return Config{}, fmt.Errorf("invalid transport: %s (must be stdio or http)", cfg.Transport)
	}

	if cfg.UptimeURL == "" {
		cfg.UptimeURL = os.Getenv("UPTIME_URL")
	}
	if cfg.ResourceURL == "" {
		cfg.ResourceURL = os.Getenv("UPTIME_RESOURCE_URL")
	}
	if cfg.ResourceURL == "" {
		host, port, _ := net.SplitHostPort(cfg.ListenAddr)
		if host == "" || host == "0.0.0.0" || host == "::" {
			host = "localhost"
		}
		if port == "" {
			port = "8080"
		}
		cfg.ResourceURL = "http://" + host + ":" + port
	}
	if cfg.ClientID == "" {
		cfg.ClientID = os.Getenv("UPTIME_OAUTH_CLIENT_ID")
	}
	if cfg.ClientSecret == "" {
		cfg.ClientSecret = os.Getenv("UPTIME_OAUTH_CLIENT_SECRET")
	}

	return cfg, nil
}
