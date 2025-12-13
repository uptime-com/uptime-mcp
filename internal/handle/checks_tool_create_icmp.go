package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateICMPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_icmp_check",
		Description: "Create a new ICMP/Ping monitoring check",
	}, h.HandleCreateICMPCheck)
}

type createICMPCheckInput struct {
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	Interval      int64    `json:"interval"`
	Locations     []string `json:"locations"`
	ContactGroups []string `json:"contact_groups"`
	Tags          []string `json:"tags,omitempty"`
	Sensitivity   int64    `json:"sensitivity,omitempty"`
	Notes         string   `json:"notes,omitempty"`
}

func (c *checksHandler) HandleCreateICMPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createICMPCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckICMP{
		Name:          in.Name,
		Address:       in.Address,
		Interval:      in.Interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
	}

	created, err := client.Checks().CreateICMP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ICMP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created ICMP check #%d: %s", created.PK, created.Name)), nil, nil
}
