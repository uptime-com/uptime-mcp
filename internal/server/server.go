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
// Token is resolved once at startup from one of two sources:
//   - UPTIME_BEARER_TOKEN env var — static token, no browser, no refresh
//   - OAuth2 PKCE browser flow — requires -uptime-url + -client-id
func runStdio(p RunParams) {
	apiBaseURL := p.Config.APIBaseURL()

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			holder, err := stdioTokenHolder(ctx, p)
			if err != nil {
				return err
			}

			p.Server.AddReceivingMiddleware(
				stdioTokenMiddleware(holder),
				clientInitMiddleware(apiBaseURL),
			)

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

// stdioTokenHolder resolves the token for stdio mode.
// Returns a tokenHolder with either a static token or an OAuth2-obtained token.
func stdioTokenHolder(ctx context.Context, p RunParams) (*tokenHolder, error) {
	// Static bearer token — skip OAuth2 entirely
	if bearerToken := os.Getenv("UPTIME_BEARER_TOKEN"); bearerToken != "" {
		p.Logger.Info("using static bearer token from UPTIME_BEARER_TOKEN")
		return newTokenHolder(&oauth2.Token{AccessToken: bearerToken}), nil
	}

	// OAuth2 browser flow — requires both -uptime-url and -client-id
	cfg := stdioOAuthConfig{
		Issuer:       p.Config.UptimeURL,
		ClientID:     p.Config.ClientID,
		ClientSecret: p.Config.ClientSecret,
		Scopes:       []string{"api/v1"},
	}
	if cfg.Issuer == "" || cfg.ClientID == "" {
		p.Logger.Error("-uptime-url and -client-id are required for stdio mode without UPTIME_BEARER_TOKEN")
		os.Exit(1)
	}

	token, err := stdioOAuthFlow(ctx, p.Logger, cfg)
	if err != nil {
		return nil, fmt.Errorf("OAuth2 authorization failed: %w", err)
	}

	holder := newTokenHolder(token)
	startTokenRefresh(ctx, p.Logger, holder, cfg)
	p.Logger.Info("authenticated via OAuth2")
	return holder, nil
}

// runHTTP sets up the MCP server for HTTP transport.
//
// Authentication strategy is determined by flags:
//   - With -client-id: OAuth2 session auth via RequireBearerToken (RFC 9728)
//   - Without -client-id: bearer passthrough (header → query → env, no verification)
//   - With -bearer-passthrough alongside -client-id: both paths enabled
func runHTTP(p RunParams) {
	apiBaseURL := p.Config.APIBaseURL()

	// Choose HTTP and MCP middleware based on auth strategy.
	httpAuth, mcpAuth := httpAuthMiddleware(p)

	p.Server.AddReceivingMiddleware(
		loggingMiddleware(os.Stderr),
		mcpAuth,
		clientInitMiddleware(apiBaseURL),
	)

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return p.Server },
		nil,
	)

	mux := http.NewServeMux()

	// RFC 9728: OAuth 2.0 Protected Resource Metadata.
	// Always registered when -uptime-url is set so that OAuth2-capable clients
	// can discover the authorization server, even in passthrough mode.
	if p.Config.UptimeURL != "" {
		mux.Handle("/.well-known/oauth-protected-resource",
			auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
				Resource:               p.Config.ResourceURL,
				AuthorizationServers:   []string{p.Config.UptimeURL},
				ScopesSupported:        []string{"api/v1", "api/v1:read"},
				BearerMethodsSupported: []string{"header"},
			}))
	}

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/", httpAuth(mcpHandler))

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

// httpAuthMiddleware selects the HTTP-level and MCP-level auth middleware
// based on the configured flags.
//
// Three possible configurations:
//  1. No -client-id → passthrough only (implicit, no flag needed)
//  2. -client-id set → OAuth2 token verification only
//  3. -client-id + -bearer-passthrough → passthrough first, OAuth2 fallback
func httpAuthMiddleware(p RunParams) (func(http.Handler) http.Handler, mcp.Middleware) {
	hasOAuth := p.Config.ClientID != ""
	hasPassthrough := p.Config.BearerPassthrough || !hasOAuth

	switch {
	case hasPassthrough && hasOAuth:
		// Both paths: try passthrough first, fall back to OAuth2 verification.
		p.Logger.Info("HTTP auth: bearer passthrough + OAuth2 verification")
		return bearerPassthroughWithOAuthFallback(p), httpTokenMiddleware()

	case hasOAuth:
		// OAuth2 only: validate all tokens against the Uptime API.
		p.Logger.Info("HTTP auth: OAuth2 verification")
		return oauthTokenVerification(p), oauthSessionMiddleware()

	default:
		// Passthrough only: forward tokens without verification.
		p.Logger.Info("HTTP auth: bearer passthrough")
		return bearerPassthrough, httpTokenMiddleware()
	}
}

// oauthTokenVerification returns HTTP middleware that validates bearer tokens
// against the Uptime API and sets TokenInfo in the request context.
func oauthTokenVerification(p RunParams) func(http.Handler) http.Handler {
	apiBaseURL := p.Config.APIBaseURL()
	resourceMetadataURL := p.Config.ResourceURL + "/.well-known/oauth-protected-resource"

	return auth.RequireBearerToken(
		uptimeTokenVerifier(apiBaseURL),
		&auth.RequireBearerTokenOptions{
			ResourceMetadataURL: resourceMetadataURL,
		},
	)
}

// bearerPassthroughWithOAuthFallback returns HTTP middleware that tries
// passthrough token extraction first (header → query → env), and falls back
// to OAuth2 token verification when no passthrough token is found.
func bearerPassthroughWithOAuthFallback(p RunParams) func(http.Handler) http.Handler {
	oauthMiddleware := oauthTokenVerification(p)

	return func(next http.Handler) http.Handler {
		// Passthrough path: if a token is found, inject it and proceed.
		passthroughHandler := bearerPassthroughOptional(next)

		// OAuth2 path: verify the token and proceed.
		oauthHandler := oauthMiddleware(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try passthrough first. If a token was found, it's in the context.
			if token := extractBearerToken(r); token != "" {
				passthroughHandler.ServeHTTP(w, r)
				return
			}

			// No passthrough token — fall back to OAuth2 verification.
			oauthHandler.ServeHTTP(w, r)
		})
	}
}
