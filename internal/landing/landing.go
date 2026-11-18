// Package landing serves the page a browser gets on opening the server's URL.
package landing

import (
	_ "embed"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

//go:embed index.html
var page []byte

// Handler serves the landing page for a GET or HEAD of "/" whose Accept header
// lists text/html, and passes every other request to next. MCP clients never
// ask for text/html, so they always reach next.
func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || (r.Method != http.MethodGet && r.Method != http.MethodHead) ||
			!acceptsHTML(r.Header.Values("Accept")) {
			next.ServeHTTP(w, r)
			return
		}
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", "public, max-age=300")
		h.Set("Vary", "Accept")
		_, _ = w.Write(page)
	})
}

// acceptsHTML reports whether the Accept header values name text/html with a
// non-zero quality. A wildcard does not count: non-browser clients send */*.
func acceptsHTML(values []string) bool {
	for _, v := range values {
		for part := range strings.SplitSeq(v, ",") {
			mediaType, params, err := mime.ParseMediaType(part)
			if err != nil || mediaType != "text/html" {
				continue
			}
			q, err := strconv.ParseFloat(params["q"], 64)
			if params["q"] == "" || (err == nil && q > 0) {
				return true
			}
		}
	}
	return false
}
