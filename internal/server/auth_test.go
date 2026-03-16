package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// ---------------------------------------------------------------------------
// Token verification
// ---------------------------------------------------------------------------

func TestUptimeTokenVerifier(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
		wantToken  bool
	}{
		{
			name:       "valid token - 200",
			statusCode: http.StatusOK,
			wantToken:  true,
		},
		{
			name:       "valid token - 405",
			statusCode: http.StatusMethodNotAllowed,
			wantToken:  true,
		},
		{
			name:       "invalid token - 401",
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "server error - 500",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodHead, r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				w.WriteHeader(tt.statusCode)
			}))
			defer ts.Close()

			verifier := uptimeTokenVerifier(ts.URL)
			info, err := verifier(context.Background(), "test-token", nil)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, info)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
			}

			if tt.wantToken {
				token, ok := info.Extra["token"].(string)
				assert.True(t, ok)
				assert.Equal(t, "test-token", token)
			}
		})
	}
}

func TestUptimeTokenVerifierInvalidReturnsErrInvalidToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	verifier := uptimeTokenVerifier(ts.URL)
	_, err := verifier(context.Background(), "bad-token", nil)

	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

// ---------------------------------------------------------------------------
// Protected resource metadata
// ---------------------------------------------------------------------------

func TestProtectedResourceMetadata(t *testing.T) {
	handler := auth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:               "http://localhost:8080",
		AuthorizationServers:   []string{"https://uptime.com"},
		ScopesSupported:        []string{"api/v1", "api/v1:read"},
		BearerMethodsSupported: []string{"header"},
	})

	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "http://localhost:8080", body["resource"])
	assert.Contains(t, body["authorization_servers"], "https://uptime.com")
}

// ---------------------------------------------------------------------------
// RequireBearerToken (SDK middleware)
// ---------------------------------------------------------------------------

func TestRequireBearerToken401(t *testing.T) {
	verifier := func(ctx context.Context, token string, req *http.Request) (*auth.TokenInfo, error) {
		return nil, auth.ErrInvalidToken
	}

	handler := auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL: "https://mcp.example.com/.well-known/oauth-protected-resource",
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Bearer")
}

func TestRequireBearerTokenSetsTokenInfo(t *testing.T) {
	verifier := func(ctx context.Context, token string, req *http.Request) (*auth.TokenInfo, error) {
		return &auth.TokenInfo{
			Expiration: time.Now().Add(5 * time.Minute),
			Extra:      map[string]any{"token": token},
		}, nil
	}

	var capturedTokenInfo *auth.TokenInfo
	handler := auth.RequireBearerToken(verifier, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTokenInfo = auth.TokenInfoFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer my-access-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedTokenInfo)
	assert.Equal(t, "my-access-token", capturedTokenInfo.Extra["token"])
}

// ---------------------------------------------------------------------------
// extractBearerToken
// ---------------------------------------------------------------------------

func TestExtractBearerToken(t *testing.T) {
	t.Run("from Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		assert.Equal(t, "header-token", extractBearerToken(req))
	})

	t.Run("from query param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?token=query-token", nil)
		assert.Equal(t, "query-token", extractBearerToken(req))
	})

	t.Run("from env var", func(t *testing.T) {
		t.Setenv("UPTIME_BEARER_TOKEN", "env-token")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, "env-token", extractBearerToken(req))
	})

	t.Run("header takes precedence over query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?token=query-token", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		assert.Equal(t, "header-token", extractBearerToken(req))
	})

	t.Run("query takes precedence over env", func(t *testing.T) {
		t.Setenv("UPTIME_BEARER_TOKEN", "env-token")
		req := httptest.NewRequest(http.MethodGet, "/?token=query-token", nil)
		assert.Equal(t, "query-token", extractBearerToken(req))
	})

	t.Run("empty when no token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, "", extractBearerToken(req))
	})
}

// ---------------------------------------------------------------------------
// bearerPassthrough (HTTP middleware)
// ---------------------------------------------------------------------------

func TestBearerPassthrough(t *testing.T) {
	t.Run("injects token into context", func(t *testing.T) {
		var capturedToken string
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedToken, _ = r.Context().Value(passthroughTokenKey{}).(string)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		w := httptest.NewRecorder()

		bearerPassthrough(inner).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "header-token", capturedToken)
	})

	t.Run("returns 401 when no token", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		bearerPassthrough(inner).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Bearer")
	})
}

// ---------------------------------------------------------------------------
// bearerPassthroughOptional (HTTP middleware)
// ---------------------------------------------------------------------------

