package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
	"golang.org/x/oauth2"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// ---------------------------------------------------------------------------
// Bearer passthrough (HTTP middleware)
// ---------------------------------------------------------------------------

// passthroughTokenKey is the context key for bearer tokens injected by
// the passthrough middleware.
type passthroughTokenKey struct{}

// extractBearerToken extracts a bearer token from the request using the
// passthrough priority order: Authorization header → query param → env var.
// Returns empty string if no token is found.
func extractBearerToken(r *http.Request) string {
	if v := r.Header.Get("Authorization"); strings.HasPrefix(v, "Bearer ") {
		return strings.TrimPrefix(v, "Bearer ")
	}
	if v := r.URL.Query().Get("token"); v != "" {
		return v
	}
	if v := os.Getenv("UPTIME_BEARER_TOKEN"); v != "" {
		return v
	}
	return ""
}

// bearerPassthrough is HTTP middleware that extracts a bearer token from
// multiple sources and injects it into the request context. Returns 401
// if no token is found.
//
// Sources are checked in order (first match wins):
//  1. Authorization: Bearer header
//  2. token= URL query parameter
//  3. UPTIME_BEARER_TOKEN environment variable
func bearerPassthrough(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), passthroughTokenKey{}, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ---------------------------------------------------------------------------
// MCP middleware — session injection
// ---------------------------------------------------------------------------

// httpTokenMiddleware creates an MCP middleware that reads the bearer token
// from the passthrough context key and creates a session from it.
// Used with bearerPassthrough HTTP middleware.
func httpTokenMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if app.SessionFromContext(ctx) != nil {
				return next(ctx, method, req)
			}

			token, _ := ctx.Value(passthroughTokenKey{}).(string)
			if token == "" {
				return nil, errors.New("authorization required")
			}

			session := &app.Session{Token: token}
			ctx = app.ContextWithSession(ctx, session)
			return next(ctx, method, req)
		}
	}
}

// stdioTokenMiddleware creates an MCP middleware that injects a session with
// the current OAuth2 access token. The token is refreshed in the background;
// this middleware always uses the latest token from the holder.
func stdioTokenMiddleware(holder *tokenHolder) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if app.SessionFromContext(ctx) != nil {
				return next(ctx, method, req)
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

// ---------------------------------------------------------------------------
// Client initialization (shared between all modes)
// ---------------------------------------------------------------------------

// clientInitMiddleware creates an MCP middleware that initializes the Uptime
// API client for the current session. This is shared between HTTP and stdio.
func clientInitMiddleware(apiBaseURL string) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			session := app.SessionFromContext(ctx)
			if session == nil {
				return nil, errors.New("no session in context")
			}

			if session.Client != nil {
				return next(ctx, method, req)
			}

			client, err := createUptimeClient(session.Token, apiBaseURL)
			if err != nil {
				return nil, err
			}

			session.Client = client
			return next(ctx, method, req)
		}
	}
}

// createUptimeClient creates an Uptime.com API client with bearer token auth.
func createUptimeClient(token, baseURL string) (upapi.API, error) {
	opts := []upapi.Option{upapi.WithBearerToken(token)}
	if baseURL != "" {
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		opts = append(opts, upapi.WithBaseURL(baseURL))
	}
	return upapi.New(opts...)
}

// ---------------------------------------------------------------------------
// Token holder (stdio mode)
// ---------------------------------------------------------------------------

// tokenHolder safely stores an OAuth2 token that may be refreshed in the
// background. Used by stdio mode to share the current token between the
// refresh goroutine and request-handling middleware.
type tokenHolder struct {
	mu    sync.RWMutex
	token *oauth2.Token
}

func newTokenHolder(token *oauth2.Token) *tokenHolder {
	return &tokenHolder{token: token}
}

func (h *tokenHolder) AccessToken() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.token == nil {
		return ""
	}
	return h.token.AccessToken
}

func (h *tokenHolder) Update(token *oauth2.Token) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.token = token
}

func (h *tokenHolder) Token() *oauth2.Token {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.token
}
