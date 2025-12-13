package handle

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const locationURIPrefix = "https://uptime.com/api/v1/probe-servers/"

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

	// Fetch all servers and find the matching one
	result, err := client.ProbeServers().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	for _, s := range result.Items {
		if s.Location == location {
			var sb strings.Builder
			fmt.Fprintf(&sb, "Location: %s\n", s.Location)
			fmt.Fprintf(&sb, "Probe Name: %s\n", s.ProbeName)
			if len(s.IPv4Addresses) > 0 {
				fmt.Fprintf(&sb, "IPv4 Addresses: %s\n", strings.Join(s.IPv4Addresses, ", "))
			}
			if len(s.IPv6Addresses) > 0 {
				fmt.Fprintf(&sb, "IPv6 Addresses: %s\n", strings.Join(s.IPv6Addresses, ", "))
			}

			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      uri,
					MIMEType: "text/plain",
					Text:     sb.String(),
				}},
			}, nil
		}
	}

	return nil, fmt.Errorf("location not found: %s", location)
}
