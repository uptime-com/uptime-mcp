package server

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

type RunParams struct {
	fx.In
	fx.Lifecycle
	Logger *slog.Logger
	Server *mcp.Server
	Config app.Config
}

func Run(p RunParams) {
	switch p.Config.Transport {
	case "stdio":
		runStdio(p)
	case "http":
		runHTTP(p)
	default:
		p.Logger.Error("unknown transport", "transport", p.Config.Transport)
		os.Exit(1)
	}
}

func runStdio(p RunParams) {
	token := os.Getenv("UPTIME_API_TOKEN")
	if token == "" {
		p.Logger.Error("UPTIME_API_TOKEN environment variable is required for stdio mode")
		os.Exit(1)
	}

	client, err := createUptimeClient(token, p.Config.APIBaseURL)
	if err != nil {
		p.Logger.Error("failed to create Uptime client", "error", err)
		os.Exit(1)
	}

	session := &app.Session{Client: client}
	p.Logger.Info("authenticated with Uptime.com API")

	// Add middleware to inject session into context for all requests
	p.Server.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return next(app.ContextWithSession(ctx, session), method, req)
		}
	})

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p.Logger.Info("starting stdio transport")
			go func() {
				transport := &mcp.StdioTransport{}
				ss, err := p.Server.Connect(ctx, transport, nil)
				if err != nil {
					p.Logger.Error("failed to connect", "error", err)
					return
				}
				if err := ss.Wait(); err != nil {
					p.Logger.Error("session error", "error", err)
				}
			}()
			return nil
		},
	})
}

func runHTTP(p RunParams) {
	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return p.Server },
		nil,
	)

	var h http.Handler = handler
	h = authMiddleware(h, p.Config.APIBaseURL)
	h = accessLog(h)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/", h)

	httpServer := &http.Server{
		Addr:    p.Config.ListenAddr,
		Handler: mux,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p.Logger.Info("starting HTTP server", "addr", p.Config.ListenAddr)
			go httpServer.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("stopping HTTP server")
			return httpServer.Shutdown(ctx)
		},
	})
}

// createUptimeClient creates an Uptime.com API client.
func createUptimeClient(token, baseURL string) (*api.Client, error) {
	config := &api.Config{
		Token:   token,
		BaseURL: baseURL,
	}
	return api.NewClient(config)
}

const defaultAPIBaseURL = "https://uptime.com/api/v1/"

// validateToken checks if the token is valid by making a HEAD request to the API root.
// Returns nil if valid (200 or 405), error otherwise (401 for invalid token).
func validateToken(ctx context.Context, token, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, baseURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMethodNotAllowed {
		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("invalid token")
	}
	return errors.New("unexpected status: " + resp.Status)
}

// authMiddleware validates the bearer token against Uptime.com API
// and attaches the authenticated client to the request context.
func authMiddleware(next http.Handler, baseURL string) http.Handler {
	logger := log.New(os.Stdout, "auth: ", 0)
	cache := newSessionCache(5 * time.Minute)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		// Try Authorization header first
		if auth := r.Header.Get("Authorization"); auth != "" {
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				http.Error(w, "bearer token required", http.StatusUnauthorized)
				return
			}
			token = strings.TrimPrefix(auth, prefix)
		} else {
			// Fall back to query parameter
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}

		// Check cache first
		if session := cache.get(token); session != nil {
			ctx := app.ContextWithSession(r.Context(), session)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Validate token with lightweight HEAD request
		if err := validateToken(r.Context(), token, baseURL); err != nil {
			logger.Println("token validation failed:", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Create Uptime client
		client, err := createUptimeClient(token, baseURL)
		if err != nil {
			logger.Println("failed to create client:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Cache and attach session to context
		session := &app.Session{Client: client}
		cache.set(token, session)
		ctx := app.ContextWithSession(r.Context(), session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// accessLog logs HTTP requests.
func accessLog(next http.Handler) http.Handler {
	logger := log.New(os.Stdout, "access: ", 0)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Println(r.Method, r.URL.Path, rw.status, time.Since(t).Milliseconds())
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
