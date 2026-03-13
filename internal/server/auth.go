package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
	"golang.org/x/oauth2"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// uptimeTokenVerifier returns a TokenVerifier that validates bearer tokens against
// the Uptime.com API by making a HEAD request to the API base URL.
func uptimeTokenVerifier(apiBaseURL string) auth.TokenVerifier {
	return func(ctx context.Context, token string, req *http.Request) (*auth.TokenInfo, error) {
		url := apiBaseURL
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}

		headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		if err != nil {
			return nil, err
		}
		headReq.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(headReq)
		if err != nil {
			return nil, err
		}
		_ = resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK, http.StatusMethodNotAllowed:
			return &auth.TokenInfo{
				Expiration: time.Now().Add(5 * time.Minute),
				Extra:      map[string]any{"token": token},
			}, nil
		case http.StatusUnauthorized:
			return nil, auth.ErrInvalidToken
		default:
			return nil, errors.New("unexpected status: " + resp.Status)
		}
	}
}

// httpTokenMiddleware creates an MCP middleware that extracts the bearer token
// from the HTTP request context (set by auth.RequireBearerToken) and creates a session.
func httpTokenMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			session := app.SessionFromContext(ctx)
			if session != nil {
				return next(ctx, method, req)
			}

			tokenInfo := auth.TokenInfoFromContext(ctx)
			if tokenInfo == nil {
				return nil, errors.New("authorization required")
			}

			token, _ := tokenInfo.Extra["token"].(string)
			if token == "" {
				return nil, errors.New("authorization required")
			}

			session = &app.Session{Token: token}
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
			session := app.SessionFromContext(ctx)
			if session != nil {
				return next(ctx, method, req)
			}

			token := holder.AccessToken()
			if token == "" {
				return nil, errors.New("no access token available")
			}

			session = &app.Session{Token: token}
			ctx = app.ContextWithSession(ctx, session)
			return next(ctx, method, req)
		}
	}
}

// clientInitMiddleware creates an MCP middleware that initializes the Uptime client.
// This is shared between HTTP and stdio modes.
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

// tokenHolder safely stores an OAuth2 token that may be refreshed in the background.
type tokenHolder struct {
	mu    sync.RWMutex
	token *oauth2.Token
}

// newTokenHolder creates a tokenHolder with an initial token.
func newTokenHolder(token *oauth2.Token) *tokenHolder {
	return &tokenHolder{token: token}
}

// AccessToken returns the current access token string.
func (h *tokenHolder) AccessToken() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.token == nil {
		return ""
	}
	return h.token.AccessToken
}

// Update replaces the stored token with a new one.
func (h *tokenHolder) Update(token *oauth2.Token) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.token = token
}

// Token returns the current token.
func (h *tokenHolder) Token() *oauth2.Token {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.token
}
