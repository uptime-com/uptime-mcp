package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateSMTPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_smtp_check",
		Description: "Update an existing SMTP email server monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateSMTPCheck)
}

type updateSMTPCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"SMTP server hostname or IP address"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes"`
	Locations     []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"port number"`
	Encryption    string   `json:"encryption,omitempty" jsonschema:"encryption mode: SSL or STARTTLS"`
	Username      string   `json:"username,omitempty" jsonschema:"SMTP authentication username"`
	Password      string   `json:"password,omitempty" jsonschema:"SMTP authentication password"`
}

func (c *checksHandler) HandleUpdateSMTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateSMTPCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckSMTP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      in.Interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		Encryption:    optString(in.Encryption),
		Username:      in.Username,
		Password:      in.Password,
	}

	updated, err := client.Checks().UpdateSMTP(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update SMTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated SMTP check #%d: %s", updated.PK, updated.Name)), nil, nil
}
