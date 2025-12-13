package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
)

func registerCreateTCPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_tcp_check",
		Description: "Create a new TCP port connectivity check",
	}, h.HandleCreateTCPCheck)
}

type createTCPCheckInput struct {
	Name         string   `json:"name"`
	Address      string   `json:"address"`
	Interval     int      `json:"interval,omitempty"`
	Locations    []string `json:"locations,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Sensitivity  int      `json:"sensitivity,omitempty"`
	Notes        string   `json:"notes,omitempty"`
	Port         int      `json:"port"`
	SendString   string   `json:"send_string,omitempty"`
	ExpectString string   `json:"expect_string,omitempty"`
}

func (c *checksHandler) HandleCreateTCPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createTCPCheckInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}
	if in.Port == 0 {
		return nil, nil, fmt.Errorf("port is required for TCP check")
	}

	check := &api.Check{
		CheckType:    "TCP",
		Name:         in.Name,
		Address:      in.Address,
		Port:         in.Port,
		Interval:     in.Interval,
		Locations:    in.Locations,
		Tags:         in.Tags,
		Sensitivity:  in.Sensitivity,
		Notes:        in.Notes,
		SendString:   in.SendString,
		ExpectString: in.ExpectString,
	}

	created, _, err := c.service.Create(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TCP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created TCP check #%d: %s", created.PK, created.Name)), nil, nil
}
