package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// recordingCBD stands in for the rest of the client, answering each request
// with the next queued status and keeping what it was sent.
type recordingCBD struct {
	upapi.CBD
	statuses []int
	auths    []string
	bodies   []string
}

func (c *recordingCBD) Do(rq *http.Request) (*http.Response, error) {
	body := ""
	if rq.Body != nil {
		b, err := io.ReadAll(rq.Body)
		if err != nil {
			return nil, err
		}
		body = string(b)
	}
	c.auths = append(c.auths, rq.Header.Get("Authorization"))
	c.bodies = append(c.bodies, body)

	status := http.StatusOK
	if len(c.statuses) > 0 {
		status, c.statuses = c.statuses[0], c.statuses[1:]
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(nil)),
		Header:     http.Header{},
	}, nil
}

// post builds a request the way upapi does, with a body worth replaying.
func post(t *testing.T, body string) *http.Request {
	t.Helper()
	rq, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
		"http://example.invalid/api/v1/checks/", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	return rq
}

func wrap(t *testing.T, inner upapi.CBD, scheme app.AuthScheme) upapi.CBD {
	t.Helper()
	cbd, err := withUpstreamAuth("tok", scheme)(inner)
	require.NoError(t, err)
	return cbd
}

func TestUpstreamAuthReplaysTheBody(t *testing.T) {
	t.Run("a corrected POST is replayed with its body intact", func(t *testing.T) {
		inner := &recordingCBD{statuses: []int{http.StatusUnauthorized, http.StatusCreated}}

		rs, err := wrap(t, inner, app.SchemeBearer).Do(post(t, `{"name":"check"}`))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rs.StatusCode)

		assert.Equal(t, []string{"Bearer tok", "Token tok"}, inner.auths)
		assert.Equal(t, []string{`{"name":"check"}`, `{"name":"check"}`}, inner.bodies,
			"the retry must carry the same body as the refused attempt")
	})

	t.Run("a bodyless request is replayed too", func(t *testing.T) {
		inner := &recordingCBD{statuses: []int{http.StatusUnauthorized, http.StatusOK}}

		rq, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
			"http://example.invalid/api/v1/checks/", nil)
		require.NoError(t, err)

		rs, err := wrap(t, inner, app.SchemeToken).Do(rq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rs.StatusCode)
		assert.Equal(t, []string{"Token tok", "Bearer tok"}, inner.auths)
	})

	t.Run("the caller's request is left untouched", func(t *testing.T) {
		inner := &recordingCBD{}
		rq := post(t, `{"name":"check"}`)

		_, err := wrap(t, inner, app.SchemeBearer).Do(rq)
		require.NoError(t, err)

		assert.Empty(t, rq.Header.Get("Authorization"),
			"the header belongs on the copy that was sent, not on the caller's request")
	})
}
