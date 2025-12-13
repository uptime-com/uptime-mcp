package uptime

import api "github.com/uptime-com/uptime-client-go"

type clientWrapper struct {
	client *api.Client
}

// NewClient wraps an api.Client to satisfy the Client interface.
func NewClient(c *api.Client) Client {
	return &clientWrapper{client: c}
}

func (w *clientWrapper) Checks() ChecksService   { return w.client.Checks }
func (w *clientWrapper) Tags() TagsService       { return w.client.Tags }
func (w *clientWrapper) Outages() OutagesService { return w.client.Outages }
