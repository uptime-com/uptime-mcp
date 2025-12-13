//go:build acc

package tools

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	api "github.com/uptime-com/uptime-client-go"

	"github.com/uptime-com/uptime-mcp/internal/uptime"
)

func newAcceptanceClient(t *testing.T) uptime.Client {
	t.Helper()
	token := os.Getenv("UPTIME_API_KEY")
	if token == "" {
		t.Log("UPTIME_API_KEY not set")
		t.FailNow()
	}

	baseURL := os.Getenv("UPTIME_API_URL")
	if baseURL == "" {
		baseURL = "https://uptime.com/api/v1/"
	}

	client, err := api.NewClient(&api.Config{
		Token:   token,
		BaseURL: baseURL,
	})
	require.NoError(t, err)
	return uptime.NewClient(client)
}
