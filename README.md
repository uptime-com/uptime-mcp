# uptime-mcp

MCP server for [Uptime.com](https://uptime.com) monitoring integration.

## Authentication

Multiple authentication methods are available, from simplest to most complete.

### Bearer token (simplest)

Set the `UPTIME_BEARER_TOKEN` environment variable to a pre-obtained Uptime.com API token.
Works with both stdio and HTTP modes. No OAuth2 configuration required — the token is used
as-is with no verification or refresh.

**Stdio mode:**

```bash
UPTIME_BEARER_TOKEN=your-token uptime-mcp -transport=stdio
```

**HTTP mode:**

```bash
UPTIME_BEARER_TOKEN=your-token uptime-mcp -transport=http -listen=:8080
```

### OAuth2 (stdio mode)

In stdio mode, the server performs a browser-based OAuth2 PKCE flow on first tool
call. Requires `-uptime-url` and `-client-id`. The MCP handshake completes without
auth; on the first actual tool call the server opens your browser to complete
authorization, then automatically refreshes tokens in the background.

```bash
uptime-mcp -transport=stdio \
  -uptime-url=https://sandbox.upeks.net \
  -client-id=your-client-id
```

### OAuth2 (HTTP mode)

In HTTP mode with `-client-id` set, the server exposes
`/.well-known/oauth-protected-resource`
([RFC 9728](https://www.rfc-editor.org/rfc/rfc9728)) so MCP clients can discover
the authorization server and perform the OAuth2 flow themselves. Bearer tokens
received from the client are forwarded to the Uptime.com API as-is.

```bash
uptime-mcp -transport=http -listen=:8080 \
  -uptime-url=https://sandbox.upeks.net \
  -client-id=your-client-id
```

### Authentication precedence

**Stdio mode** — token resolved once on first tool call:

| Priority | Source                        | Behavior                                                             |
|----------|-------------------------------|----------------------------------------------------------------------|
| 1        | `UPTIME_BEARER_TOKEN` env var | Static token, no browser, no refresh                                 |
| 2        | OAuth2 PKCE flow              | Requires `-uptime-url` + `-client-id`, opens browser, auto-refreshes |

**HTTP mode** — token resolved per-request:

| Priority | Source                         | Behavior                     |
|----------|--------------------------------|------------------------------|
| 1        | `Authorization: Bearer` header | Forwarded to Uptime.com API  |
| 2        | `?token=` query parameter      | Forwarded to Uptime.com API  |
| 3        | `UPTIME_BEARER_TOKEN` env var  | Forwarded to Uptime.com API  |

## Usage

### Stdio mode

For direct integration with any MCP client (Claude Desktop, Cursor, etc.):

```bash
uptime-mcp -transport=stdio \
  -uptime-url=https://uptime.com \
  -client-id=your-client-id
```

See [Authentication](#authentication) for all available auth methods.

### HTTP mode

Run as an HTTP server:

```bash
uptime-mcp -transport=http -listen=:8080 \
  -uptime-url=https://uptime.com \
  -client-id=your-client-id
```

See [Authentication](#authentication) for token resolution order and per-request token options.

### CLI flags

| Flag                  | Default                   | Description                                                 |
|-----------------------|---------------------------|-------------------------------------------------------------|
| `-transport`          | `stdio`                   | Transport mode: `stdio` or `http`                           |
| `-listen`             | `:8080`                   | HTTP listen address (http mode only)                        |
| `-uptime-url`         |                           | Uptime.com instance URL (e.g., `https://uptime.com`)        |
| `-resource-url`       | `http://localhost:{port}` | Public URL of this server (for reverse proxy setups)        |
| `-client-id`          |                           | OAuth2 client ID                                            |
| `-client-secret`      |                           | OAuth2 client secret (confidential clients)                 |
| `-log-level`          | `info`                    | Log level: `debug`, `info`, `warn`, `error`                 |
| `-version`            |                           | Print version and exit                                      |

## Tools

### Checks

| Tool                       | Description                                     |
|----------------------------|-------------------------------------------------|
| `list_checks`              | List monitoring checks with optional filtering  |
| `get_check`                | Get detailed information about a specific check |
| `get_check_stats`          | Get uptime statistics for a check               |
| `delete_check`             | Delete a monitoring check                       |
| `create_<type>_check`      | Create a check (see types below)                |
| `update_<type>_check`      | Update a check (see types below)                |

Supported check types: `http`, `dns`, `ssl`, `icmp`, `tcp`, `udp`, `smtp`, `imap`,
`pop`, `ssh`, `ntp`, `whois`, `rdap`, `blacklist`, `malware`, `heartbeat`, `webhook`,
`group`, `pagespeed`, `rum`.

### Locations

| Tool             | Description                                 |
|------------------|---------------------------------------------|
| `list_locations` | List available probe server locations       |
| `get_location`   | Get location details including IP addresses |

### Contacts

| Tool              | Description                                 |
|-------------------|---------------------------------------------|
| `list_contacts`   | List contact groups for alert notifications |
| `get_contact`     | Get contact group details                   |
| `create_contact`  | Create a new contact group                  |
| `update_contact`  | Update a contact group                      |
| `delete_contact`  | Delete a contact group                      |

### Tags

| Tool         | Description            |
|--------------|------------------------|
| `list_tags`  | List tags              |
| `get_tag`    | Get tag details        |
| `create_tag` | Create a new tag       |
| `update_tag` | Update an existing tag |
| `delete_tag` | Delete a tag           |

### Dashboards

| Tool               | Description          |
|--------------------|----------------------|
| `list_dashboards`  | List dashboards      |
| `get_dashboard`    | Get dashboard details|
| `create_dashboard` | Create a dashboard   |
| `update_dashboard` | Update a dashboard   |
| `delete_dashboard` | Delete a dashboard   |

### Status Pages

| Tool                             | Description                       |
|----------------------------------|-----------------------------------|
| `list_status_pages`              | List status pages                 |
| `get_status_page`                | Get status page details           |
| `create_status_page`             | Create a status page              |
| `update_status_page`             | Update a status page              |
| `delete_status_page`             | Delete a status page              |
| `list_status_page_components`    | List components on a status page  |
| `get_status_page_component`      | Get component details             |
| `create_status_page_component`   | Create a status page component    |
| `update_status_page_component`   | Update a status page component    |
| `delete_status_page_component`   | Delete a status page component    |
| `list_status_page_incidents`     | List incidents on a status page   |
| `get_status_page_incident`       | Get incident details              |
| `create_status_page_incident`    | Create a status page incident     |
| `update_status_page_incident`    | Update a status page incident     |
| `delete_status_page_incident`    | Delete a status page incident     |

### Alerts & Outages

| Tool           | Description        |
|----------------|--------------------|
| `list_alerts`  | List alerts        |
| `get_alert`    | Get alert details  |
| `ignore_alert` | Ignore an alert    |
| `list_outages` | List outages       |
| `get_outage`   | Get outage details |
