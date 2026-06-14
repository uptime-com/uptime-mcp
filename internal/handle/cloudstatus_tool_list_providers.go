package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerListCloudStatusProvidersTool(srv *mcp.Server, h *cloudStatusHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_cloudstatus_providers",
		Description: "List all available cloud status provider names (e.g. AWS, Azure, GCP, Cloudflare). Use this to discover which providers can be monitored with create_cloudstatus_check, then use search_cloudstatus_services to find specific service components within a provider.",
	}, h.HandleListCloudStatusProviders)
}

type listCloudStatusProvidersInput struct{}

func (h *cloudStatusHandler) HandleListCloudStatusProviders(_ context.Context, _ *mcp.CallToolRequest, _ listCloudStatusProvidersInput) (*mcp.CallToolResult, any, error) {
	providers := h.index.Providers()

	var sb strings.Builder
	fmt.Fprintf(&sb, "Available cloud status providers (%d):\n\n", len(providers))
	for _, p := range providers {
		fmt.Fprintf(&sb, "- %s\n", p)
	}

	return textResult(sb.String()), nil, nil
}
