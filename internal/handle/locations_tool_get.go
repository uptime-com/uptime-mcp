package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetLocationTool(srv *mcp.Server, h *locationsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_location",
		Description: "Get detailed information about a probe server location including IP addresses",
	}, h.HandleGetLocation)
}

type getLocationInput struct {
	Location string `json:"location" jsonschema:"location identifier as returned by list_locations"`
}

func (h *locationsHandler) HandleGetLocation(ctx context.Context, _ *mcp.CallToolRequest, in getLocationInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Location == "" {
		return nil, nil, fmt.Errorf("location is required")
	}

	if excludedLocations[in.Location] {
		return nil, nil, fmt.Errorf("pseudo-location not supported: %s", in.Location)
	}

	var sb strings.Builder
	if err := h.loadLocation(ctx, client, in.Location, &sb); err != nil {
		return nil, nil, err
	}

	return textResult(sb.String()), nil, nil
}
