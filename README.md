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

In stdio mode, the server performs a browser-based OAuth2 PKCE flow on startup.
Requires `-uptime-url` and `-client-id`. On launch, the server opens your browser to
complete authorization, then automatically refreshes tokens in the background.

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

**Stdio mode** — token resolved once at startup:

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
  -uptime-url=https://sandbox.upeks.net \
  -client-id=your-client-id
```

See [Authentication](#authentication) for all available auth methods.

### HTTP mode

Run as an HTTP server:

```bash
uptime-mcp -transport=http -listen=:8080 \
  -uptime-url=https://sandbox.upeks.net \
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

| Tool                     | Description                                     |
|--------------------------|-------------------------------------------------|
| `list_checks`            | List monitoring checks with optional filtering  |
| `get_check`              | Get detailed information about a specific check |
| `get_check_stats`        | Get uptime statistics for a check               |
| `delete_check`           | Delete a monitoring check                       |
| `create_http_check`      | Create an HTTP/HTTPS monitoring check           |
| `create_dns_check`       | Create a DNS record monitoring check            |
| `create_ssl_check`       | Create an SSL certificate expiry check          |
| `create_icmp_check`      | Create an ICMP (ping) check                     |
| `create_tcp_check`       | Create a TCP port connectivity check            |
| `create_udp_check`       | Create a UDP service check                      |
| `create_smtp_check`      | Create an SMTP mail server check                |
| `create_imap_check`      | Create an IMAP mail server check                |
| `create_pop_check`       | Create a POP3 mail server check                 |
| `create_ssh_check`       | Create an SSH connectivity check                |
| `create_ntp_check`       | Create an NTP time server check                 |
| `create_dns_check`       | Create a DNS monitoring check                   |
| `create_whois_check`     | Create a WHOIS domain expiry check              |
| `create_rdap_check`      | Create an RDAP domain expiry check              |
| `create_blacklist_check` | Create a blacklist monitoring check             |
| `create_malware_check`   | Create a malware scanning check                 |
| `create_heartbeat_check` | Create a heartbeat (push) check                 |
| `create_webhook_check`   | Create a webhook (push) check                   |
| `create_group_check`     | Create a group check aggregating other checks   |
| `create_pagespeed_check` | Create a Lighthouse page speed check            |
| `create_rum_check`       | Create a Real User Monitoring check             |
| `create_rum2_check`      | Create a RUM v2 check                           |

### Locations

| Tool             | Description                                 |
|------------------|---------------------------------------------|
| `list_locations` | List available probe server locations       |
| `get_location`   | Get location details including IP addresses |

### Contacts

| Tool             | Description                                 |
|------------------|---------------------------------------------|
| `list_contacts`  | List contact groups for alert notifications |
| `get_contact`    | Get contact group details                   |
| `create_contact` | Create a new contact group                  |
| `delete_contact` | Delete a contact group                      |

### Tags

| Tool         | Description            |
|--------------|------------------------|
| `list_tags`  | List tags              |
| `get_tag`    | Get tag details        |
| `create_tag` | Create a new tag       |
| `update_tag` | Update an existing tag |

### Alerts & Outages

| Tool           | Description        |
|----------------|--------------------|
| `list_alerts`  | List alerts        |
| `get_alert`    | Get alert details  |
| `ignore_alert` | Ignore an alert    |
| `list_outages` | List outages       |
| `get_outage`   | Get outage details |
