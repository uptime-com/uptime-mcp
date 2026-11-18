package handle

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uptime-com/uptime-mcp/internal/server"
)

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.Len(t, result.Content, 1)
	return result.Content[0].(*mcp.TextContent).Text
}

// linkWithPrefix returns the first whitespace-separated field of text that
// starts with prefix, parsed as a URL.
func linkWithPrefix(t *testing.T, text, prefix string) *url.URL {
	t.Helper()
	for _, field := range strings.Fields(text) {
		if strings.HasPrefix(field, prefix) {
			u, err := url.Parse(field)
			require.NoError(t, err)
			return u
		}
	}
	t.Fatalf("no %s link in %q", prefix, text)
	return nil
}

func TestHandleReportIssue(t *testing.T) {
	h := &reportHandler{info: server.Info{Version: "0.18.0", Commit: "abc1234"}}

	t.Run("bug report becomes a prefilled issue link", func(t *testing.T) {
		result, _, err := h.HandleReportIssue(t.Context(), nil, reportIssueInput{
			Kind:  "bug",
			Title: "get_alert prints an empty location",
			Body:  "get_alert 8720057 printed `Location: ` with nothing after it.",
		})
		require.NoError(t, err)
		text := resultText(t, result)

		q := linkWithPrefix(t, text, repoURL+"/issues/new").Query()
		assert.Equal(t, "Bug: get_alert prints an empty location", q.Get("title"))
		assert.Contains(t, q.Get("body"), "get_alert 8720057 printed `Location: ` with nothing after it.")
		assert.Contains(t, q.Get("body"), "uptime-mcp 0.18.0 (abc1234)")
		assert.Empty(t, q.Get("labels"))
		assert.Contains(t, text, "Nothing has been filed")
	})

	t.Run("feature request carries its own prefix", func(t *testing.T) {
		result, _, err := h.HandleReportIssue(t.Context(), nil, reportIssueInput{
			Kind:  "feature",
			Title: "Return a check's response time series",
			Body:  "Needed hourly response times to compare two regions.",
		})
		require.NoError(t, err)
		q := linkWithPrefix(t, resultText(t, result), repoURL+"/issues/new").Query()
		assert.Equal(t, "Feature request: Return a check's response time series", q.Get("title"))
	})

	t.Run("existing issue gets a comment draft instead of a new issue", func(t *testing.T) {
		result, _, err := h.HandleReportIssue(t.Context(), nil, reportIssueInput{
			Kind:          "feature",
			Title:         "Return a check's response time series",
			Body:          "Also needed: per-location percentiles.",
			ExistingIssue: 66,
		})
		require.NoError(t, err)
		text := resultText(t, result)
		assert.Contains(t, text, repoURL+"/issues/66")
		assert.NotContains(t, text, "/issues/new")
		assert.Contains(t, text, "Also needed: per-location percentiles.")
		assert.Contains(t, text, "uptime-mcp 0.18.0 (abc1234)")
	})

	t.Run("long body is cut to fit a URL", func(t *testing.T) {
		result, _, err := h.HandleReportIssue(t.Context(), nil, reportIssueInput{
			Kind:  "bug",
			Title: "Huge output",
			Body:  strings.Repeat("лог ", 5000),
		})
		require.NoError(t, err)
		u := linkWithPrefix(t, resultText(t, result), repoURL+"/issues/new")
		assert.LessOrEqual(t, len(u.String()), maxIssueURLLen)
		body := u.Query().Get("body")
		assert.Contains(t, body, "truncated")
		assert.Contains(t, body, "uptime-mcp 0.18.0 (abc1234)")
	})

	t.Run("rejects an unknown kind", func(t *testing.T) {
		_, _, err := h.HandleReportIssue(t.Context(), nil, reportIssueInput{Kind: "question", Title: "x", Body: "y"})
		assert.ErrorContains(t, err, "kind")
	})

	t.Run("requires a title", func(t *testing.T) {
		_, _, err := h.HandleReportIssue(t.Context(), nil, reportIssueInput{Kind: "bug", Body: "y"})
		assert.ErrorContains(t, err, "title")
	})
}

func TestHandleSearchIssues(t *testing.T) {
	t.Run("lists matching issues and steers toward commenting", func(t *testing.T) {
		var gotQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/search/issues", r.URL.Path)
			gotQuery = r.URL.Query().Get("q")
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"total_count":1,"items":[{"number":66,"title":"Expose a check's response time series","state":"open","comments":2,"html_url":"https://github.com/uptime-com/uptime-mcp/issues/66"}]}`)
		}))
		t.Cleanup(srv.Close)

		h := &reportHandler{githubAPI: srv.URL, http: srv.Client()}
		result, _, err := h.HandleSearchIssues(t.Context(), nil, searchIssuesInput{Query: "response time"})
		require.NoError(t, err)
		text := resultText(t, result)

		assert.Contains(t, gotQuery, "repo:uptime-com/uptime-mcp")
		assert.Contains(t, gotQuery, "is:issue")
		assert.Contains(t, gotQuery, "response time")
		assert.Contains(t, text, "#66 [open, 2 comments] Expose a check's response time series")
		assert.Contains(t, text, "https://github.com/uptime-com/uptime-mcp/issues/66")
		assert.Contains(t, text, "existing_issue")
	})

	t.Run("no matches points to a new issue", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{"total_count":0,"items":[]}`)
		}))
		t.Cleanup(srv.Close)

		h := &reportHandler{githubAPI: srv.URL, http: srv.Client()}
		result, _, err := h.HandleSearchIssues(t.Context(), nil, searchIssuesInput{Query: "nothing like this"})
		require.NoError(t, err)
		assert.Contains(t, resultText(t, result), "No existing issue matches")
	})

	t.Run("failed search falls back to a link the user can open", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `{"message":"API rate limit exceeded"}`)
		}))
		t.Cleanup(srv.Close)

		h := &reportHandler{githubAPI: srv.URL, http: srv.Client()}
		result, _, err := h.HandleSearchIssues(t.Context(), nil, searchIssuesInput{Query: "alerts"})
		require.NoError(t, err)
		text := resultText(t, result)
		assert.Contains(t, text, "403")
		q := linkWithPrefix(t, text, repoURL+"/issues?").Query()
		assert.Contains(t, q.Get("q"), "alerts")
	})

	t.Run("requires a query", func(t *testing.T) {
		h := &reportHandler{}
		_, _, err := h.HandleSearchIssues(t.Context(), nil, searchIssuesInput{})
		assert.ErrorContains(t, err, "query")
	})
}
