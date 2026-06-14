package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateTCPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_tcp_check",
		Description: "Create a new TCP port connectivity check. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreateTCPCheck)
}

type createTCPCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Address       string   `json:"address" jsonschema:"hostname or IP address to connect to"`
	Interval      int64    `json:"interval" jsonschema:"check frequency in minutes, defaults to 5"`
	Locations     []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting, 0 uses account default"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port" jsonschema:"TCP port number to connect to"`
	SendString    string   `json:"send_string,omitempty" jsonschema:"string to send after connecting"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"expected string in the response, check fails if not found"`
}

func (c *checksHandler) HandleCreateTCPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createTCPCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}
	if in.Port == 0 {
		return nil, nil, fmt.Errorf("port is required for TCP check")
	}

	interval := in.Interval
	if interval == 0 {
		interval = 5
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckTCP{
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

	created, err := client.Checks().CreateTCP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TCP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created TCP check #%d: %s", created.PK, created.Name)), nil, nil
}
