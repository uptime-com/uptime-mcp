package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
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
	apiBaseURL := p.Config.APIBaseURL()

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Perform OAuth2 browser flow
			oauthCfg := stdioOAuthConfig{
				Issuer:       p.Config.OAuthIssuer,
				ClientID:     p.Config.ClientID,
				ClientSecret: p.Config.ClientSecret,
				Scopes:       []string{"api/v1"},
			}

			if oauthCfg.Issuer == "" || oauthCfg.ClientID == "" {
				p.Logger.Error("-oauth-issuer and -client-id are required for stdio mode")
				os.Exit(1)
			}

			token, err := stdioOAuthFlow(ctx, p.Logger, oauthCfg)
			if err != nil {
				return fmt.Errorf("OAuth2 authorization failed: %w", err)
			}

			holder := newTokenHolder(token)

			// Start background token refresh
			startTokenRefresh(ctx, p.Logger, holder, oauthCfg)

			// Add middlewares: inject token → initialize client
			p.Server.AddReceivingMiddleware(
				stdioTokenMiddleware(holder),
				clientInitMiddleware(apiBaseURL),
			)

			p.Logger.Info("authenticated via OAuth2")

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
	apiBaseURL := p.Config.APIBaseURL()
	resourceURL := p.Config.ResourceURL

	// Add middlewares: extract token from context → initialize client
	p.Server.AddReceivingMiddleware(
		loggingMiddleware(os.Stderr),
		httpTokenMiddleware(),
		clientInitMiddleware(apiBaseURL),
	)

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return p.Server },
		nil,
	)

	verifier := uptimeTokenVerifier(apiBaseURL)

	mux := http.NewServeMux()

	// RFC 9728: OAuth 2.0 Protected Resource Metadata
	mux.Handle("/.well-known/oauth-protected-resource",
		auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
			Resource:               resourceURL,
			AuthorizationServers:   []string{p.Config.OAuthIssuer},
			ScopesSupported:        []string{"api/v1", "api/v1:read"},
			BearerMethodsSupported: []string{"header"},
		}))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Protect MCP endpoints with bearer token verification
	resourceMetadataURL := resourceURL + "/.well-known/oauth-protected-resource"
	mux.Handle("/", auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL: resourceMetadataURL,
	})(mcpHandler))

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
