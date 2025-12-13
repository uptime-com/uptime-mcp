package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const checkURIPrefix = "https://uptime.com/api/v1/checks/"

func registerCheckResource(srv *mcp.Server, h *checksHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: checkURIPrefix + "{id}",
		Name:        "check",
		Description: "Uptime.com monitoring check details",
		MIMEType:    "text/plain",
	}, h.handleCheckResource)
}

func (h *checksHandler) handleCheckResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, checkURIPrefix)
	if idStr == uri {
		return nil, fmt.Errorf("invalid check URI: %s", uri)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid check ID: %s", idStr)
	}

	check, err := h.service.Get(ctx, upapi.PrimaryKey(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get check: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Check #%d: %s\n", check.PK, check.Name)
	fmt.Fprintf(&sb, "Type: %s\n", check.CheckType)
	fmt.Fprintf(&sb, "Address: %s\n", check.Address)
	if check.Port > 0 {
		fmt.Fprintf(&sb, "Port: %d\n", check.Port)
	}
	fmt.Fprintf(&sb, "Interval: %d seconds\n", check.Interval)
	fmt.Fprintf(&sb, "Sensitivity: %d\n", check.Sensitivity)
	if len(check.Locations) > 0 {
		fmt.Fprintf(&sb, "Locations: %s\n", strings.Join(check.Locations, ", "))
	}
	if len(check.Tags) > 0 {
		fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(check.Tags, ", "))
	}
	if check.Notes != "" {
		fmt.Fprintf(&sb, "Notes: %s\n", check.Notes)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     sb.String(),
		}},
	}, nil
}
