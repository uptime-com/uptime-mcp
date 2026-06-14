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
		Description: "List available probe server location identifiers. Pass these identifiers directly in the locations field when creating checks.",
	}, h.handleListLocations)
}

type listLocationsInput struct {
	Search string `json:"search,omitempty" jsonschema:"filter locations by name"`
}

// excludedLocations are pseudo-locations that cannot be used for check creation.
var excludedLocations = map[string]bool{
	"AUTO": true,
	"TEST": true,
}

func (h *locationsHandler) handleListLocations(ctx context.Context, _ *mcp.CallToolRequest, in listLocationsInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	result, err := client.Checks().ListLocations(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list locations: %w", err)
	}

	filtered := make([]string, 0, len(result.Items))
	search := strings.ToLower(in.Search)

	for _, loc := range result.Items {
		if excludedLocations[loc] {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(loc), search) {
			continue
		}
		filtered = append(filtered, loc)
	}

	var sb strings.Builder
	if in.Search != "" {
		fmt.Fprintf(&sb, "Found %d locations matching '%s'.\n", len(filtered), in.Search)
	} else {
		fmt.Fprintf(&sb, "Found %d locations.\n", len(filtered))
	}
	fmt.Fprintf(&sb, "Use these identifiers in the locations field when creating checks:\n\n")
	for _, loc := range filtered {
		fmt.Fprintf(&sb, "- %s\n", loc)
	}

	return textResult(sb.String()), nil, nil
}
