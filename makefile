.PHONY: help test e2e run/http run/claude

UPTIME_URL ?= https://sandbox.upeks.net
UPTIME_OAUTH_CLIENT_ID ?= MYH77e5qvqbYjKU01EBgIYU8ZQwhVGxpmfUKBUHU

help: ## Show this help
	@grep -E '^[a-zA-Z0-9/]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

test: ## Run unit tests
	go test ./...

e2e: ## Run e2e tests, requires UPTIME_BEARER_TOKEN
	go test -tags=e2e -v ./e2e/...

run/http: ## Start HTTP server on :8080
	go run ./app/uptime-mcp \
		-transport=http \
		-listen=:8080 \
		-uptime-url=$(UPTIME_URL) \
		-log-level=debug

run/claude: ## Launch Claude Code with stdio MCP
	@printf '{"mcpServers":{"uptime":{"type":"stdio","command":"go","args":["run","./app/uptime-mcp","-transport=stdio","-uptime-url=$(UPTIME_URL)","-client-id=$(UPTIME_OAUTH_CLIENT_ID)","-log-level=debug"]}}}' >/tmp/claude-uptime-mcp.json
	claude --mcp-config /tmp/claude-uptime-mcp.json
