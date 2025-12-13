package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

var CreateDNSCheckToolModule = fx.Module("tool.create_dns_check",
	fx.Invoke(func(srv *mcp.Server, checks *checksHandler) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "create_dns_check",
			Description: "Create a new DNS monitoring check",
		}, checks.HandleCreateDNSCheck)
	}),
)

type createDNSCheckInput struct {
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	Interval      int      `json:"interval,omitempty"`
	Locations     []string `json:"locations,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Sensitivity   int      `json:"sensitivity,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	DNSServer     string   `json:"dns_server,omitempty"`
	DNSRecordType string   `json:"dns_record_type,omitempty"`
	ExpectString  string   `json:"expect_string,omitempty"`
}

func (c *checksHandler) HandleCreateDNSCheck(ctx context.Context, _ *mcp.CallToolRequest, in createDNSCheckInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	check := &api.Check{
		CheckType:     "DNS",
		Name:          in.Name,
		Address:       in.Address,
		Interval:      in.Interval,
		Locations:     in.Locations,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		DNSServer:     in.DNSServer,
		DNSRecordType: in.DNSRecordType,
		ExpectString:  in.ExpectString,
	}

	created, _, err := c.service.Create(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DNS check: %w", err)
	}

	return textResult(fmt.Sprintf("Created DNS check #%d: %s", created.PK, created.Name)), nil, nil
}
