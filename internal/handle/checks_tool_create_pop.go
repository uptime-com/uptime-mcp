package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreatePOPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_pop_check",
		Description: "Create a new POP3 email server monitoring check. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreatePOPCheck)
}

type createPOPCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Address       string   `json:"address" jsonschema:"POP3 server hostname or IP address"`
	Interval      int64    `json:"interval" jsonschema:"check frequency in minutes, defaults to 5"`
	Locations     []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting, 0 uses account default"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"port number, defaults to 110 for POP3 or 995 for POP3S"`
	Encryption    string   `json:"encryption,omitempty" jsonschema:"encryption mode: SSL or STARTTLS"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"expected string in the server response"`
}

func (c *checksHandler) HandleCreatePOPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createPOPCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckPOP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		Encryption:    in.Encryption,
		ExpectString:  in.ExpectString,
	}

	created, err := client.Checks().CreatePOP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create POP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created POP check #%d: %s", created.PK, created.Name)), nil, nil
}
