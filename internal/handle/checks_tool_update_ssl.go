package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateSSLCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_ssl_check",
		Description: "Update an existing SSL certificate monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateSSLCheck)
}

type updateSSLCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"hostname to check SSL certificate for"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"port number"`
	Protocol      string   `json:"protocol,omitempty" jsonschema:"protocol to use, e.g. HTTPS, IMAPS, POP3S, SMTPS"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"days before certificate expiry to trigger an alert"`
}

func (c *checksHandler) HandleUpdateSSLCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateSSLCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckSSLCert{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Notes:         in.Notes,
		Protocol:      in.Protocol,
		Threshold:     in.Threshold,
	}

	updated, err := client.Checks().UpdateSSLCert(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update SSL check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated SSL check #%d: %s", updated.PK, updated.Name)), nil, nil
}
