package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListChecksTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_checks",
		Description: "List monitoring checks with optional filtering by search term, tag, or check type",
	}, h.HandleListChecks)
}

// listChecksInput defines parameters for listing checksHandler.
type listChecksInput struct {
	Search   string `json:"search,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Type     string `json:"type,omitempty"`
	IsPaused bool   `json:"is_paused,omitempty"`
	Page     int64  `json:"page,omitempty"`
	PageSize int64  `json:"page_size,omitempty"`
}

func (c *checksHandler) HandleListChecks(ctx context.Context, _ *mcp.CallToolRequest, in listChecksInput) (*mcp.CallToolResult, any, error) {
	opts := upapi.CheckListOptions{
		Search:                in.Search,
		MonitoringServiceType: in.Type,
		IsPaused:              in.IsPaused,
		Page:                  in.Page,
		PageSize:              in.PageSize,
	}
	if in.Tag != "" {
		opts.Tag = []string{in.Tag}
	}

	checks, err := c.service.List(ctx, opts)
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
