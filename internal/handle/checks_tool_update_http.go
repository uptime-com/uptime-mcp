package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateHTTPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_http_check",
		Description: "Update an existing HTTP/HTTPS monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateHTTPCheck)
}

type updateHTTPCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"URL to monitor"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes"`
	Locations     []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Port          int64    `json:"port,omitempty" jsonschema:"port number"`
	Username      string   `json:"username,omitempty" jsonschema:"HTTP basic auth username"`
	Password      string   `json:"password,omitempty" jsonschema:"HTTP basic auth password"`
	Headers       string   `json:"headers,omitempty" jsonschema:"custom HTTP headers, one per line as Header: Value"`
	SendString    string   `json:"send_string,omitempty" jsonschema:"string to send in the request body"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"string expected in the response body"`
}

func (c *checksHandler) HandleUpdateHTTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateHTTPCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckHTTP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      in.Interval,
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

	updated, err := client.Checks().UpdateHTTP(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update HTTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated HTTP check #%d: %s", updated.PK, updated.Name)), nil, nil
}
