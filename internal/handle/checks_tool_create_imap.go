package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateIMAPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_imap_check",
		Description: "Create a new IMAP email server monitoring check",
	}, h.HandleCreateIMAPCheck)
}

type createIMAPCheckInput struct {
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	Interval      int64    `json:"interval"`
	Locations     []string `json:"locations"`
	ContactGroups []string `json:"contact_groups"`
	Tags          []string `json:"tags,omitempty"`
	Sensitivity   int64    `json:"sensitivity,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	Port          int64    `json:"port,omitempty"`
	Encryption    string   `json:"encryption,omitempty"`
	ExpectString  string   `json:"expect_string,omitempty"`
}

func (c *checksHandler) HandleCreateIMAPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createIMAPCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckIMAP{
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
		ExpectString:  in.ExpectString,
	}

	created, err := client.Checks().CreateIMAP(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create IMAP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created IMAP check #%d: %s", created.PK, created.Name)), nil, nil
}
