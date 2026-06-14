package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateNTPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_ntp_check",
		Description: "Create a new NTP (Network Time Protocol) monitoring check. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreateNTPCheck)
}

type createNTPCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Address       string   `json:"address" jsonschema:"NTP server hostname or IP address"`
	Interval      int64    `json:"interval" jsonschema:"check frequency in minutes, defaults to 5"`
	Locations     []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting, 0 uses account default"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"NTP port number, defaults to 123"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"maximum allowed time offset in milliseconds"`
}

func (c *checksHandler) HandleCreateNTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createNTPCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	interval := in.Interval
	if interval == 0 {
		interval = 5
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckNTP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Threshold:     in.Threshold,
		Notes:         in.Notes,
	}

	created, err := client.Checks().CreateNTP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create NTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created NTP check #%d: %s", created.PK, created.Name)), nil, nil
}
