package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerListLocationsTool(srv *mcp.Server, h *locationsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_locations",
		Description: "List all available probe server locations for monitoring checks",
	}, h.handleListLocations)
}

type listLocationsInput struct {
	Search string `json:"search,omitempty"`
}

func (h *locationsHandler) handleListLocations(ctx context.Context, _ *mcp.CallToolRequest, in listLocationsInput) (*mcp.CallToolResult, any, error) {
	servers, err := h.service.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list locations: %w", err)
	}

	// Filter by search if provided
	if in.Search != "" {
		search := strings.ToLower(in.Search)
		filtered := make([]struct{ Location, ProbeName string }, 0)
		for _, s := range servers {
			if strings.Contains(strings.ToLower(s.Location), search) ||
				strings.Contains(strings.ToLower(s.ProbeName), search) {
				filtered = append(filtered, struct{ Location, ProbeName string }{s.Location, s.ProbeName})
			}
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Found %d locations matching '%s':\n\n", len(filtered), in.Search)
		for _, s := range filtered {
			fmt.Fprintf(&sb, "- %s (%s)\n", s.Location, s.ProbeName)
		}
		return textResult(sb.String()), nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d probe server locations:\n\n", len(servers))
	for _, s := range servers {
		fmt.Fprintf(&sb, "- %s (%s)\n", s.Location, s.ProbeName)
	}

	return textResult(sb.String()), nil, nil
}
