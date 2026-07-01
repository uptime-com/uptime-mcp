package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"go.uber.org/fx"
	"golang.org/x/oauth2"

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

// runStdio sets up the MCP server for stdio transport.
//
// Authentication is lazy — the OAuth2 browser flow (if needed) runs on the
// first incoming MCP request, not at startup. This matches the behavior of
// other MCP plugins (e.g. Atlassian) and avoids blocking startup with a
// browser prompt.
//
// Token sources (checked in order on first request):
//   - UPTIME_BEARER_TOKEN env var — static token, no browser, no refresh
//   - OAuth2 PKCE browser flow — requires -uptime-url + -client-id
func runStdio(p RunParams) {
	apiBaseURL := p.Config.APIBaseURL()

	p.Server.AddReceivingMiddleware(
		stdioLazyAuthMiddleware(p),
		clientInitMiddleware(apiBaseURL),
	)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
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

// stdioLazyAuthMiddleware returns an MCP middleware that resolves the bearer
// token lazily. If UPTIME_BEARER_TOKEN is set, it's used immediately (no
// browser interaction). Otherwise, the OAuth2 browser flow is deferred until
// the first tool call — protocol methods like "initialize" and "tools/list"
// pass through without triggering auth.
//
// After the initial resolution, subsequent requests use the cached token
// (which may be auto-refreshed in the background for OAuth2 tokens).
func stdioLazyAuthMiddleware(p RunParams) mcp.Middleware {
	// Static bearer token can be resolved eagerly — no browser interaction.
	if bearerToken := os.Getenv("UPTIME_BEARER_TOKEN"); bearerToken != "" {
		p.Logger.Info("using static bearer token from UPTIME_BEARER_TOKEN")
		holder := newTokenHolder(&oauth2.Token{AccessToken: bearerToken})
		return stdioTokenMiddleware(holder)
	}

	// OAuth2 browser flow — validate config now, defer actual flow to first
	// tool call so the server can complete MCP handshake without blocking.
	cfg := stdioOAuthConfig{
		Issuer:       p.Config.OAuthIssuer(),
		ClientID:     p.Config.ClientID,
		ClientSecret: p.Config.ClientSecret,
		Scopes:       []string{"api/v1"},
	}
	if cfg.Issuer == "" || cfg.ClientID == "" {
		p.Logger.Error("-uptime-url (or -oauth-url) and -client-id are required for stdio mode without UPTIME_BEARER_TOKEN")
		os.Exit(1)
	}

	var (
		once    sync.Once
		holder  *tokenHolder
		authErr error
	)

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// Let protocol methods through without auth — they don't need a
			// session and triggering a browser flow for "initialize" or
			// "tools/list" would block the MCP handshake.
			if !methodRequiresAuth(method) {
				return next(ctx, method, req)
			}

			if app.SessionFromContext(ctx) != nil {
				return next(ctx, method, req)
			}

			// Perform OAuth2 browser flow on first tool call only.
			once.Do(func() {
				p.Logger.Info("first tool call received, starting OAuth2 browser flow")
				token, err := stdioOAuthFlow(ctx, p.Logger, cfg)
				if err != nil {
					authErr = fmt.Errorf("OAuth2 authorization failed: %w", err)
					return
				}

				holder = newTokenHolder(token)
				startTokenRefresh(context.Background(), p.Logger, holder, cfg)
				p.Logger.Info("authenticated via OAuth2")
			})
			if authErr != nil {
				return nil, authErr
			}

			token := holder.AccessToken()
			if token == "" {
				return nil, errors.New("no access token available")
			}

			session := &app.Session{Token: token}
			ctx = app.ContextWithSession(ctx, session)
			return next(ctx, method, req)
		}
	}
}

// methodRequiresAuth reports whether an MCP method needs an authenticated
// session. Protocol-level methods (handshake, capability discovery) do not;
// tool calls and resource reads do.
func methodRequiresAuth(method string) bool {
	switch method {
	case "initialize", "ping",
		"tools/list",
		"prompts/list",
		"resources/list", "resources/templates/list",
		"notifications/initialized", "notifications/cancelled":
		return false
	default:
		return true
	}
}

// runHTTP sets up the MCP server for HTTP transport.
//
// Authentication is passthrough: bearer tokens from the Authorization header,
// query parameter, or UPTIME_BEARER_TOKEN env var are forwarded to the Uptime
// API without server-side verification. The API itself rejects invalid tokens.
//
// When -client-id is set, the server also serves RFC 9728 protected resource
// metadata so OAuth2-capable MCP clients can discover the authorization server
// and obtain tokens themselves.
func runHTTP(p RunParams) {
	apiBaseURL := p.Config.APIBaseURL()

	p.Server.AddReceivingMiddleware(
		loggingMiddleware(os.Stderr),
		httpTokenMiddleware(),
		clientInitMiddleware(apiBaseURL),
	)

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return p.Server },
		nil,
	)

	mux := http.NewServeMux()

	// RFC 9728: OAuth 2.0 Protected Resource Metadata.
	// Registered when an OAuth issuer is configured (via -oauth-url or
	// -uptime-url) so that OAuth2-capable clients can discover the
	// authorization server and obtain tokens.
	if issuer := p.Config.OAuthIssuer(); issuer != "" {
		mux.Handle("/.well-known/oauth-protected-resource",
			auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
				Resource:               p.Config.ResourceURL,
				AuthorizationServers:   []string{issuer},
				ScopesSupported:        []string{"api/v1", "api/v1:read"},
				BearerMethodsSupported: []string{"header"},
			}))
	}

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.Handle("/", bearerPassthrough(mcpHandler))

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
