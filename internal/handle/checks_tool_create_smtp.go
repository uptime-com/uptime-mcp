package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateSMTPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_smtp_check",
		Description: "Create a new SMTP email server monitoring check",
	}, h.HandleCreateSMTPCheck)
}

type createSMTPCheckInput struct {
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	Interval      int64    `json:"interval,omitempty"`
	Locations     []string `json:"locations,omitempty"`
	ContactGroups []string `json:"contact_groups,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Sensitivity   int64    `json:"sensitivity,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	Port          int64    `json:"port,omitempty"`
	Encryption    string   `json:"encryption,omitempty"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password,omitempty"`
}

func (c *checksHandler) HandleCreateSMTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createSMTPCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckSMTP{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Interval:      in.Interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		Notes:         in.Notes,
		Encryption:    in.Encryption,
		Username:      in.Username,
		Password:      in.Password,
	}

	created, err := client.Checks().CreateSMTP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create SMTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created SMTP check #%d: %s", created.PK, created.Name)), nil, nil
}
