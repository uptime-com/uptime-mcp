package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

var CreateSSLCheckToolModule = fx.Module("tool.create_ssl_check",
	fx.Invoke(func(srv *mcp.Server) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "create_ssl_check",
			Description: "Create a new SSL certificate monitoring check",
		}, HandleCreateSSLCheck)
	}),
)

type createSSLCheckInput struct {
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Interval    int      `json:"interval,omitempty"`
	Locations   []string `json:"locations,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Sensitivity int      `json:"sensitivity,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	Port        int      `json:"port,omitempty"`
	Protocol    string   `json:"protocol,omitempty"`
}

func HandleCreateSSLCheck(ctx context.Context, _ *mcp.CallToolRequest, in createSSLCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	check := &api.Check{
		CheckType:   "SSL",
		Name:        in.Name,
		Address:     in.Address,
		Port:        in.Port,
		Interval:    in.Interval,
		Locations:   in.Locations,
		Tags:        in.Tags,
		Sensitivity: in.Sensitivity,
		Notes:       in.Notes,
		Protocol:    in.Protocol,
	}

	created, _, err := client.Checks.Create(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create SSL check: %w", err)
	}

	return textResult(fmt.Sprintf("Created SSL check #%d: %s", created.PK, created.Name)), nil, nil
}
