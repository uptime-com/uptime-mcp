package server

import (
	"context"
	"errors"
	"net/http"
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

// passthroughTokenKey is the context key for the access token injected by the
// passthrough middleware.
type passthroughTokenKey struct{}

// extractToken returns the OAuth2 access token from the request's
// Authorization: Bearer header, or "" when it carries none.
func extractToken(r *http.Request) string {
	v, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return ""
	}
	return v
}

// bearerPassthrough is HTTP middleware that injects the request's OAuth2
// access token into its context, answering 401 when it carries none.
func bearerPassthrough(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
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

// httpTokenMiddleware creates an MCP middleware that reads the access token
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
			// Protocol methods don't need an API client.
			if !methodRequiresAuth(method) {
				return next(ctx, method, req)
			}

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

// createUptimeClient creates an Uptime.com API client that presents token as
// an OAuth2 bearer token.
func createUptimeClient(token string, baseURL string) (upapi.API, error) {
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
