package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var DeleteCheckToolModule = fx.Module("tool.delete_check",
	fx.Invoke(func(srv *mcp.Server) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "delete_check",
			Description: "Delete a monitoring check by ID",
		}, HandleDeleteCheck)
	}),
)

type deleteCheckInput struct {
	ID int `json:"id"`
}

func HandleDeleteCheck(ctx context.Context, _ *mcp.CallToolRequest, in deleteCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	_, err = client.Checks.Delete(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete check: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted check #%d", in.ID)), nil, nil
}
