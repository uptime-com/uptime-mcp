package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

var Module = fx.Module("server",
	fx.Provide(New),
	fx.Invoke(Run),
)

func New() *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{
		Name:    "uptime-mcp",
		Version: "0.1.0",
	}, nil)
}

type RunParams struct {
	fx.In
	fx.Lifecycle
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
		log.Fatalf("unknown transport: %s", p.Config.Transport)
	}
}

func runStdio(p RunParams) {
	logger := log.New(os.Stderr, "[uptime-mcp] ", log.LstdFlags)

	// For stdio mode, get token from environment and add middleware to inject session
	token := os.Getenv("UPTIME_API_TOKEN")
	if token == "" {
		logger.Printf("warning: UPTIME_API_TOKEN not set")
	}

	var session *app.Session
	if token != "" {
		client, err := createUptimeClient(token, p.Config.APIBaseURL)
		if err != nil {
			logger.Printf("failed to create Uptime client: %v", err)
		} else {
			session = &app.Session{Client: client}
		}
	}

	// Add middleware to inject session into context for all requests
	p.Server.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if session != nil {
				ctx = app.ContextWithSession(ctx, session)
			}
			return next(ctx, method, req)
		}
	})

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				transport := &mcp.StdioTransport{}
				ss, err := p.Server.Connect(ctx, transport, nil)
				if err != nil {
					logger.Printf("failed to connect: %v", err)
					return
				}
				if err := ss.Wait(); err != nil {
					logger.Printf("session error: %v", err)
				}
			}()
			return nil
		},
	})
}

func runHTTP(p RunParams) {
	logger := log.New(os.Stderr, "[uptime-mcp] ", log.LstdFlags)

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return p.Server },
		nil,
	)

	var h http.Handler = handler
	h = authMiddleware(h, p.Config.APIBaseURL, logger)
	h = accessLog(h, logger)

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
			logger.Printf("starting HTTP server on %s", p.Config.ListenAddr)
			go httpServer.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return httpServer.Shutdown(ctx)
		},
	})
}

// createUptimeClient creates and validates an Uptime.com API client.
func createUptimeClient(token, baseURL string) (*api.Client, error) {
	config := &api.Config{
		Token:   token,
		BaseURL: baseURL,
	}
	return api.NewClient(config)
}

// authMiddleware validates the bearer token against Uptime.com API
// and attaches the authenticated client to the request context.
func authMiddleware(next http.Handler, baseURL string, logger *log.Logger) http.Handler {
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

		// Create Uptime client and validate token by making a simple API call
		client, err := createUptimeClient(token, baseURL)
		if err != nil {
			logger.Printf("failed to create client: %v", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Validate token by listing checks (with minimal page size)
		_, _, err = client.Checks.List(r.Context(), &api.CheckListOptions{PageSize: 1})
		if err != nil {
			logger.Printf("token validation failed: %v", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Attach session to context
		session := &app.Session{Client: client}
		ctx := app.ContextWithSession(r.Context(), session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// accessLog logs HTTP requests.
func accessLog(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Printf("%s %s %d", r.Method, r.URL.Path, rw.status)
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
