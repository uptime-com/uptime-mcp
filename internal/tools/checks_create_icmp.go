package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

var CreateICMPCheckToolModule = fx.Module("tool.create_icmp_check",
	fx.Invoke(func(srv *mcp.Server, c *checksHandler) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "create_icmp_check",
			Description: "Create a new ICMP/Ping monitoring check",
		}, c.HandleCreateICMPCheck)
	}),
)

type createICMPCheckInput struct {
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Interval    int      `json:"interval,omitempty"`
	Locations   []string `json:"locations,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Sensitivity int      `json:"sensitivity,omitempty"`
	Notes       string   `json:"notes,omitempty"`
}

func (c *checksHandler) HandleCreateICMPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createICMPCheckInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	check := &api.Check{
		CheckType:   "ICMP",
		Name:        in.Name,
		Address:     in.Address,
		Interval:    in.Interval,
		Locations:   in.Locations,
		Tags:        in.Tags,
		Sensitivity: in.Sensitivity,
		Notes:       in.Notes,
	}

	created, _, err := c.service.Create(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create ICMP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created ICMP check #%d: %s", created.PK, created.Name)), nil, nil
}
