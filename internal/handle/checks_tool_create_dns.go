package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateDNSCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_dns_check",
		Description: "Create a new DNS monitoring check",
	}, h.HandleCreateDNSCheck)
}

type createDNSCheckInput struct {
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	Interval      int64    `json:"interval,omitempty"`
	Locations     []string `json:"locations,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Sensitivity   int64    `json:"sensitivity,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	DNSServer     string   `json:"dns_server,omitempty"`
	DNSRecordType string   `json:"dns_record_type,omitempty"`
	ExpectString  string   `json:"expect_string,omitempty"`
}

func (c *checksHandler) HandleCreateDNSCheck(ctx context.Context, _ *mcp.CallToolRequest, in createDNSCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	check := upapi.CheckDNS{
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

	created, err := client.Checks().CreateDNS(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DNS check: %w", err)
	}

	return textResult(fmt.Sprintf("Created DNS check #%d: %s", created.PK, created.Name)), nil, nil
}
