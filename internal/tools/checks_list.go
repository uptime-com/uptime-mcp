package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

// ListChecksToolModule registers the list_checks tool.
var ListChecksToolModule = fx.Module("tool.list_checks",
	fx.Invoke(func(srv *mcp.Server, c *checksHandler) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "list_checks",
			Description: "List monitoring checks with optional filtering by search term, tag, or check type",
		}, c.HandleListChecks)
	}),
)

// listChecksInput defines parameters for listing checksHandler.
type listChecksInput struct {
	Search   string `json:"search,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Type     string `json:"type,omitempty"`
	IsPaused bool   `json:"is_paused,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

func (c *checksHandler) HandleListChecks(ctx context.Context, _ *mcp.CallToolRequest, in listChecksInput) (*mcp.CallToolResult, any, error) {
	opts := &api.CheckListOptions{
		Search:                in.Search,
		MonitoringServiceType: in.Type,
		IsPaused:              in.IsPaused,
		Page:                  in.Page,
		PageSize:              in.PageSize,
	}
	if in.Tag != "" {
		opts.Tag = []string{in.Tag}
	}

	checks, _, err := c.service.List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list checks: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d checks:\n\n", len(checks))
	for _, ch := range checks {
		fmt.Fprintf(&sb, "- [%d] %s (%s) - %s\n", ch.PK, ch.Name, ch.CheckType, ch.Address)
	}

	return textResult(sb.String()), nil, nil
}
