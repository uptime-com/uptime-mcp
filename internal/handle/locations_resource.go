package handle

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const locationURIPrefix = "uptime://locations/"

func registerLocationResource(srv *mcp.Server, h *locationsHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: locationURIPrefix + "{location}",
		Name:        "location",
		Description: "Uptime.com probe server location details including IP addresses",
		MIMEType:    "text/plain",
	}, h.handleLocationResource)
}

func (h *locationsHandler) handleLocationResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	uri := req.Params.URI

	locationEncoded := strings.TrimPrefix(uri, locationURIPrefix)
	if locationEncoded == uri {
		return nil, fmt.Errorf("invalid location URI: %s", uri)
	}

	location, err := url.PathUnescape(locationEncoded)
	if err != nil {
		return nil, fmt.Errorf("invalid location encoding: %s", locationEncoded)
	}

	if excludedLocations[location] {
		return nil, fmt.Errorf("pseudo-location not supported: %s", location)
	}

	var sb strings.Builder
	if err := h.loadLocation(ctx, client, location, &sb); err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     sb.String(),
		}},
	}, nil
}

func (h *locationsHandler) loadLocation(ctx context.Context, client upapi.API, location string, sb *strings.Builder) error {
	result, err := client.ProbeServers().List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list locations: %w", err)
	}

	for _, s := range result.Items {
		if s.Location == location {
			fmt.Fprintf(sb, "Location: %s\n", s.Location)
			fmt.Fprintf(sb, "Probe Name: %s\n", s.ProbeName)
			if len(s.IPv4Addresses) > 0 {
				fmt.Fprintf(sb, "IPv4 Addresses: %s\n", strings.Join(s.IPv4Addresses, ", "))
			}
			if len(s.IPv6Addresses) > 0 {
				fmt.Fprintf(sb, "IPv6 Addresses: %s\n", strings.Join(s.IPv6Addresses, ", "))
			}
			return nil
		}
	}

	return fmt.Errorf("location not found: %s", location)
}
