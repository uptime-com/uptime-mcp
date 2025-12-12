package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

const headerUptimeAPIKey = "X-Uptime-API-Key"

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
	apiKey := os.Getenv("UPTIME_API_KEY")
	if apiKey == "" {
		p.Logger.Error("UPTIME_API_KEY environment variable is required for stdio mode")
		os.Exit(1)
	}

	// Add middlewares: inject API key → initialize client
	p.Server.AddReceivingMiddleware(stdioKeyMiddleware(apiKey))
	p.Server.AddReceivingMiddleware(clientInitMiddleware(p.Config.APIBaseURL))

	p.Logger.Info("configured with Uptime.com API key")

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
	// Add middlewares: extract API key from header → initialize client
	p.Server.AddReceivingMiddleware(
		loggingMiddleware(os.Stderr),
		httpKeyMiddleware(),
		clientInitMiddleware(p.Config.APIBaseURL),
	)

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return p.Server },
		nil,
	)

	// Wrap with API key extraction (validation happens at MCP level)
	var h http.Handler = extractAPIKey(handler)

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
			go func() {
				_ = httpServer.ListenAndServe()
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("stopping HTTP server")
			return httpServer.Shutdown(ctx)
		},
	})
}
