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

// excludedLocations are pseudo-locations that cannot be used for check creation.
var excludedLocations = map[string]bool{
	"AUTO": true,
	"TEST": true,
}

func (h *locationsHandler) handleListLocations(ctx context.Context, _ *mcp.CallToolRequest, in listLocationsInput) (*mcp.CallToolResult, any, error) {
	servers, err := h.service.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list locations: %w", err)
	}

	// Filter out pseudo-locations and apply search
	type loc struct{ Location, ProbeName string }
	filtered := make([]loc, 0, len(servers))
	search := strings.ToLower(in.Search)

	for _, s := range servers {
		if excludedLocations[s.Location] {
			continue
		}
		if search != "" {
			if !strings.Contains(strings.ToLower(s.Location), search) &&
				!strings.Contains(strings.ToLower(s.ProbeName), search) {
				continue
			}
		}
		filtered = append(filtered, loc{s.Location, s.ProbeName})
	}

	var sb strings.Builder
	if in.Search != "" {
		fmt.Fprintf(&sb, "Found %d locations matching '%s':\n\n", len(filtered), in.Search)
	} else {
		fmt.Fprintf(&sb, "Found %d probe server locations:\n\n", len(filtered))
	}
	for _, s := range filtered {
		fmt.Fprintf(&sb, "- %s (%s)\n", s.Location, s.ProbeName)
	}

	return textResult(sb.String()), nil, nil
}
