package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateUDPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_udp_check",
		Description: "Create a new UDP monitoring check. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreateUDPCheck)
}

type createUDPCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Address       string   `json:"address" jsonschema:"hostname or IP address to monitor"`
	Port          int64    `json:"port" jsonschema:"UDP port number to check"`
	Interval      int64    `json:"interval" jsonschema:"check frequency in minutes, defaults to 5"`
	Locations     []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting, 0 uses account default"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	SendString    string   `json:"send_string,omitempty" jsonschema:"string to send to the UDP port"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"expected string in the response, check fails if not found"`
}

func (c *checksHandler) HandleCreateUDPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createUDPCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}
	if in.Port == 0 {
		return nil, nil, fmt.Errorf("port is required for UDP check")
	}

	interval := in.Interval
	if interval == 0 {
		interval = 5
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckUDP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		SendString:    in.SendString,
		ExpectString:  in.ExpectString,
	}

	created, err := client.Checks().CreateUDP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create UDP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created UDP check #%d: %s", created.PK, created.Name)), nil, nil
}
