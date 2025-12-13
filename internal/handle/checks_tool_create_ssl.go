package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateSSLCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_ssl_check",
		Description: "Create a new SSL certificate monitoring check",
	}, h.HandleCreateSSLCheck)
}

type createSSLCheckInput struct {
	Name          string   `json:"name"`
	Address       string   `json:"address"`
	Locations     []string `json:"locations"`
	ContactGroups []string `json:"contact_groups"`
	Tags          []string `json:"tags,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	Port          int64    `json:"port,omitempty"`
	Protocol      string   `json:"protocol,omitempty"`
	Threshold     int64    `json:"threshold,omitempty"`
}

func (c *checksHandler) HandleCreateSSLCheck(ctx context.Context, _ *mcp.CallToolRequest, in createSSLCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckSSLCert{
		Name:          in.Name,
		Address:       in.Address,
		Port:          in.Port,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Notes:         in.Notes,
		Protocol:      in.Protocol,
		Threshold:     in.Threshold,
	}

	created, err := client.Checks().CreateSSLCert(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create SSL check: %w", err)
	}

	return textResult(fmt.Sprintf("Created SSL check #%d: %s", created.PK, created.Name)), nil, nil
}
