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

	// OAuthIssuer is the OAuth2 authorization server URL (e.g., https://uptime.com).
	// The API base URL is derived as OAuthIssuer + "/api/v1/".
	OAuthIssuer string

	// ResourceURL is this server's public URL for OAuth2 protected resource metadata.
	// Defaults to http://localhost{ListenAddr} when not set.
	ResourceURL string

	// ClientID is the OAuth2 client ID for the stdio browser flow.
	ClientID string

	// ClientSecret is the OAuth2 client secret for the stdio browser flow.
	ClientSecret string

	// LogLevel for application logger (debug, info, warn, error)
	LogLevel *slog.LevelVar
}

// APIBaseURL returns the Uptime.com API base URL derived from OAuthIssuer.
func (c Config) APIBaseURL() string {
	issuer := c.OAuthIssuer
	if issuer == "" {
		issuer = "https://uptime.com"
	}
	return strings.TrimRight(issuer, "/") + "/api/v1/"
}

// provideConfig parses CLI flags and returns application configuration.
func provideConfig() (Config, error) {
	cfg := Config{
		LogLevel: new(slog.LevelVar),
	}
	cfg.LogLevel.Set(slog.LevelError) // default

	flag.StringVar(&cfg.Transport, "transport", "stdio", "Transport mode: stdio or http")
	flag.StringVar(&cfg.ListenAddr, "listen", ":8080", "HTTP listen address (only used with -transport=http)")
	flag.StringVar(&cfg.OAuthIssuer, "oauth-issuer", "", "OAuth2 authorization server URL (e.g., https://uptime.com)")
	flag.StringVar(&cfg.ResourceURL, "resource-url", "", "Public URL of this server (for OAuth2 resource metadata, defaults to http://localhost:{listen})")
	flag.StringVar(&cfg.ClientID, "client-id", "", "OAuth2 client ID (stdio mode)")
	flag.StringVar(&cfg.ClientSecret, "client-secret", "", "OAuth2 client secret (stdio mode)")
	flag.TextVar(cfg.LogLevel, "log-level", cfg.LogLevel, "Log level: debug, info, warn, error")
	flag.Parse()

	if cfg.Transport != "stdio" && cfg.Transport != "http" {
		return Config{}, fmt.Errorf("invalid transport: %s (must be stdio or http)", cfg.Transport)
	}

	if cfg.OAuthIssuer == "" {
		cfg.OAuthIssuer = os.Getenv("UPTIME_OAUTH_ISSUER")
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
