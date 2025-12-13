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
		Description: "Create a new HTTP/HTTPS monitoring check",
	}, h.HandleCreateHTTPCheck)
}

type createHTTPCheckInput struct {
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	Interval      int64    `json:"interval,omitempty"`
	Locations     []string `json:"locations,omitempty"`
	ContactGroups []string `json:"contact_groups,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Sensitivity   int64    `json:"sensitivity,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	Port          int64    `json:"port,omitempty"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password,omitempty"`
	Headers       string   `json:"headers,omitempty"`
	SendString    string   `json:"send_string,omitempty"`
	ExpectString  string   `json:"expect_string,omitempty"`
}

func (c *checksHandler) HandleCreateHTTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createHTTPCheckInput) (*mcp.CallToolResult, any, error) {
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

	created, err := c.service.CreateHTTP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create HTTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created HTTP check #%d: %s", created.PK, created.Name)), nil, nil
}
