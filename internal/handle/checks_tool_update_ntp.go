package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateNTPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_ntp_check",
		Description: "Update an existing NTP (Network Time Protocol) monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateNTPCheck)
}

type updateNTPCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"NTP server hostname or IP address"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes"`
	Locations     []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"NTP port number"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"maximum allowed time offset in milliseconds"`
}

func (c *checksHandler) HandleUpdateNTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateNTPCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckNTP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      in.Interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Threshold:     in.Threshold,
		Notes:         in.Notes,
	}

	updated, err := client.Checks().UpdateNTP(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update NTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated NTP check #%d: %s", updated.PK, updated.Name)), nil, nil
}
