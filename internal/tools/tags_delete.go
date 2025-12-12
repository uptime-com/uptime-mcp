package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var DeleteTagToolModule = fx.Module("tool.delete_tag",
	fx.Invoke(func(srv *mcp.Server) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "delete_tag",
			Description: "Delete a check tag by ID",
		}, HandleDeleteTag)
	}),
)

type deleteTagInput struct {
	ID int `json:"id"`
}

func HandleDeleteTag(ctx context.Context, _ *mcp.CallToolRequest, in deleteTagInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	_, err = client.Tags.Delete(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete tag: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted tag #%d", in.ID)), nil, nil
}
