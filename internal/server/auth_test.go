package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

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
// extractCredential
// ---------------------------------------------------------------------------

func TestExtractCredential(t *testing.T) {
	t.Run("from Authorization header, as an OAuth2 access token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		assert.Equal(t, credential{"header-token", app.SchemeBearer}, extractCredential(req))
	})

	t.Run("from query param, as an API key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?token=query-token", nil)
		assert.Equal(t, credential{"query-token", app.SchemeToken}, extractCredential(req))
	})

	t.Run("from env var, as an API key", func(t *testing.T) {
		t.Setenv("UPTIME_BEARER_TOKEN", "env-token")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, credential{"env-token", app.SchemeToken}, extractCredential(req))
	})

	t.Run("header takes precedence over query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?token=query-token", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		assert.Equal(t, credential{"header-token", app.SchemeBearer}, extractCredential(req))
	})

	t.Run("query takes precedence over env", func(t *testing.T) {
		t.Setenv("UPTIME_BEARER_TOKEN", "env-token")
		req := httptest.NewRequest(http.MethodGet, "/?token=query-token", nil)
		assert.Equal(t, credential{"query-token", app.SchemeToken}, extractCredential(req))
	})

	t.Run("empty when no credential", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		assert.Equal(t, credential{}, extractCredential(req))
	})
}

// ---------------------------------------------------------------------------
// bearerPassthrough (HTTP middleware)
// ---------------------------------------------------------------------------

func TestBearerPassthrough(t *testing.T) {
	t.Run("injects credential into context", func(t *testing.T) {
		var captured credential
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			captured, _ = r.Context().Value(passthroughCredentialKey{}).(credential)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		w := httptest.NewRecorder()

		bearerPassthrough(inner).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, credential{"header-token", app.SchemeBearer}, captured)
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
// httpTokenMiddleware (MCP middleware)
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

	t.Run("creates session from passthrough credential", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), passthroughCredentialKey{},
			credential{"pass-token", app.SchemeBearer})

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
		assert.Equal(t, app.SchemeBearer, session.Scheme)
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
		client, err := createUptimeClient("test-token", app.SchemeBearer, "http://example.com")
		require.NoError(t, err)

		session := &app.Session{
			Token:  "test-token",
			Scheme: app.SchemeBearer,
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

// fakeAuthServer is an authorization server that approves every request. It
// publishes RFC 8414 metadata naming the issuer with a trailing slash, as
// uptime.com does, and offers registration only when registration is true.
type fakeAuthServer struct {
	*httptest.Server
	registration bool
	registered   oauthex.ClientRegistrationMetadata
	tokenForm    map[string][]string
}

func newFakeAuthServer(t *testing.T, registration bool) *fakeAuthServer {
	t.Helper()
	as := &fakeAuthServer{registration: registration}
	as.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/oauth-authorization-server":
			meta := map[string]any{
				"issuer":                           as.URL + "/",
				"authorization_endpoint":           as.URL + "/o/authorize/",
				"token_endpoint":                   as.URL + "/o/token/",
				"response_types_supported":         []string{"code"},
				"code_challenge_methods_supported": []string{"S256"},
			}
			if as.registration {
				meta["registration_endpoint"] = as.URL + "/o/register/"
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(meta)

		case "/o/register/":
			if !as.registration {
				http.NotFound(w, r)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&as.registered); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"client_id":                  "registered-client-id",
				"redirect_uris":              as.registered.RedirectURIs,
				"token_endpoint_auth_method": "none",
			})

		case "/o/authorize/":
			redirectURI := r.URL.Query().Get("redirect_uri")
			state := r.URL.Query().Get("state")
			http.Redirect(w, r, redirectURI+"?code=test-auth-code&state="+state, http.StatusFound)

		case "/o/token/":
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			as.tokenForm = r.PostForm
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
	t.Cleanup(as.Close)
	return as
}

// approveInBrowser stands in for the user's browser: it follows the
// authorization server's redirect to the flow's callback.
func approveInBrowser(t *testing.T) {
	t.Helper()
	orig := openBrowserFunc
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
			resp2, err := client.Get(resp.Header.Get("Location"))
			if err != nil {
				return err
			}
			resp2.Body.Close()
		}
		return nil
	}
	t.Cleanup(func() { openBrowserFunc = orig })
}

