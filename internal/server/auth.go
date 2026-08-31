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

// passthroughCredentialKey is the context key for credentials injected by
// the passthrough middleware.
type passthroughCredentialKey struct{}

// credential is an api/v1 credential together with the scheme its source
// implies.
type credential struct {
	token  string
	scheme app.AuthScheme
}

// extractCredential extracts a credential from the request using the
// passthrough priority order: Authorization header → query param → env var.
// The zero credential means the request carries none.
//
// The scheme is the likelier of the two for that source rather than a reading
// of the credential: an Authorization header is what a client that walked the
// protected-resource metadata sends, and the query parameter and environment
// variable are where an operator puts an account API key. Every source carries
// either credential in practice, so the scheme is corrected against api/v1
// itself when it is wrong, and the guess costs a round trip rather than a
// refusal.
func extractCredential(r *http.Request) credential {
	if v := r.Header.Get("Authorization"); strings.HasPrefix(v, "Bearer ") {
		return credential{strings.TrimPrefix(v, "Bearer "), app.SchemeBearer}
	}
	if v := r.URL.Query().Get("token"); v != "" {
		return credential{v, app.SchemeToken}
	}
	if v := os.Getenv("UPTIME_BEARER_TOKEN"); v != "" {
		return credential{v, app.SchemeToken}
	}
	return credential{}
}

// bearerPassthrough is HTTP middleware that extracts a credential from
// multiple sources and injects it into the request context. Returns 401
// if no credential is found.
//
// Sources are checked in order (first match wins):
//  1. Authorization: Bearer header
//  2. token= URL query parameter
//  3. UPTIME_BEARER_TOKEN environment variable
func bearerPassthrough(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cred := extractCredential(r)
		if cred.token == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), passthroughCredentialKey{}, cred)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ---------------------------------------------------------------------------
// MCP middleware — session injection
// ---------------------------------------------------------------------------

// httpTokenMiddleware creates an MCP middleware that reads the credential
// from the passthrough context key and creates a session from it.
// Used with bearerPassthrough HTTP middleware.
func httpTokenMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if app.SessionFromContext(ctx) != nil {
				return next(ctx, method, req)
			}

			cred, _ := ctx.Value(passthroughCredentialKey{}).(credential)
			if cred.token == "" {
				return nil, errors.New("authorization required")
			}

			session := &app.Session{Token: cred.token, Scheme: cred.scheme}
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

			session := &app.Session{Token: token, Scheme: app.SchemeBearer}
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

			client, err := createUptimeClient(session.Token, session.Scheme, apiBaseURL)
			if err != nil {
				return nil, err
			}

			session.Client = client
			return next(ctx, method, req)
		}
	}
}

// createUptimeClient creates an Uptime.com API client that authenticates with
// token under scheme, falling back to the other scheme if api/v1 refuses it.
func createUptimeClient(token string, scheme app.AuthScheme, baseURL string) (upapi.API, error) {
	opts := []upapi.Option{withUpstreamAuth(token, scheme)}
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
