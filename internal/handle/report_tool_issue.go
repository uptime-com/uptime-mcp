package handle

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxIssueURLLen keeps a prefilled new-issue link under the length GitHub
// rejects with 414 URI Too Long.
const maxIssueURLLen = 8000

const truncationNote = "\n\n[truncated: the full report was longer than a link can carry]"

var issueTitlePrefix = map[string]string{
	"bug":     "Bug: ",
	"feature": "Feature request: ",
}

func registerReportIssueTool(srv *mcp.Server, h *reportHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "report_issue",
		Description: "Draft a bug report or feature request for this MCP server's GitHub issue tracker. " +
			"Returns a link for the user to open, review and submit; nothing is filed by this call. " +
			"Run search_issues first and pass existing_issue when one matches, to draft a comment on it instead of a new issue.",
	}, h.HandleReportIssue)
}

type reportIssueInput struct {
	Kind          string `json:"kind" jsonschema:"bug or feature"`
	Title         string `json:"title" jsonschema:"one-line summary naming the tool or capability, anonymized like the body"`
	Body          string `json:"body" jsonschema:"for a bug: the tool, the input, what came back and what was expected; for a feature: the task and what was missing. Issues are public: replace tokens, keys, emails, phone numbers, the user's domains, URLs and IPs, and account, contact and check names with placeholders such as <domain> or <check name>"`
	ExistingIssue int    `json:"existing_issue,omitempty" jsonschema:"number of a matching issue found by search_issues; drafts a comment on it instead of a new issue"`
}

func (h *reportHandler) HandleReportIssue(_ context.Context, _ *mcp.CallToolRequest, in reportIssueInput) (*mcp.CallToolResult, any, error) {
	prefix, ok := issueTitlePrefix[in.Kind]
	if !ok {
		return nil, nil, fmt.Errorf("kind must be bug or feature, got %q", in.Kind)
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, nil, errors.New("title is required")
	}

	footer := fmt.Sprintf("\n\n---\nReported through report_issue, uptime-mcp %s (%s).", h.info.Version, h.info.Commit)

	if in.ExistingIssue > 0 {
		issue := fmt.Sprintf("%s/issues/%d", repoURL, in.ExistingIssue)
		return textResult(fmt.Sprintf(
			"Give the user this issue and the comment below; they post it there. Nothing has been filed.\n\n%s\n\n%s%s\n",
			issue, in.Body, footer)), nil, nil
	}

	link := newIssueLink(prefix+in.Title, in.Body, footer)
	return textResult(fmt.Sprintf(
		"Give the user this link; opening it lets them review and submit the issue. Nothing has been filed.\n\n%s\n", link)), nil, nil
}

// newIssueLink builds a prefilled new-issue URL no longer than maxIssueURLLen,
// cutting body as needed while keeping footer whole.
func newIssueLink(title, body, footer string) string {
	build := func(b string) string {
		return repoURL + "/issues/new?" + url.Values{"title": {title}, "body": {b + footer}}.Encode()
	}
	link := build(body)
	for len(link) > maxIssueURLLen && body != "" {
		cut := len(body) * 9 / 10
		for cut > 0 && !utf8.RuneStart(body[cut]) {
			cut--
		}
		body = body[:cut]
		link = build(body + truncationNote)
	}
	return link
}
