package uptime

import (
	"context"
	"net/http"

	api "github.com/uptime-com/uptime-client-go"
)

// ChecksService defines check operations.
type ChecksService interface {
	List(ctx context.Context, opt *api.CheckListOptions) ([]*api.Check, *http.Response, error)
	Get(ctx context.Context, pk int) (*api.Check, *http.Response, error)
	Create(ctx context.Context, check *api.Check) (*api.Check, *http.Response, error)
	Update(ctx context.Context, check *api.Check) (*api.Check, *http.Response, error)
	Delete(ctx context.Context, pk int) (*http.Response, error)
	Stats(ctx context.Context, pk int, opt *api.CheckStatsOptions) (*api.CheckStatsResponse, *http.Response, error)
}

// TagsService defines tag operations.
type TagsService interface {
	List(ctx context.Context, opt *api.TagListOptions) ([]*api.Tag, *http.Response, error)
	Get(ctx context.Context, pk int) (*api.Tag, *http.Response, error)
	Create(ctx context.Context, tag *api.Tag) (*api.Tag, *http.Response, error)
	Update(ctx context.Context, tag *api.Tag) (*api.Tag, *http.Response, error)
	Delete(ctx context.Context, pk int) (*http.Response, error)
}

// OutagesService defines outage operations.
type OutagesService interface {
	List(ctx context.Context, opt *api.OutageListOptions) ([]*api.Outage, *http.Response, error)
	Get(ctx context.Context, pk string) (*api.Outage, *http.Response, error)
}

// Client provides access to Uptime API services.
type Client interface {
	Checks() ChecksService
	Tags() TagsService
	Outages() OutagesService
}
