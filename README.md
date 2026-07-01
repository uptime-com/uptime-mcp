# uptime-mcp

[![CI](https://github.com/uptime-com/uptime-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/uptime-com/uptime-mcp/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/uptime-com/uptime-mcp)](https://goreportcard.com/report/github.com/uptime-com/uptime-mcp)
[![Go Reference](https://pkg.go.dev/badge/github.com/uptime-com/uptime-mcp.svg)](https://pkg.go.dev/github.com/uptime-com/uptime-mcp)
[![GitHub License](https://img.shields.io/github/license/uptime-com/uptime-mcp)](LICENSE)
[![GitHub Release](https://img.shields.io/github/v/tag/uptime-com/uptime-mcp?label=release)](https://github.com/uptime-com/uptime-mcp/tags)
[![Go Version](https://img.shields.io/github/go-mod/go-version/uptime-com/uptime-mcp)](go.mod)
[![GitHub Stars](https://img.shields.io/github/stars/uptime-com/uptime-mcp)](https://github.com/uptime-com/uptime-mcp/stargazers)

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server for
[Uptime.com](https://uptime.com). It exposes Uptime.com website, server, and
infrastructure monitoring as MCP tools, so AI assistants such as Claude can
create and manage checks, read alerts and outages, manage status pages, and
inspect account usage on your behalf.

The server speaks two transports: **stdio** (for local MCP clients like Claude
Desktop, Claude Code, and Cursor) and **streamable HTTP** (for hosted
deployments). It authenticates with a static API bearer token or a
browser-based OAuth2 PKCE flow.

## Claude Code plugin

If you use [Claude Code](https://claude.com/claude-code), the easiest way in is
the [`uptime-skills`](https://github.com/uptime-com/uptime-skills) plugin. It
bundles this MCP server with task-focused **skills** — curated context and
workflows for Uptime.com (choosing probe locations, tuning outage sensitivity,
picking the right check type for a target) — plus a sensible default permission
set. The tools alone give Claude the ability to call the API; the skills give it
the know-how to use them well, so you get sound monitoring decisions instead of
raw API calls.

```bash
/plugin marketplace add uptime-com/uptime-skills
/plugin install uptime@uptime-com
```

Then authenticate (browser OAuth, tokens stored and refreshed for you):

```bash
/mcp
```

See [uptime-com/uptime-skills](https://github.com/uptime-com/uptime-skills) for
team and project-level setup.

## Hosted server

Uptime.com runs an official hosted instance, so you do not have to run anything
yourself. Point any streamable-HTTP MCP client at:

```
https://mcp.uptime.com/mcp
```

Authenticate with your Uptime.com API token (generate one under **Settings → API
& Integrations**). Example MCP client configuration:

```json
{
  "mcpServers": {
    "uptime": {
      "type": "http",
      "url": "https://mcp.uptime.com/mcp",
      "headers": {
        "Authorization": "Bearer <your-api-token>"
      }
    }
  }
}
```

The endpoint also advertises [RFC 9728](https://www.rfc-editor.org/rfc/rfc9728)
protected-resource metadata, so OAuth2-capable MCP clients can discover the
authorization server (`https://uptime.com`) and obtain tokens themselves instead
of supplying a static token.

Prefer to run it yourself? Continue below.

## Quick start

The fastest way to try it is with `go run` (Go 1.26+) and a personal Uptime.com
API token. Generate a token in the Uptime.com UI under **Settings → API &
Integrations**, then:

```bash
export UPTIME_BEARER_TOKEN=<your-api-token>
go run github.com/uptime-com/uptime-mcp@latest -transport=stdio
```

### Claude Desktop / Claude Code / Cursor (bearer token)

Add the server to your MCP client config (e.g. `claude_desktop_config.json`,
Cursor `mcp.json`, or `claude mcp add-json`):

```json
{
  "mcpServers": {
    "uptime": {
      "command": "uptime-mcp",
      "args": ["-transport=stdio"],
      "env": {
        "UPTIME_BEARER_TOKEN": "<your-api-token>"
      }
    }
  }
}
```

Replace `"command": "uptime-mcp"` with the absolute path to a downloaded binary,
or run the container image:

```json
{
  "mcpServers": {
    "uptime": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "UPTIME_BEARER_TOKEN",
        "ghcr.io/uptime-com/uptime-mcp:latest",
        "-transport=stdio"
      ],
      "env": {
        "UPTIME_BEARER_TOKEN": "<your-api-token>"
      }
    }
  }
}
```

### Claude Desktop (OAuth2, no static token)

Instead of a static token, let the server run a browser-based OAuth2 login on
the first tool call. This needs an OAuth application registered in your
Uptime.com account (a client ID). No token is stored in your config:

```json
{
  "mcpServers": {
    "uptime": {
      "command": "uptime-mcp",
      "args": [
        "-transport=stdio",
        "-uptime-url=https://uptime.com",
        "-client-id=<your-client-id>"
      ]
    }
  }
}
```

On the first tool call the server opens your browser to complete authorization,
then refreshes the token in the background. See [Authentication](#authentication).

## Installation

### Prebuilt binary

Download the archive for your platform from the
[Releases](https://github.com/uptime-com/uptime-mcp/releases) page (Linux,
macOS, and Windows on amd64 and arm64), extract it, and put `uptime-mcp` on your
`PATH`:

```bash
tar xzf uptime-mcp_<version>_<os>_<arch>.tar.gz
sudo mv uptime-mcp /usr/local/bin/
uptime-mcp -version
```

### Container image

Published to GitHub Container Registry. Tags: `:<version>` (immutable, e.g.
`:0.16.0`), `:latest` (newest stable release), `:main` (rolling `main` HEAD).

```bash
docker pull ghcr.io/uptime-com/uptime-mcp:latest
docker run -i --rm -e UPTIME_BEARER_TOKEN ghcr.io/uptime-com/uptime-mcp:latest -transport=stdio
```

### Helm chart (HTTP mode on Kubernetes)

Published as an OCI chart. Per-environment URLs are not baked in; set them at
install time.

```bash
helm install uptime-mcp oci://ghcr.io/uptime-com/uptime-mcp/charts/uptime-mcp \
  --set config.uptimeUrl=https://uptime.com
```

### go install

```bash
go install github.com/uptime-com/uptime-mcp@latest
```

The binary is installed as `uptime-mcp` in `$(go env GOPATH)/bin`.

### Build from source

```bash
git clone https://github.com/uptime-com/uptime-mcp.git
cd uptime-mcp
go build -o uptime-mcp .
```

## Configuration

Configuration is via CLI flags, with environment-variable fallbacks for the
sensitive values.

| Flag             | Env fallback                 | Default                   | Description                                                         |
|------------------|------------------------------|---------------------------|---------------------------------------------------------------------|
| `-transport`     | —                            | `stdio`                   | Transport mode: `stdio` or `http`.                                  |
| `-listen`        | —                            | `:8080`                   | HTTP listen address (HTTP mode only).                               |
| `-uptime-url`    | `UPTIME_URL`                 | _(required)_              | Uptime.com instance URL, e.g. `https://uptime.com`. The API base is `<uptime-url>/api/v1/`. |
| `-api-url`       | `UPTIME_API_URL`             | _(from `-uptime-url`)_    | Full API base URL override, used verbatim (e.g. `http://uptime.svc.cluster.local/api/v1/`). |
| `-oauth-url`     | `UPTIME_OAUTH_URL`           | _(from `-uptime-url`)_    | Full OAuth2 authorization server URL override, used verbatim as the issuer. |
| `-resource-url`  | `UPTIME_RESOURCE_URL`        | `http://localhost:<port>` | Public URL of this server, for OAuth2 protected-resource metadata.  |
| `-client-id`     | `UPTIME_OAUTH_CLIENT_ID`     | _(empty)_                 | OAuth2 client ID.                                                   |
| `-client-secret` | `UPTIME_OAUTH_CLIENT_SECRET` | _(empty)_                 | OAuth2 client secret (confidential clients).                        |
| `-log-level`     | —                            | `error`                   | Log level: `debug`, `info`, `warn`, `error`.                        |
| `-version`       | —                            | —                         | Print version and commit, then exit.                               |

Token environment variable:

| Variable              | Description                                                               |
|-----------------------|---------------------------------------------------------------------------|
| `UPTIME_BEARER_TOKEN` | Static Uptime.com API token. Forwarded as-is, without verification or refresh. |

## Authentication

### Bearer token (simplest)

Set `UPTIME_BEARER_TOKEN` to a pre-obtained Uptime.com API token. Works in both
stdio and HTTP modes; the token is forwarded to the Uptime.com API as-is, with
no OAuth configuration, verification, or refresh.

```bash
UPTIME_BEARER_TOKEN=<your-api-token> uptime-mcp -transport=stdio
```

### OAuth2 (stdio, browser PKCE)

In stdio mode, when `UPTIME_BEARER_TOKEN` is not set, the server performs a
browser-based OAuth2 **PKCE** flow lazily, on the first tool call rather than at
startup. This keeps the MCP handshake (`initialize`, `tools/list`) fast. It
needs `-uptime-url` and `-client-id`; scope `api/v1` is requested against
`<uptime-url>/o/authorize/` and `<uptime-url>/o/token/`, and tokens are
refreshed in the background.

```bash
uptime-mcp -transport=stdio \
  -uptime-url=https://uptime.com \
  -client-id=<your-client-id>
```

### HTTP (per-request bearer + RFC 9728 discovery)

In HTTP mode the server is a token passthrough. For each request it resolves a
token in this order and forwards it to the Uptime.com API:

| Priority | Source                          |
|----------|---------------------------------|
| 1        | `Authorization: Bearer <token>` |
| 2        | `?token=<token>` query parameter |
| 3        | `UPTIME_BEARER_TOKEN` env var    |

When `-uptime-url` is set, the server also serves
[RFC 9728](https://www.rfc-editor.org/rfc/rfc9728) protected-resource metadata at
`/.well-known/oauth-protected-resource`, advertising the Uptime.com
authorization server and scopes (`api/v1`, `api/v1:read`) so OAuth2-capable MCP
clients can obtain tokens themselves.

```bash
uptime-mcp -transport=http -listen=:8080 \
  -uptime-url=https://uptime.com \
  -client-id=<your-client-id>
```

> Dynamic Client Registration ([RFC 7591](https://www.rfc-editor.org/rfc/rfc7591))
> is planned, so OAuth2 clients will not need a pre-registered client ID.

## HTTP mode and health endpoint

In HTTP mode the server listens on `-listen` (default `:8080`) and exposes:

- `POST /` — the streamable-HTTP MCP endpoint.
- `GET /healthz` — liveness/readiness probe, returns `204 No Content`.
- `GET /.well-known/oauth-protected-resource` — RFC 9728 metadata (only when
  `-uptime-url` is set).

```bash
uptime-mcp -transport=http -listen=:8080 -uptime-url=https://uptime.com
curl -i http://localhost:8080/healthz   # -> HTTP/1.1 204 No Content
```

## Features

The server registers tools across the Uptime.com domains below. Tool names are
stable; use `tools/list` from your MCP client to see full input schemas.

<details>
<summary><b>Checks</b> — list, inspect, and manage monitoring checks</summary>

| Tool                  | Description                                     |
|-----------------------|-------------------------------------------------|
| `list_checks`         | List monitoring checks with optional filtering. |
| `get_check`           | Get details for a specific check.               |
| `get_check_stats`     | Get uptime statistics for a check.              |
| `delete_check`        | Delete a check.                                 |
| `create_<type>_check` | Create a check of a given type (see below).     |
| `update_<type>_check` | Update a check of a given type (see below).     |

Supported check types (`create_*` and `update_*`):
`http`, `dns`, `ssl`, `icmp`, `tcp`, `udp`, `smtp`, `imap`, `pop`, `ssh`,
`ntp`, `whois`, `rdap`, `blacklist`, `malware`, `heartbeat`, `webhook`,
`group`, `pagespeed`, `rum`, `rum2`, `cloudstatus`, `api`, `transaction`.
</details>

<details>
<summary><b>Locations</b></summary>

`list_locations`, `get_location` — discover probe-server locations and their IP
addresses.
</details>

<details>
<summary><b>Contacts</b></summary>

`list_contacts`, `get_contact`, `create_contact`, `update_contact`,
`delete_contact` — manage contact groups used for alert notifications.
</details>

<details>
<summary><b>Tags</b></summary>

`list_tags`, `get_tag`, `create_tag`, `update_tag`, `delete_tag`.
</details>

<details>
<summary><b>Dashboards</b></summary>

`list_dashboards`, `get_dashboard`, `create_dashboard`, `update_dashboard`,
`delete_dashboard`.
</details>

<details>
<summary><b>Status pages, components &amp; incidents</b></summary>

- Pages: `list_status_pages`, `get_status_page`, `create_status_page`,
  `update_status_page`, `delete_status_page`.
- Components: `list_status_page_components`, `get_status_page_component`,
  `create_status_page_component`, `update_status_page_component`,
  `delete_status_page_component`.
- Incidents: `list_status_page_incidents`, `get_status_page_incident`,
  `create_status_page_incident`, `update_status_page_incident`,
  `delete_status_page_incident`.
</details>

<details>
<summary><b>Alerts &amp; outages</b></summary>

`list_alerts`, `get_alert`, `ignore_alert`, `list_outages`, `get_outage`.
</details>

<details>
<summary><b>Cloud status</b></summary>

`list_cloudstatus_providers`, `search_cloudstatus_services` — discover cloud
providers and services for `cloudstatus` checks.
</details>

<details>
<summary><b>Account</b></summary>

`get_account_usage` — account usage and plan limits.
</details>

## Development

Requirements: Go 1.26+.

```bash
make test            # run unit tests
make e2e             # run e2e tests (requires UPTIME_BEARER_TOKEN)
make run/http        # run the HTTP server locally on :8080
go build -o uptime-mcp .
```

The e2e suite talks to a live Uptime.com account and is build-tagged `e2e`; it
runs only when you provide a valid `UPTIME_BEARER_TOKEN`:

```bash
UPTIME_BEARER_TOKEN=<your-api-token> make e2e
```

Mocks are generated with [mockery](https://vektra.github.io/mockery/) v3+
(`.mockery.yaml`).

## Contributing

Contributions are welcome. Please open an issue to discuss substantial changes
before sending a pull request, keep changes focused, and run `make test` before
submitting. By contributing you agree that your contributions are licensed under
the project's MIT license.

## License

Licensed under the [MIT License](LICENSE). `SPDX-License-Identifier: MIT`.
