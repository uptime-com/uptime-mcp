package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateUDPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_udp_check",
		Description: "Update an existing UDP monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateUDPCheck)
}

type updateUDPCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"hostname or IP address to monitor"`
	Port          int64    `json:"port,omitempty" jsonschema:"UDP port number to check"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes"`
	Locations     []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	SendString    string   `json:"send_string,omitempty" jsonschema:"string to send to the UDP port"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"expected string in the response"`
}

func (c *checksHandler) HandleUpdateUDPCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateUDPCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckUDP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      in.Interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		SendString:    in.SendString,
		ExpectString:  in.ExpectString,
	}

	updated, err := client.Checks().UpdateUDP(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update UDP check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated UDP check #%d: %s", updated.PK, updated.Name)), nil, nil
}
