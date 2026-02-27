package server

const instructions = `Uptime.com monitoring MCP server. Use the provided tools to manage checks, contacts, tags, alerts, and outages.

## Monitoring setup guidelines

### Discovery workflow

When asked to set up monitoring for a domain:

1. Identify the domain from the provided URL or address.
2. Use list_locations to discover available probe locations.
3. Use list_contacts to discover contact groups for alert notifications.
4. Plan checks based on the domain's services and the guidance below.
5. Create checks, starting with the most critical (HTTP, DNS, SSL).

### Location selection

Use list_locations to discover available probe locations. When setting up checks, choose locations based on the target audience:

- Baseline: at least 3 locations per check for reliable outage detection.
- Thorough monitoring: 5 locations for higher confidence and faster incident confirmation.

Distribute locations according to where the monitored service's users are:

- US-focused: pick 3 US locations across different regions (e.g. East, West, Central), plus 1-2 outside the US (EU, SE Asia) for global coverage.
- EU-focused: pick 3 EU locations across different countries, plus 1 US and 1 SE Asia location.
- Global: spread locations evenly across US, EU, and Asia-Pacific.

Always include at least one location outside the primary audience region to detect routing or CDN issues.

### Sensitivity

Set sensitivity (number of locations that must confirm an outage) to at least 2 to avoid false positives from single-location network issues.

### Check types to consider

When monitoring a domain, consider these check types beyond basic HTTP:

- DNS: verify A/AAAA records resolve correctly. Use dns_record_type to check specific record types (A, AAAA, CNAME, MX, NS, TXT). Essential for detecting delegation or propagation issues.
- SSL certificate: monitor expiry dates to prevent certificate lapses. Locations are assigned automatically by the server.
- ICMP: basic reachability check, useful as a network-layer baseline.
- TCP: port connectivity checks for non-HTTP services (databases, custom protocols).
- Email monitoring: if the domain handles email, create a DNS check with dns_record_type set to MX to verify mail exchange records. Consider SMTP, IMAP, or POP checks for mail server availability.
- WHOIS/RDAP: monitor domain registration expiry to prevent accidental domain lapses. Use threshold to set how many days before expiry to alert. Locations are assigned automatically.
- Blacklist: check if a domain or IP appears on spam/abuse blacklists. Locations are assigned automatically.
- Malware: scan a website against the Google Safe Browsing list for malware and viruses. Locations are assigned automatically.
- SSH: verify SSH server connectivity on a specific port.
- UDP/NTP: monitor UDP services and NTP time servers.
- Heartbeat/Webhook: for services that push status. The server generates a unique URL for your service to ping (heartbeat) or post status updates to (webhook).
- Group: aggregate multiple existing checks into a single logical check.
- Page Speed: measure website performance using Google Lighthouse.
`
