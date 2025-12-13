package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
)

func registerCreateSMTPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_smtp_check",
		Description: "Create a new SMTP email server monitoring check",
	}, h.HandleCreateSMTPCheck)
}

type createSMTPCheckInput struct {
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Interval    int      `json:"interval,omitempty"`
	Locations   []string `json:"locations,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Sensitivity int      `json:"sensitivity,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	Port        int      `json:"port,omitempty"`
	Encryption  string   `json:"encryption,omitempty"`
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
}

func (c *checksHandler) HandleCreateSMTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createSMTPCheckInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	check := &api.Check{
		CheckType:   "SMTP",
		Name:        in.Name,
		Address:     in.Address,
		Port:        in.Port,
		Interval:    in.Interval,
		Locations:   in.Locations,
		Tags:        in.Tags,
		Sensitivity: in.Sensitivity,
		Notes:       in.Notes,
		Encryption:  in.Encryption,
		Username:    in.Username,
		Password:    in.Password,
	}

	created, _, err := c.service.Create(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create SMTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created SMTP check #%d: %s", created.PK, created.Name)), nil, nil
}
