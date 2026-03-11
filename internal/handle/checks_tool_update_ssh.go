package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateSSHCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_ssh_check",
		Description: "Update an existing SSH connectivity monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateSSHCheck)
}

type updateSSHCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"hostname or IP address of the SSH server"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes"`
	Locations     []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"SSH port number"`
}

func (c *checksHandler) HandleUpdateSSHCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateSSHCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckSSH{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      in.Interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
	}

	updated, err := client.Checks().UpdateSSH(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update SSH check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated SSH check #%d: %s", updated.PK, updated.Name)), nil, nil
}
