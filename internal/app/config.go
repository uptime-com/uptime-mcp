package app

// Config holds the application configuration from CLI flags.
type Config struct {
	// Transport mode: "http" or "stdio"
	Transport string

	// HTTP server settings (only used when Transport == "http")
	ListenAddr string

	// Uptime.com API base URL (default: https://uptime.com/api/v1/)
	APIBaseURL string
}
