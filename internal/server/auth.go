package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// stdioKeyMiddleware creates an MCP middleware that injects the API key into session.
// Used in STDIO mode where the key comes from environment variable.
func stdioKeyMiddleware(apiKey string) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			session := app.SessionFromContext(ctx)
			if session == nil {
				session = &app.Session{APIKey: apiKey}
				ctx = app.ContextWithSession(ctx, session)
			}
			return next(ctx, method, req)
		}
	}
}

// httpKeyMiddleware creates an MCP middleware that extracts the API key from HTTP header.
// Used in HTTP mode where the key is passed via X-Uptime-API-Key header.
func httpKeyMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			session := app.SessionFromContext(ctx)
			if session == nil {
				apiKey := req.GetExtra().Header.Get(headerUptimeAPIKey)
				if apiKey == "" {
					return nil, errors.New("authorization required")
				}
				session = &app.Session{APIKey: apiKey}
				ctx = app.ContextWithSession(ctx, session)
			}
			return next(ctx, method, req)
		}
	}
}

// clientInitMiddleware creates an MCP middleware that initializes the Uptime client.
// This is shared between STDIO and HTTP modes.
func clientInitMiddleware(baseURL string) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			session := app.SessionFromContext(ctx)
			if session == nil {
				return nil, errors.New("no session in context")
			}

			// Skip if client already initialized
			if session.Client != nil {
				return next(ctx, method, req)
			}

			// Validate API key
			if err := validateAPIKey(ctx, session.APIKey, baseURL); err != nil {
				return nil, err
			}

			// Create client
			client, err := createUptimeClient(session.APIKey, baseURL)
			if err != nil {
				return nil, err
			}

			session.Client = client
			return next(ctx, method, req)
		}
	}
}

// validateAPIKey checks if the API key is valid by making a HEAD request to the API root.
// Returns nil if valid (200 or 405), error otherwise (401 for invalid key).
func validateAPIKey(ctx context.Context, apiKey, baseURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, baseURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMethodNotAllowed {
		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("invalid API key")
	}
	return errors.New("unexpected status: " + resp.Status)
}

// extractAPIKey extracts the API key from the request and sets it in the internal header.
func extractAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var apiKey string

		// Try Authorization header first
		if auth := r.Header.Get("Authorization"); auth != "" {
			const prefix = "Bearer "
			if strings.HasPrefix(auth, prefix) {
				apiKey = strings.TrimPrefix(auth, prefix)
			}
			r.Header.Del("Authorization")
		}

		// Fall back to query parameter
		if apiKey == "" {
			apiKey = r.URL.Query().Get("key")
		}

		// Set internal header (validation happens at MCP level)
		r.Header.Set(headerUptimeAPIKey, apiKey)

		next.ServeHTTP(w, r)
	})
}

// createUptimeClient creates an Uptime.com API client.
func createUptimeClient(apiKey, baseURL string) (*api.Client, error) {
	config := &api.Config{
		Token:   apiKey,
		BaseURL: baseURL,
	}
	return api.NewClient(config)
}