func TestBearerPassthroughOptional(t *testing.T) {
	t.Run("injects token when present", func(t *testing.T) {
		var capturedToken string
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedToken, _ = r.Context().Value(passthroughTokenKey{}).(string)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		w := httptest.NewRecorder()

		bearerPassthroughOptional(inner).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "header-token", capturedToken)
	})

	t.Run("proceeds without token", func(t *testing.T) {
		called := false
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			token, _ := r.Context().Value(passthroughTokenKey{}).(string)
			assert.Empty(t, token)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		bearerPassthroughOptional(inner).ServeHTTP(w, req)

		assert.True(t, called)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ---------------------------------------------------------------------------
// httpTokenMiddleware (MCP middleware — passthrough path)
// ---------------------------------------------------------------------------

func TestHttpTokenMiddleware(t *testing.T) {
	t.Run("preserves existing session", func(t *testing.T) {
		existingSession := &app.Session{Token: "existing-token"}
		ctx := app.ContextWithSession(context.Background(), existingSession)

		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		middleware := httpTokenMiddleware()
		handler := middleware(next)

		_, err := handler(ctx, "test/method", nil)
		require.NoError(t, err)

		session := app.SessionFromContext(capturedCtx)
		assert.Equal(t, "existing-token", session.Token)
	})

	t.Run("creates session from passthrough token", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), passthroughTokenKey{}, "pass-token")

		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		middleware := httpTokenMiddleware()
		handler := middleware(next)

		_, err := handler(ctx, "test/method", nil)
		require.NoError(t, err)

		session := app.SessionFromContext(capturedCtx)
		require.NotNil(t, session)
		assert.Equal(t, "pass-token", session.Token)
	})

	t.Run("returns error without token", func(t *testing.T) {
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			t.Fatal("next should not be called")
			return nil, nil
		}

		middleware := httpTokenMiddleware()
		handler := middleware(next)

		_, err := handler(context.Background(), "test/method", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authorization required")
	})
}

// ---------------------------------------------------------------------------
// oauthSessionMiddleware (MCP middleware — OAuth2 path)
// ---------------------------------------------------------------------------

func TestOAuthSessionMiddleware(t *testing.T) {
	t.Run("preserves existing session", func(t *testing.T) {
		existingSession := &app.Session{Token: "existing-token"}
		ctx := app.ContextWithSession(context.Background(), existingSession)

		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		handler := oauthSessionMiddleware()(next)

		_, err := handler(ctx, "test/method", nil)
		require.NoError(t, err)

		session := app.SessionFromContext(capturedCtx)
		assert.Equal(t, "existing-token", session.Token)
	})

	t.Run("creates session from verified token info", func(t *testing.T) {
		// Simulate what auth.RequireBearerToken does: inject TokenInfo into context.
		// We use the SDK's exported function to set it properly via a test helper.
		tokenInfo := &auth.TokenInfo{
			Expiration: time.Now().Add(5 * time.Minute),
			Extra:      map[string]any{"token": "verified-token"},
		}

		// Use RequireBearerToken to inject TokenInfo, then extract context.
		var capturedCtx context.Context
		verifier := func(ctx context.Context, token string, req *http.Request) (*auth.TokenInfo, error) {
			return tokenInfo, nil
		}
		httpHandler := auth.RequireBearerToken(verifier, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Now we have a context with TokenInfo set by the SDK.
			ctx := r.Context()

			next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
				capturedCtx = ctx
				return nil, nil
			}

			handler := oauthSessionMiddleware()(next)
			_, _ = handler(ctx, "test/method", nil)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer verified-token")
		w := httptest.NewRecorder()
		httpHandler.ServeHTTP(w, req)

		require.NotNil(t, capturedCtx)
		session := app.SessionFromContext(capturedCtx)
		require.NotNil(t, session)
		assert.Equal(t, "verified-token", session.Token)
	})

	t.Run("returns error without token info", func(t *testing.T) {
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			t.Fatal("next should not be called")
			return nil, nil
		}

		handler := oauthSessionMiddleware()(next)

		_, err := handler(context.Background(), "test/method", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no verified token")
	})
}

// ---------------------------------------------------------------------------
// stdioTokenMiddleware (MCP middleware — stdio path)
// ---------------------------------------------------------------------------

