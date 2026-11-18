package handle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerSearchIssuesTool(srv *mcp.Server, h *reportHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search_issues",
		Description: "Search this MCP server's GitHub issues for an existing bug report or feature request. Call this before report_issue: commenting on a matching issue is preferred to opening a new one.",
	}, h.HandleSearchIssues)
}

type searchIssuesInput struct {
	Query string `json:"query" jsonschema:"keywords describing the problem or the missing capability, e.g. a tool name and the symptom"`
}

type githubIssueSearch struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		Number   int    `json:"number"`
		Title    string `json:"title"`
		State    string `json:"state"`
		Comments int    `json:"comments"`
		HTMLURL  string `json:"html_url"`
	} `json:"items"`
}

func (h *reportHandler) HandleSearchIssues(ctx context.Context, _ *mcp.CallToolRequest, in searchIssuesInput) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(in.Query) == "" {
		return nil, nil, errors.New("query is required")
	}

	found, err := h.searchIssues(ctx, in.Query)
	if err != nil {
		return textResult(fmt.Sprintf(
			"Searching GitHub failed: %v\nThe user can search existing issues at %s\n",
			err, issueSearchURL(in.Query))), nil, nil
	}

	var sb strings.Builder
	if len(found.Items) == 0 {
		fmt.Fprintf(&sb, "No existing issue matches %q. Use report_issue to draft a new one.\n", in.Query)
		return textResult(sb.String()), nil, nil
	}

	fmt.Fprintf(&sb, "%d existing issue(s) match %q; showing %d:\n\n", found.TotalCount, in.Query, len(found.Items))
	for _, it := range found.Items {
		fmt.Fprintf(&sb, "- #%d [%s, %d comments] %s\n  %s\n", it.Number, it.State, it.Comments, it.Title, it.HTMLURL)
	}
	sb.WriteString("\nIf one of these describes the same problem or request, call report_issue with existing_issue set to its number, " +
		"so the user adds a comment there instead of opening a duplicate. A closed issue that still reproduces is worth a comment too.\n")
	return textResult(sb.String()), nil, nil
}

// searchIssues queries GitHub's issue search, unauthenticated, for issues in
// this repository; GitHub allows a few such requests per minute per client IP.
func (h *reportHandler) searchIssues(ctx context.Context, query string) (*githubIssueSearch, error) {
	q := url.Values{
		"q":        {fmt.Sprintf("repo:%s is:issue %s", repoSlug, query)},
		"per_page": {"10"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.githubAPI+"/search/issues?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	rs, err := h.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rs.Body.Close() }()
	if rs.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub answered %s", rs.Status)
	}

	var found githubIssueSearch
	if err := json.NewDecoder(rs.Body).Decode(&found); err != nil {
		return nil, fmt.Errorf("decoding GitHub's answer: %w", err)
	}
	return &found, nil
}

// issueSearchURL is the repository's issue search page for query.
func issueSearchURL(query string) string {
	return repoURL + "/issues?" + url.Values{"q": {"is:issue " + query}}.Encode()
}
