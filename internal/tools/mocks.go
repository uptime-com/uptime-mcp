package tools

import "github.com/uptime-com/uptime-mcp/internal/uptime"

type (
	// mockeryClient wraps uptime.Client for mock generation.
	mockeryClient interface {
		uptime.Client
	}

	// mockeryChecksService wraps uptime.ChecksService for mock generation.
	mockeryChecksService interface {
		uptime.ChecksService
	}

	// mockeryTagsService wraps uptime.TagsService for mock generation.
	mockeryTagsService interface {
		uptime.TagsService
	}

	// mockeryOutagesService wraps uptime.OutagesService for mock generation.
	mockeryOutagesService interface {
		uptime.OutagesService
	}
)