func TestStdioTokenMiddleware(t *testing.T) {
	t.Run("injects token from holder", func(t *testing.T) {
		holder := newTokenHolder(&oauth2.Token{AccessToken: "stdio-token"})

		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		middleware := stdioTokenMiddleware(holder)
		handler := middleware(next)

		_, err := handler(context.Background(), "test/method", nil)
		require.NoError(t, err)

		session := app.SessionFromContext(capturedCtx)
		require.NotNil(t, session)
		assert.Equal(t, "stdio-token", session.Token)
	})

	t.Run("preserves existing session", func(t *testing.T) {
		holder := newTokenHolder(&oauth2.Token{AccessToken: "new-token"})
		existingSession := &app.Session{Token: "existing-token"}
		ctx := app.ContextWithSession(context.Background(), existingSession)

		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		middleware := stdioTokenMiddleware(holder)
		handler := middleware(next)

		_, err := handler(ctx, "test/method", nil)
		require.NoError(t, err)

		session := app.SessionFromContext(capturedCtx)
		assert.Equal(t, "existing-token", session.Token)
	})

	t.Run("returns error when no token", func(t *testing.T) {
		holder := newTokenHolder(nil)

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			t.Fatal("next should not be called")
			return nil, nil
		}

		middleware := stdioTokenMiddleware(holder)
		handler := middleware(next)

		_, err := handler(context.Background(), "test/method", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no access token")
	})
}

// ---------------------------------------------------------------------------
// clientInitMiddleware (shared)
// ---------------------------------------------------------------------------

func TestClientInitMiddleware(t *testing.T) {
	t.Run("error when no session", func(t *testing.T) {
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			t.Fatal("next should not be called")
			return nil, nil
		}

		middleware := clientInitMiddleware("http://example.com")
		handler := middleware(next)

		_, err := handler(context.Background(), "test/method", nil)
		require.Error(t, err)
		assert.Equal(t, "no session in context", err.Error())
	})

	t.Run("skips when client already initialized", func(t *testing.T) {
		client, err := createUptimeClient("test-token", "http://example.com")
		require.NoError(t, err)

		session := &app.Session{
			Token:  "test-token",
			Client: client,
		}
		ctx := app.ContextWithSession(context.Background(), session)

		called := false
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			called = true
			return nil, nil
		}

		middleware := clientInitMiddleware("http://example.com")
		handler := middleware(next)

		_, err = handler(ctx, "test/method", nil)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("creates client from token", func(t *testing.T) {
		session := &app.Session{Token: "valid-token"}
		ctx := app.ContextWithSession(context.Background(), session)

		called := false
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			called = true
			s := app.SessionFromContext(ctx)
			assert.NotNil(t, s.Client)
			return nil, nil
		}

		middleware := clientInitMiddleware("http://example.com")
		handler := middleware(next)

		_, err := handler(ctx, "test/method", nil)
		require.NoError(t, err)
		assert.True(t, called)
	})
}

// ---------------------------------------------------------------------------
// tokenHolder (stdio mode)
// ---------------------------------------------------------------------------

func TestTokenHolder(t *testing.T) {
	t.Run("access token from holder", func(t *testing.T) {
		holder := newTokenHolder(&oauth2.Token{AccessToken: "abc123"})
		assert.Equal(t, "abc123", holder.AccessToken())
	})

	t.Run("empty when nil token", func(t *testing.T) {
		holder := newTokenHolder(nil)
		assert.Equal(t, "", holder.AccessToken())
	})

	t.Run("update replaces token", func(t *testing.T) {
		holder := newTokenHolder(&oauth2.Token{AccessToken: "old"})
		holder.Update(&oauth2.Token{AccessToken: "new"})
		assert.Equal(t, "new", holder.AccessToken())
	})
}

// ---------------------------------------------------------------------------
// stdioOAuthFlow
// ---------------------------------------------------------------------------

func TestStdioOAuthFlow(t *testing.T) {
	// Mock authorization server
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/o/authorize/":
			redirectURI := r.URL.Query().Get("redirect_uri")
			state := r.URL.Query().Get("state")
			http.Redirect(w, r, redirectURI+"?code=test-auth-code&state="+state, http.StatusFound)

		case "/o/token/":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer authServer.Close()

	origOpenBrowser := openBrowserFunc
	openBrowserFunc = func(url string) error {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusFound {
			loc := resp.Header.Get("Location")
			resp2, err := client.Get(loc)
			if err != nil {
				return err
			}
			resp2.Body.Close()
		}
		return nil
	}
	defer func() { openBrowserFunc = origOpenBrowser }()

	cfg := stdioOAuthConfig{
		Issuer:   authServer.URL,
		ClientID: "test-client-id",
		Scopes:   []string{"api/v1"},
	}

	token, err := stdioOAuthFlow(context.Background(), noopLogger(), cfg)
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.Equal(t, "test-access-token", token.AccessToken)
	assert.Equal(t, "test-refresh-token", token.RefreshToken)
}

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}
