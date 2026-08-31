package server

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// withUpstreamAuth authenticates every api/v1 request with token, presented
// under scheme, and corrects the scheme once if api/v1 refuses it.
//
// The correction is what makes the caller's scheme a starting point rather
// than a promise: an OAuth2 access token and an account API key are both
// opaque strings, so a credential arriving by a route that carries either one
// cannot be classified before it is used.
func withUpstreamAuth(token string, scheme app.AuthScheme) upapi.Option {
	return func(cbd upapi.CBD) (upapi.CBD, error) {
		return &upstreamAuthCBD{CBD: cbd, token: token, scheme: scheme}, nil
	}
}

type upstreamAuthCBD struct {
	upapi.CBD
	token string

	mu       sync.Mutex
	scheme   app.AuthScheme
	resolved bool
}

// Do sends rq authenticated under the current scheme, retrying once under the
// other scheme if api/v1 answers 401, and keeps whichever was accepted.
//
// A 401 means api/v1 rejected the credential before reading the request, so
// replaying is safe for every method rather than for the idempotent ones.
func (c *upstreamAuthCBD) Do(rq *http.Request) (*http.Response, error) {
	scheme, resolved := c.current()

	body, err := drainRequest(rq)
	if err != nil {
		return nil, err
	}

	rs, err := c.attempt(rq, scheme, body)
	if err != nil {
		return nil, err
	}
	if rs.StatusCode != http.StatusUnauthorized {
		c.settle(scheme)
		return rs, nil
	}
	if resolved {
		return rs, nil
	}

	discard(rs)

	other := scheme.Other()
	rs, err = c.attempt(rq, other, body)
	if err != nil {
		return nil, err
	}
	if rs.StatusCode != http.StatusUnauthorized {
		c.settle(other)
	}
	return rs, nil
}

// attempt sends a copy of rq carrying body and the Authorization header for
// scheme, leaving rq itself untouched so it can be sent again.
func (c *upstreamAuthCBD) attempt(rq *http.Request, scheme app.AuthScheme, body []byte) (*http.Response, error) {
	out := rq.Clone(rq.Context())
	if body != nil {
		out.Body = io.NopCloser(bytes.NewReader(body))
		out.ContentLength = int64(len(body))
	}
	out.Header.Set("Authorization", string(scheme)+" "+c.token)
	return c.CBD.Do(out)
}

func (c *upstreamAuthCBD) current() (app.AuthScheme, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.scheme, c.resolved
}

func (c *upstreamAuthCBD) settle(scheme app.AuthScheme) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scheme, c.resolved = scheme, true
}

// drainRequest reads rq's body into memory so the request can be sent twice,
// returning nil for a request that carries none.
func drainRequest(rq *http.Request) ([]byte, error) {
	if rq.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(rq.Body)
	_ = rq.Body.Close()
	if err != nil {
		return nil, err
	}
	rq.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

// discard reads and closes rs so its connection returns to the pool.
func discard(rs *http.Response) {
	_, _ = io.Copy(io.Discard, rs.Body)
	_ = rs.Body.Close()
}
