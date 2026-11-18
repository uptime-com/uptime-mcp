package landing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const browserAccept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"

func TestHandler(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		accept  []string
		landing bool
	}{
		{"browser", http.MethodGet, "/", []string{browserAccept}, true},
		{"browser HEAD", http.MethodHead, "/", []string{browserAccept}, true},
		{"Accept split over header lines", http.MethodGet, "/", []string{"application/json", "text/html"}, true},
		{"MCP POST", http.MethodPost, "/", []string{"application/json, text/event-stream"}, false},
		{"MCP POST asking for HTML", http.MethodPost, "/", []string{browserAccept}, false},
		{"MCP SSE stream", http.MethodGet, "/", []string{"text/event-stream"}, false},
		{"wildcard only", http.MethodGet, "/", []string{"*/*"}, false},
		{"no Accept", http.MethodGet, "/", nil, false},
		{"HTML refused", http.MethodGet, "/", []string{"text/html;q=0, */*"}, false},
		{"other path", http.MethodGet, "/other", []string{browserAccept}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})
			req := httptest.NewRequest(tt.method, tt.path, nil)
			for _, v := range tt.accept {
				req.Header.Add("Accept", v)
			}
			rec := httptest.NewRecorder()

			Handler(next).ServeHTTP(rec, req)

			if !tt.landing {
				assert.Equal(t, http.StatusTeapot, rec.Code)
				return
			}
			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
			assert.Equal(t, "Accept", rec.Header().Get("Vary"))
			assert.Contains(t, rec.Body.String(), "Ask your AI assistant about Uptime.com")
		})
	}
}
