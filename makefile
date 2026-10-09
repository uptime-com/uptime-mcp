.DEFAULT_GOAL := help

.PHONY: help build image package test e2e lint clean run/http run/claude

# Version derived from git tags (e.g. v0.16.0 -> 0.16.0). Without a tag, fall
# back to a SemVer-valid pre-release (0.0.0-<sha>) so `helm package --version`
# accepts it. --always is intentionally omitted: a bare SHA would shadow the
# fallback and break the chart version.
GIT_VERSION := $(shell git describe --tags --dirty 2>/dev/null || echo "0.0.0-$(shell git rev-parse --short HEAD)")
VERSION     := $(GIT_VERSION:v%=%)
COMMIT      := $(shell git rev-parse --short HEAD)
KO_DOCKER_REPO := ghcr.io/uptime-com/uptime-mcp
# helm push appends the chart name, landing it at $(CHART_REPO)/uptime-mcp.
CHART_REPO     := ghcr.io/uptime-com/uptime-mcp/charts
# Image platforms for the multi-arch manifest (override on the command line).
PLATFORMS      := linux/amd64,linux/arm64
# Image tags to push. CI overrides this for the floating main/latest pointers.
TAGS           ?= $(VERSION)
export KO_DOCKER_REPO VERSION COMMIT

# Uptime.com instance URL. Override for self-hosted/regional instances.
UPTIME_URL ?= https://uptime.com

# OAuth2 client ID for the stdio browser login used by `run/claude`.
# Register your own OAuth application in your Uptime.com account
# (Settings -> API & Integrations) and pass it in, e.g.:
#   make run/claude UPTIME_OAUTH_CLIENT_ID=xxxxxxxx
# Alternatively export UPTIME_BEARER_TOKEN to skip OAuth entirely.
UPTIME_OAUTH_CLIENT_ID ?=

help: ## Show this help
	@grep -E '^[a-zA-Z0-9/]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

build: image package ## Build and push container image and Helm chart

image: ## Build and push the multi-arch container image with ko
	@echo "Building and pushing image: $(KO_DOCKER_REPO):$(TAGS)"
	ko build --bare --platform=$(PLATFORMS) --tags $(TAGS) .

package: ## Package and push the Helm chart to the OCI registry
	@echo "Packaging and pushing chart: $(CHART_REPO)/uptime-mcp:$(VERSION)"
	@helm package charts/uptime-mcp --version $(VERSION) --app-version $(VERSION) --destination .build/
	@helm push .build/uptime-mcp-$(VERSION).tgz oci://$(CHART_REPO)
	@rm .build/uptime-mcp-$(VERSION).tgz
	@echo "Chart pushed successfully"

test: ## Run unit tests
	go test ./...

e2e: ## Run e2e tests, requires UPTIME_BEARER_TOKEN
	go test -tags=e2e -v ./internal/e2e/...

lint: ## Run linters
	golangci-lint run ./...

clean: ## Remove build artifacts
	rm -f uptime-mcp
	rm -rf .build/ dist/

run/http: ## Start HTTP server on :8080
	go run . \
		-transport=http \
		-listen=:8080 \
		-uptime-url=$(UPTIME_URL) \
		-log-level=debug

run/claude: ## Launch Claude Code with stdio MCP (needs UPTIME_OAUTH_CLIENT_ID or UPTIME_BEARER_TOKEN)
	@if [ -z "$(UPTIME_OAUTH_CLIENT_ID)" ] && [ -z "$$UPTIME_BEARER_TOKEN" ]; then \
		echo "set UPTIME_OAUTH_CLIENT_ID=<your-client-id> or export UPTIME_BEARER_TOKEN=<token>"; exit 1; \
	fi
	@printf '{"mcpServers":{"uptime":{"type":"stdio","command":"go","args":["run",".","-transport=stdio","-uptime-url=$(UPTIME_URL)","-client-id=$(UPTIME_OAUTH_CLIENT_ID)","-log-level=debug"]}}}' >"$${TMPDIR:-/tmp}/claude-uptime-mcp.json"
	claude --mcp-config "$${TMPDIR:-/tmp}/claude-uptime-mcp.json"
