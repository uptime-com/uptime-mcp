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

For [Claude Code](https://claude.com/claude-code), the
[`uptime-skills`](https://github.com/uptime-com/uptime-skills) plugin bundles
the hosted server with skills for Uptime.com work, such as choosing probe
locations, tuning outage sensitivity and picking a check type, and with a
default permission set.

To install the plugin, run these commands in Claude Code:

```text
/plugin marketplace add uptime-com/uptime-skills
/plugin install uptime@uptime-com
```

Then run `/mcp` and sign in to Uptime.com in the browser window it opens.
Claude Code stores and refreshes the token.

For team and project-level setup, see the
[`uptime-skills` README](https://github.com/uptime-com/uptime-skills).

## Hosted server

Uptime.com runs an official hosted instance. Point any streamable-HTTP MCP
client at:

```text
https://mcp.uptime.com/mcp
```

To authenticate with an Uptime.com API token, generate one under
**Settings > API & Integrations** and send it as a bearer token:

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

An MCP client that supports OAuth2 needs no token in its configuration. The
endpoint publishes [RFC 9728](https://www.rfc-editor.org/rfc/rfc9728)
protected-resource metadata naming the authorization server,
`https://uptime.com`, and the client signs in there itself.

## Quick start

To run the server locally with Go 1.26 or later and an Uptime.com API token
from **Settings > API & Integrations**:

```bash
export UPTIME_BEARER_TOKEN=<your-api-token>
go run github.com/uptime-com/uptime-mcp@latest -transport=stdio
```

### Claude Desktop, Claude Code or Cursor with an API token

Add the server to your MCP client configuration, for example
`claude_desktop_config.json`, Cursor's `mcp.json`, or `claude mcp add-json`:

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

Where `uptime-mcp` is not on the client's `PATH`, set `command` to the binary's
absolute path, or run the container image:

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

### Claude Desktop with a browser login

Without a token in the configuration, the server signs you in through the
browser on the first tool call. It registers itself with Uptime.com as an OAuth2
client, so nothing needs setting up in your account first:

```json
{
  "mcpServers": {
    "uptime": {
      "command": "uptime-mcp",
      "args": [
        "-transport=stdio",
        "-uptime-url=https://uptime.com"
      ]
    }
  }
}
```

The server refreshes the token in the background. For details, see
[Authentication](#authentication).

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

The image is published to GitHub Container Registry with two tags:
`:<version>`, which is immutable (for example, `:0.16.0`), and `:latest`, which
follows the newest release.

```bash
docker pull ghcr.io/uptime-com/uptime-mcp:latest
docker run -i --rm -e UPTIME_BEARER_TOKEN ghcr.io/uptime-com/uptime-mcp:latest -transport=stdio
```

### Helm chart (HTTP mode on Kubernetes)

The chart is published as an OCI artifact and carries no Uptime.com URL; set it
at install time:

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

The server reads its configuration from command-line flags. Where a flag is not
set, the environment variable beside it is read instead.

| Flag             | Environment variable         | Default                   | Description                                                         |
|------------------|------------------------------|---------------------------|---------------------------------------------------------------------|
| `-transport`     | None                         | `stdio`                   | Transport mode: `stdio` or `http`.                                  |
| `-listen`        | None                         | `:8080`                   | HTTP listen address (HTTP mode only).                               |
| `-uptime-url`    | `UPTIME_URL`                 | None                      | Uptime.com instance URL, for example `https://uptime.com`. Required for the stdio browser login and for the protected-resource metadata. The API base is `<uptime-url>/api/v1/`, or `https://uptime.com/api/v1/` when unset. |
| `-api-url`       | `UPTIME_API_URL`             | _(from `-uptime-url`)_    | Full API base URL override, used verbatim, for example `http://uptime.svc.cluster.local/api/v1/`. |
| `-oauth-url`     | `UPTIME_OAUTH_URL`           | _(from `-uptime-url`)_    | Full OAuth2 authorization server URL override, used verbatim as the issuer. |
| `-resource-url`  | `UPTIME_RESOURCE_URL`        | `http://localhost:<port>` | Public URL of this server, for OAuth2 protected-resource metadata.  |
| `-client-id`     | `UPTIME_OAUTH_CLIENT_ID`     | None                      | **Deprecated.** Pre-registered OAuth2 client ID for the stdio browser login; without it the server registers a client itself. |
| `-client-secret` | `UPTIME_OAUTH_CLIENT_SECRET` | None                      | **Deprecated.** Secret of the `-client-id` client (confidential clients). |
| `-log-level`     | None                         | `error`                   | Log level: `debug`, `info`, `warn`, `error`.                        |
| `-version`       | None                         | None                      | Print version and commit, then exit.                               |

One variable has no flag:

| Variable              | Description                                                               |
|-----------------------|---------------------------------------------------------------------------|
| `UPTIME_BEARER_TOKEN` | Static Uptime.com API token. Forwarded as-is, without verification or refresh. |

## Authentication

### Bearer token

Set `UPTIME_BEARER_TOKEN` to an Uptime.com API token. It works in both stdio and
HTTP modes. The server forwards the token to the Uptime.com API as it is, with
no OAuth2 configuration, verification or refresh.

```bash
UPTIME_BEARER_TOKEN=<your-api-token> uptime-mcp -transport=stdio
```

### Browser login (stdio)

In stdio mode without `UPTIME_BEARER_TOKEN`, the server runs an OAuth2
authorization code flow with PKCE in the browser. The flow starts on the first
tool call, not at startup, so the MCP handshake (`initialize`, `tools/list`)
never waits on a browser. It needs `-uptime-url`.

The server reads the authorization server's
[RFC 8414](https://www.rfc-editor.org/rfc/rfc8414) metadata at
`<uptime-url>/.well-known/oauth-authorization-server` and registers a public
client for its local callback ([RFC 7591](https://www.rfc-editor.org/rfc/rfc7591)).
It requests scope `api/v1` and refreshes the token in the background. The
registration is not kept: each server process registers again on its first
login.

```bash
uptime-mcp -transport=stdio -uptime-url=https://uptime.com
```

The deprecated `-client-id` (with `-client-secret` for a confidential client)
skips registration and uses that client against `<uptime-url>/o/authorize/` and
`<uptime-url>/o/token/`.

### HTTP mode

In HTTP mode the server passes each request's token through to the Uptime.com
API. It takes the token from the first of these sources that has one:

| Priority | Source                          |
|----------|---------------------------------|
| 1        | `Authorization: Bearer <token>` |
| 2        | `?token=<token>` query parameter |
| 3        | `UPTIME_BEARER_TOKEN` env var    |

Each source accepts an OAuth2 access token or an account API key. The Uptime.com
API expects the first as `Authorization: Bearer` and the second as
`Authorization: Token`, and rejects either under the other scheme. The server
sends the scheme the source implies and switches once if the API refuses it, so
there is nothing to configure.

When `-uptime-url` is set, the server also serves
[RFC 9728](https://www.rfc-editor.org/rfc/rfc9728) protected-resource metadata at
`/.well-known/oauth-protected-resource`, advertising the Uptime.com
authorization server and the scopes `api/v1` and `api/v1:read`. An MCP client
that supports OAuth2 registers with that authorization server and obtains its
own tokens; this server needs no client ID for it.

```bash
uptime-mcp -transport=http -listen=:8080 -uptime-url=https://uptime.com
```

## HTTP endpoints

In HTTP mode the server listens on `-listen` (default `:8080`) and serves these
endpoints:

| Endpoint                                     | Purpose                                                              |
|----------------------------------------------|----------------------------------------------------------------------|
| `POST /`                                     | The streamable-HTTP MCP endpoint.                                    |
| `GET /healthz`                               | Liveness and readiness probe. Returns `204 No Content`.              |
| `GET /.well-known/oauth-protected-resource`  | RFC 9728 metadata. Served only when `-uptime-url` is set.            |

To check a running server:

```bash
curl -i http://localhost:8080/healthz
```

The output starts with:

```text
HTTP/1.1 204 No Content
```

## Features

The server registers the tools below. For the full input schemas, call
`tools/list` from your MCP client.

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

To regenerate the mocks after a change to the Uptime.com client interface, run
[mockery](https://vektra.github.io/mockery/) v3 or later from the repository
root. It reads `.mockery.yaml`:

```bash
mockery
```

## Contributing

Contributions are welcome. Open an issue to discuss a substantial change before
sending a pull request, keep each change focused, and run `make test` before
submitting. A change that users can notice adds a line under `## [Unreleased]`
in [`CHANGELOG.md`](CHANGELOG.md). By contributing, you agree that your
contributions are licensed under the project's MIT license.

## License

Licensed under the [MIT License](LICENSE). `SPDX-License-Identifier: MIT`.
