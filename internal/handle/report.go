package handle

import (
	"net/http"
	"time"

	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/server"
)

const (
	repoSlug         = "uptime-com/uptime-mcp"
	repoURL          = "https://github.com/" + repoSlug
	defaultGitHubAPI = "https://api.github.com"
)

type reportHandler struct {
	info      server.Info
	githubAPI string
	http      *http.Client
}

func provideReportHandler(info server.Info) *reportHandler {
	return &reportHandler{
		info:      info,
		githubAPI: defaultGitHubAPI,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

var reportModule = fx.Module("tool.report",
	fx.Provide(provideReportHandler),
	fx.Invoke(registerSearchIssuesTool),
	fx.Invoke(registerReportIssueTool),
)
