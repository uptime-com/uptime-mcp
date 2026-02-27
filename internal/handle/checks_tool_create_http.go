package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateHTTPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_http_check",
		Description: "Create a new HTTP/HTTPS monitoring check. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreateHTTPCheck)
}

type createHTTPCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Address       string   `json:"address" jsonschema:"URL to monitor, e.g. https://example.com"`
	Interval      int64    `json:"interval" jsonschema:"check frequency in minutes, defaults to 5"`
	Locations     []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting, 0 uses account default"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"port number, defaults to 80 for HTTP or 443 for HTTPS"`
	Username      string   `json:"username,omitempty" jsonschema:"HTTP basic auth username"`
	Password      string   `json:"password,omitempty" jsonschema:"HTTP basic auth password"`
	Headers       string   `json:"headers,omitempty" jsonschema:"custom HTTP headers, one per line as Header: Value"`
	SendString    string   `json:"send_string,omitempty" jsonschema:"string to send in the request body"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"string expected in the response body, check fails if not found"`
}

func (c *checksHandler) HandleCreateHTTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createHTTPCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	// Default interval to 5 minutes if not specified
	interval := in.Interval
	if interval == 0 {
		interval = 5
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckHTTP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		Username:      in.Username,
		Password:      in.Password,
		Headers:       in.Headers,
		SendString:    in.SendString,
		ExpectString:  in.ExpectString,
	}

	created, err := client.Checks().CreateHTTP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create HTTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created HTTP check #%d: %s", created.PK, created.Name)), nil, nil
}
