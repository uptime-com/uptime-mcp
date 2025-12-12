package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		authHeader   string
		queryKey     string
		wantKey      string
		wantAuthGone bool
	}{
		{
			name:         "bearer header",
			authHeader:   "Bearer test-api-key",
			wantKey:      "test-api-key",
			wantAuthGone: true,
		},
		{
			name:     "query parameter",
			queryKey: "query-api-key",
			wantKey:  "query-api-key",
		},
		{
			name:         "header takes precedence",
			authHeader:   "Bearer header-key",
			queryKey:     "query-key",
			wantKey:      "header-key",
			wantAuthGone: true,
		},
		{
			name:    "empty when neither present",
			wantKey: "",
		},
		{
			name:         "non-bearer auth header ignored",
			authHeader:   "Basic dXNlcjpwYXNz",
			wantKey:      "",
			wantAuthGone: true,
		},
		{
			name:         "bearer prefix only",
			authHeader:   "Bearer ",
			wantKey:      "",
			wantAuthGone: true,
		},
		{
			name:         "bearer with extra spaces preserved",
			authHeader:   "Bearer  key-with-space",
			wantKey:      " key-with-space",
			wantAuthGone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedKey string
			var capturedAuthHeader string

			handler := extractAPIKey(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedKey = r.Header.Get(headerUptimeAPIKey)
				capturedAuthHeader = r.Header.Get("Authorization")
			}))

			url := "/"
			if tt.queryKey != "" {
				url = "/?key=" + tt.queryKey
			}
			req := httptest.NewRequest(http.MethodPost, url, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			handler.ServeHTTP(httptest.NewRecorder(), req)

			if capturedKey != tt.wantKey {
				t.Errorf("got key %q, want %q", capturedKey, tt.wantKey)
			}
			if tt.wantAuthGone && capturedAuthHeader != "" {
				t.Errorf("Authorization header should be removed, got %q", capturedAuthHeader)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    string
	}{
		{
			name:       "valid key - 200",
			statusCode: http.StatusOK,
			wantErr:    "",
		},
		{
			name:       "valid key - 405",
			statusCode: http.StatusMethodNotAllowed,
			wantErr:    "",
		},
		{
			name:       "invalid key - 401",
			statusCode: http.StatusUnauthorized,
			wantErr:    "invalid API key",
		},
		{
			name:       "server error - 500",
			statusCode: http.StatusInternalServerError,
			wantErr:    "unexpected status: 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodHead {
					t.Errorf("expected HEAD request, got %s", r.Method)
				}
				if auth := r.Header.Get("Authorization"); auth != "Token test-key" {
					t.Errorf("expected Authorization 'Token test-key', got %q", auth)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer ts.Close()

			err := validateAPIKey(context.Background(), "test-key", ts.URL)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("got error %q, want %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestStdioKeyMiddleware(t *testing.T) {
	t.Run("creates session when none exists", func(t *testing.T) {
		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		middleware := stdioKeyMiddleware("test-api-key")
		handler := middleware(next)

		_, err := handler(context.Background(), "test/method", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		session := app.SessionFromContext(capturedCtx)
		if session == nil {
			t.Fatal("expected session in context")
		}
		if session.APIKey != "test-api-key" {
			t.Errorf("got APIKey %q, want %q", session.APIKey, "test-api-key")
		}
	})

	t.Run("preserves existing session", func(t *testing.T) {
		existingSession := &app.Session{APIKey: "existing-key"}
		ctx := app.ContextWithSession(context.Background(), existingSession)

		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		middleware := stdioKeyMiddleware("new-key")
		handler := middleware(next)

		_, err := handler(ctx, "test/method", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		session := app.SessionFromContext(capturedCtx)
		if session.APIKey != "existing-key" {
			t.Errorf("session should be preserved, got APIKey %q", session.APIKey)
		}
	})
}

func TestHttpKeyMiddleware(t *testing.T) {
	t.Run("extracts API key from header", func(t *testing.T) {
		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		middleware := httpKeyMiddleware()
		handler := middleware(next)

		header := make(http.Header)
		header.Set(headerUptimeAPIKey, "test-api-key")
		req := &mcp.ServerRequest[mcp.Params]{
			Extra: &mcp.RequestExtra{
				Header: header,
			},
		}

		_, err := handler(context.Background(), "test/method", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		session := app.SessionFromContext(capturedCtx)
		if session == nil {
			t.Fatal("expected session in context")
		}
		if session.APIKey != "test-api-key" {
			t.Errorf("got APIKey %q, want %q", session.APIKey, "test-api-key")
		}
	})

	t.Run("returns error when header missing", func(t *testing.T) {
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			t.Fatal("next should not be called")
			return nil, nil
		}

		middleware := httpKeyMiddleware()
		handler := middleware(next)

		req := &mcp.ServerRequest[mcp.Params]{
			Extra: &mcp.RequestExtra{
				Header: http.Header{},
			},
		}

		_, err := handler(context.Background(), "test/method", req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "authorization required" {
			t.Errorf("got error %q, want %q", err.Error(), "authorization required")
		}
	})

	t.Run("preserves existing session", func(t *testing.T) {
		existingSession := &app.Session{APIKey: "existing-key"}
		ctx := app.ContextWithSession(context.Background(), existingSession)

		var capturedCtx context.Context
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			capturedCtx = ctx
			return nil, nil
		}

		middleware := httpKeyMiddleware()
		handler := middleware(next)

		header := make(http.Header)
		header.Set(headerUptimeAPIKey, "new-key")
		req := &mcp.ServerRequest[mcp.Params]{
			Extra: &mcp.RequestExtra{
				Header: header,
			},
		}

		_, err := handler(ctx, "test/method", req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		session := app.SessionFromContext(capturedCtx)
		if session.APIKey != "existing-key" {
			t.Errorf("session should be preserved, got APIKey %q", session.APIKey)
		}
	})
}

func TestClientInitMiddleware(t *testing.T) {
	t.Run("error when no session", func(t *testing.T) {
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			t.Fatal("next should not be called")
			return nil, nil
		}

		middleware := clientInitMiddleware("http://example.com")
		handler := middleware(next)

		_, err := handler(context.Background(), "test/method", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "no session in context" {
			t.Errorf("got error %q, want %q", err.Error(), "no session in context")
		}
	})

	t.Run("skips when client already initialized", func(t *testing.T) {
		// Create a real client to test the "already initialized" path
		client, err := createUptimeClient("test-key", "http://example.com")
		if err != nil {
			t.Fatalf("failed to create test client: %v", err)
		}
		session := &app.Session{
			APIKey: "test-key",
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("next handler should be called")
		}
	})

	t.Run("validates and creates client", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		session := &app.Session{APIKey: "valid-key"}
		ctx := app.ContextWithSession(context.Background(), session)

		called := false
		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			called = true
			// Verify client was set
			s := app.SessionFromContext(ctx)
			if s.Client == nil {
				t.Error("client should be initialized")
			}
			return nil, nil
		}

		middleware := clientInitMiddleware(ts.URL)
		handler := middleware(next)

		_, err := handler(ctx, "test/method", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Error("next handler should be called")
		}
	})

	t.Run("returns validation error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer ts.Close()

		session := &app.Session{APIKey: "invalid-key"}
		ctx := app.ContextWithSession(context.Background(), session)

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			t.Fatal("next should not be called on validation error")
			return nil, nil
		}

		middleware := clientInitMiddleware(ts.URL)
		handler := middleware(next)

		_, err := handler(ctx, "test/method", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "invalid API key" {
			t.Errorf("got error %q, want %q", err.Error(), "invalid API key")
		}
	})
}