func TestStdioOAuthFlow(t *testing.T) {
	t.Run("pre-registered client", func(t *testing.T) {
		as := newFakeAuthServer(t, false)
		approveInBrowser(t)

		token, oauthCfg, err := stdioOAuthFlow(context.Background(), noopLogger(), stdioOAuthConfig{
			Issuer:   as.URL,
			ClientID: "test-client-id",
			Scopes:   []string{"api/v1"},
		})
		require.NoError(t, err)
		assert.Equal(t, "test-access-token", token.AccessToken)
		assert.Equal(t, "test-refresh-token", token.RefreshToken)
		assert.Equal(t, "test-client-id", oauthCfg.ClientID)
	})

	t.Run("registers a public client for the callback", func(t *testing.T) {
		as := newFakeAuthServer(t, true)
		approveInBrowser(t)

		token, oauthCfg, err := stdioOAuthFlow(context.Background(), noopLogger(), stdioOAuthConfig{
			Issuer: as.URL,
			Scopes: []string{"api/v1"},
		})
		require.NoError(t, err)
		assert.Equal(t, "test-access-token", token.AccessToken)

		assert.Equal(t, []string{oauthCfg.RedirectURL}, as.registered.RedirectURIs)
		assert.Equal(t, "none", as.registered.TokenEndpointAuthMethod)
		assert.Equal(t, []string{"authorization_code", "refresh_token"}, as.registered.GrantTypes)

		assert.Equal(t, "registered-client-id", oauthCfg.ClientID)
		assert.Equal(t, as.URL+"/o/token/", oauthCfg.Endpoint.TokenURL)
		assert.Equal(t, []string{"registered-client-id"}, as.tokenForm["client_id"])
	})

	t.Run("no registration endpoint", func(t *testing.T) {
		as := newFakeAuthServer(t, false)
		approveInBrowser(t)

		_, _, err := stdioOAuthFlow(context.Background(), noopLogger(), stdioOAuthConfig{
			Issuer: as.URL,
			Scopes: []string{"api/v1"},
		})
		assert.ErrorContains(t, err, "does not offer client registration")
	})
}

func TestDiscoverAuthServerIssuerSlash(t *testing.T) {
	as := newFakeAuthServer(t, true)
	for _, issuer := range []string{as.URL, as.URL + "/"} {
		meta, err := discoverAuthServer(context.Background(), issuer)
		require.NoError(t, err, issuer)
		assert.Equal(t, as.URL+"/o/register/", meta.RegistrationEndpoint)
	}
}

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

// ---------------------------------------------------------------------------
// createUptimeClient auth scheme selection
// ---------------------------------------------------------------------------

// authRecorder is an api/v1 stand-in that accepts exactly one Authorization
// scheme and records every header offered to it.
type authRecorder struct {
	accepts app.AuthScheme
	seen    []string
}

func (a *authRecorder) server(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		a.seen = append(a.seen, got)
		if !strings.HasPrefix(got, string(a.accepts)+" ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestCreateUptimeClientAuthScheme(t *testing.T) {
	// An opaque credential carries no evidence of which scheme it belongs to,
	// so both cases below use one and differ only in where it came from.
	const opaque = "IqQOoxTFmvT8ONWtdA3sc9wbAzc0rE"

	t.Run("sends the scheme the caller named", func(t *testing.T) {
		cases := []struct {
			name   string
			scheme app.AuthScheme
		}{
			{"oauth access token", app.SchemeBearer},
			{"account api key", app.SchemeToken},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				api := &authRecorder{accepts: tc.scheme}

				client, err := createUptimeClient(opaque, tc.scheme, api.server(t).URL)
				require.NoError(t, err)

				_, err = client.AccountUsage().Get(context.Background())
				require.NoError(t, err)

				assert.Equal(t, []string{string(tc.scheme) + " " + opaque}, api.seen,
					"the named scheme should be offered first and alone")
			})
		}
	})

	t.Run("corrects the scheme when api/v1 refuses it, once", func(t *testing.T) {
		api := &authRecorder{accepts: app.SchemeToken}

		client, err := createUptimeClient(opaque, app.SchemeBearer, api.server(t).URL)
		require.NoError(t, err)

		_, err = client.AccountUsage().Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"Bearer " + opaque, "Token " + opaque}, api.seen)

		// The accepted scheme is kept, so the refusal is not paid again.
		_, err = client.AccountUsage().Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"Bearer " + opaque, "Token " + opaque, "Token " + opaque}, api.seen)
	})

	t.Run("offers each scheme at most once for a credential api/v1 refuses", func(t *testing.T) {
		api := &authRecorder{accepts: "Neither"}

		client, err := createUptimeClient(opaque, app.SchemeBearer, api.server(t).URL)
		require.NoError(t, err)

		_, err = client.AccountUsage().Get(context.Background())
		require.Error(t, err)
		assert.Equal(t, []string{"Bearer " + opaque, "Token " + opaque}, api.seen)
	})
}
