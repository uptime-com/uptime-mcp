package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
)

func registerCreateHTTPCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_http_check",
		Description: "Create a new HTTP/HTTPS monitoring check",
	}, h.HandleCreateHTTPCheck)
}

type createHTTPCheckInput struct {
	Name         string   `json:"name"`
	Address      string   `json:"address"`
	Interval     int      `json:"interval,omitempty"`
	Locations    []string `json:"locations,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Sensitivity  int      `json:"sensitivity,omitempty"`
	Notes        string   `json:"notes,omitempty"`
	Port         int      `json:"port,omitempty"`
	Username     string   `json:"username,omitempty"`
	Password     string   `json:"password,omitempty"`
	Headers      string   `json:"headers,omitempty"`
	SendString   string   `json:"send_string,omitempty"`
	ExpectString string   `json:"expect_string,omitempty"`
}

func (c *checksHandler) HandleCreateHTTPCheck(ctx context.Context, _ *mcp.CallToolRequest, in createHTTPCheckInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	check := &api.Check{
		CheckType:    "HTTP",
		Name:         in.Name,
		Address:      in.Address,
		Port:         in.Port,
		Interval:     in.Interval,
		Locations:    in.Locations,
		Tags:         in.Tags,
		Sensitivity:  in.Sensitivity,
		Notes:        in.Notes,
		Username:     in.Username,
		Password:     in.Password,
		Headers:      in.Headers,
		SendString:   in.SendString,
		ExpectString: in.ExpectString,
	}

	created, _, err := c.service.Create(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create HTTP check: %w", err)
	}

	return textResult(fmt.Sprintf("Created HTTP check #%d: %s", created.PK, created.Name)), nil, nil
}
