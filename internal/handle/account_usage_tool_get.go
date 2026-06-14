package handle

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetAccountUsageTool(srv *mcp.Server, h *accountUsageHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_account_usage",
		Description: "Get account usage and plan limits including check counts, check type limits, and feature availability. Use this to verify remaining capacity before creating checks.",
	}, h.handleGetAccountUsage)
}

type getAccountUsageInput struct{}

func (h *accountUsageHandler) handleGetAccountUsage(ctx context.Context, _ *mcp.CallToolRequest, _ getAccountUsageInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	usage, err := client.AccountUsage().Get(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get account usage: %w", err)
	}

	keys := make([]string, 0, len(*usage))
	for k := range *usage {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("Account Usage and Plan Limits:\n\n")
	for _, k := range keys {
		fmt.Fprintf(&sb, "- %s: %v\n", k, (*usage)[k])
	}

	return textResult(sb.String()), nil, nil
}
