package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_check",
		Description: "Get details of a specific monitoring check by ID",
	}, h.HandleGetCheck)
}

type getCheckInput struct {
	ID int `json:"id"`
}

func (c *checksHandler) HandleGetCheck(ctx context.Context, _ *mcp.CallToolRequest, in getCheckInput) (*mcp.CallToolResult, any, error) {
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	check, _, err := c.service.Get(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get check: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Check #%d: %s\n", check.PK, check.Name)
	fmt.Fprintf(&sb, "Type: %s\n", check.CheckType)
	fmt.Fprintf(&sb, "Address: %s\n", check.Address)
	if check.Port > 0 {
		fmt.Fprintf(&sb, "Port: %d\n", check.Port)
	}
	fmt.Fprintf(&sb, "Interval: %d seconds\n", check.Interval)
	fmt.Fprintf(&sb, "Sensitivity: %d\n", check.Sensitivity)
	if len(check.Locations) > 0 {
		fmt.Fprintf(&sb, "Locations: %s\n", strings.Join(check.Locations, ", "))
	}
	if len(check.Tags) > 0 {
		fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(check.Tags, ", "))
	}
	if check.Notes != "" {
		fmt.Fprintf(&sb, "Notes: %s\n", check.Notes)
	}

	return textResult(sb.String()), nil, nil
}
