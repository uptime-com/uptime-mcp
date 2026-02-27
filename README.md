# uptime-mcp

MCP server for [Uptime.com](https://uptime.com) monitoring integration.

## Claude Code plugin

Make sure the `UPTIME_API_KEY` environment variable is set in your shell:

```bash
export UPTIME_API_KEY="<your-token>"
```

### Quick start (ad-hoc)

Run Claude Code with the plugin loaded for a single session, without modifying your config:

```bash
claude --plugin-dir /path/to/uptime-mcp
```

### Permanent install

To add the plugin permanently:

```bash
claude plugin add --transport stdio https://github.com/uptime-com/uptime-mcp
```

The plugin automatically downloads the correct binary for your platform on first use.

## Standalone usage

### Stdio mode

For direct integration with any MCP client (Claude Desktop, Cursor, etc.):

```bash
UPTIME_API_TOKEN= <your-token >uptime-mcp -transport=stdio
```

### HTTP mode

Run as an HTTP server with bearer token authentication:

```bash
uptime-mcp -transport=http -listen=:8080
```

Each request must include an `Authorization: Bearer <token>` header. The token is validated against the Uptime.com API
on every request.

### CLI flags

| Flag         | Default                      | Description                                 |
|--------------|------------------------------|---------------------------------------------|
| `-transport` | `stdio`                      | Transport mode: `stdio` or `http`           |
| `-listen`    | `:8080`                      | HTTP listen address (http mode only)        |
| `-api-url`   | `https://uptime.com/api/v1/` | Uptime.com API base URL                     |
| `-log-level` | `info`                       | Log level: `debug`, `info`, `warn`, `error` |
| `-version`   |                              | Print version and exit                      |

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