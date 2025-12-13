package handle

import "github.com/uptime-com/uptime-client-go/v2/pkg/upapi"

// These interfaces are used by mockery for code generation only.
//
//nolint:unused
type (
	// mockeryClient wraps upapi.API for mock generation.
	mockeryClient interface {
		upapi.API
	}

	// mockeryChecksService wraps upapi.ChecksEndpoint for mock generation.
	mockeryChecksService interface {
		upapi.ChecksEndpoint
	}

	// mockeryTagsService wraps upapi.TagsEndpoint for mock generation.
	mockeryTagsService interface {
		upapi.TagsEndpoint
	}

	// mockeryOutagesService wraps upapi.OutagesEndpoint for mock generation.
	mockeryOutagesService interface {
		upapi.OutagesEndpoint
	}

	// mockeryContactsService wraps upapi.ContactsEndpoint for mock generation.
	mockeryContactsService interface {
		upapi.ContactsEndpoint
	}
)
