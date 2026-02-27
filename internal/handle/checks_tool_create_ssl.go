package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateSSLCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_ssl_check",
		Description: "Create a new SSL certificate monitoring check. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreateSSLCheck)
}

type createSSLCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Address       string   `json:"address" jsonschema:"hostname to check SSL certificate for, e.g. example.com"`
	Locations     []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"port number, defaults to 443"`
	Protocol      string   `json:"protocol,omitempty" jsonschema:"protocol to use, e.g. HTTPS, IMAPS, POP3S, SMTPS"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"days before certificate expiry to trigger an alert"`
}

func (c *checksHandler) HandleCreateSSLCheck(ctx context.Context, _ *mcp.CallToolRequest, in createSSLCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckSSLCert{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Notes:         in.Notes,
		Protocol:      in.Protocol,
		Threshold:     in.Threshold,
	}

	created, err := client.Checks().CreateSSLCert(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create SSL check: %w", err)
	}

	return textResult(fmt.Sprintf("Created SSL check #%d: %s", created.PK, created.Name)), nil, nil
}
