---
name: monitoring-setup
description: >-
  This skill should be used when the user asks to "set up monitoring",
  "create checks for a domain", "monitor a website", "configure uptime checks",
  "add monitoring for <domain>", or mentions comprehensive monitoring coverage.
  Provides practical check creation constraints and workflow patterns that
  complement the MCP server's general guidelines.
---

# Monitoring setup — operational knowledge

Practical constraints and workflow patterns for creating monitoring checks.
This complements the MCP server instructions (which cover check types and location strategy)
with operational knowledge discovered during real monitoring setups.

## Check type constraints

### Location-based checks (require explicit locations)

HTTP, DNS, ICMP, TCP, UDP, SMTP, IMAP, POP, SSH, NTP — select 3–5 probe locations,
set sensitivity ≥ 2.

### Auto-located checks (locations assigned by server)

SSL, Blacklist, Malware, WHOIS, RDAP — do NOT pass locations. The server assigns them
automatically. Passing locations will cause a validation error.

### Constrained checks

**Page Speed** has unique restrictions:

- Maximum **1 location** (validation error if more)
- Minimum interval **1440 minutes** (1 day)
- Must use `Dedicated-*` location prefix (e.g. `Dedicated-United Kingdom-London`)
- Standard locations will fail — only dedicated probe nodes run Lighthouse

## Parallel batch creation workflow

Maximize throughput by creating checks in dependency-ordered batches:

1. **Tag first** — create a tag (e.g. `example.com`) to group all checks.
2. **Batch 1 — location-based** (all parallel): HTTP, DNS (A/MX/NS), ICMP, TCP.
3. **Batch 2 — auto-located** (all parallel): SSL, Blacklist, Malware, WHOIS, RDAP.
4. **Batch 3 — constrained** (last): Page Speed.

All checks within a batch are independent and can be created in a single parallel tool call.

## DNS layering

For comprehensive DNS coverage, create three checks:

| Record | Target                             | Catches                             |
|--------|------------------------------------|-------------------------------------|
| A      | subdomain (e.g. `www.example.com`) | resolution failures for the service |
| MX     | parent domain (e.g. `example.com`) | mail routing breakage               |
| NS     | parent domain (e.g. `example.com`) | nameserver delegation issues        |

NS breakage cascades into A and MX failures, so it provides the earliest signal.

## Domain vs subdomain rules

| Check type                          | Target                     | Why                                                                |
|-------------------------------------|----------------------------|--------------------------------------------------------------------|
| HTTP, ICMP, SSL, Blacklist, Malware | subdomain or full URL      | checks the actual service endpoint                                 |
| DNS A/AAAA/CNAME                    | subdomain                  | resolves the specific host                                         |
| DNS MX/NS                           | parent (registered) domain | MX and NS are zone-level records                                   |
| WHOIS, RDAP                         | parent (registered) domain | WHOIS/RDAP data exists only for registered domains, not subdomains |

WHOIS and RDAP both require `expect_string` — set it to the domain name being monitored
(e.g. `example.com`). Creating both provides redundancy since WHOIS servers can be unreliable.

## Common pitfalls

- **Page Speed with multiple locations** → "Max 1 locations allowed" — use exactly one `Dedicated-*` location.
- **Page Speed interval < 1440** → "minimum interval for this check type is 1 days" — use 1440 or higher.
- **WHOIS/RDAP on subdomain** → will fail or return no data — always use the registered parent domain.
- **Passing locations to SSL/Blacklist/Malware** → validation error — omit locations entirely.
- **Sensitivity = 1 with many locations** → excessive false positives — use ≥ 2.
