package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateDNSCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_dns_check",
		Description: "Update an existing DNS monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateDNSCheck)
}

type updateDNSCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"domain name to resolve"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes"`
	Locations     []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	DNSServer     string   `json:"dns_server,omitempty" jsonschema:"DNS server to query"`
	DNSRecordType string   `json:"dns_record_type,omitempty" jsonschema:"DNS record type to query, e.g. A, AAAA, CNAME, MX, NS, TXT"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"expected value in DNS response"`
}

func (c *checksHandler) HandleUpdateDNSCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateDNSCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckDNS{
		Name:          in.Name,
		Address:       in.Address,
		Interval:      in.Interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		DNSServer:     in.DNSServer,
		DNSRecordType: in.DNSRecordType,
		ExpectString:  in.ExpectString,
	}

	updated, err := client.Checks().UpdateDNS(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update DNS check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated DNS check #%d: %s", updated.PK, updated.Name)), nil, nil
}
