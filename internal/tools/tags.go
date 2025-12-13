package tools

import "github.com/uptime-com/uptime-mcp/internal/uptime"

type tags struct {
	service uptime.TagsService
}

func provideTags(c uptime.Client) *tags {
	return &tags{service: c.Tags()}
}
