package server

const instructions = `Uptime.com monitoring: checks, contacts, tags, alerts, outages, dashboards, status pages and account usage.

## Setting up monitoring for a domain

1. Call list_locations for the probe locations and list_contacts for the contact groups that receive alerts.
2. Plan the checks from the services the domain runs, using the coverage below.
3. Create the most critical checks first: HTTP, DNS, SSL.

## Locations

Give each check at least 3 locations, or 5 for faster outage confirmation. Place them where the service's users are, and always include at least one outside that region to catch routing and CDN faults:

- US audience: 3 US regions (East, West, Central), plus 1-2 in the EU or Southeast Asia.
- EU audience: 3 EU countries, plus 1 US and 1 Southeast Asia location.
- Global audience: spread evenly across the US, the EU and Asia-Pacific.

## Sensitivity

Set sensitivity, the number of locations that must confirm an outage, to 2 or more, so one location's network fault raises no alert.

## Coverage for a domain

Beyond HTTP, add:

- DNS: the A and AAAA records, which catch delegation and propagation faults.
- SSL: certificate expiry.
- ICMP: a network-layer reachability baseline.
- WHOIS or RDAP: domain registration expiry, with threshold set to the days of warning wanted.
- Email, when the domain receives mail: a DNS check with dns_record_type MX, and SMTP, IMAP or POP checks for the mail servers.
- TCP, SSH, UDP or NTP: for services the domain runs beyond the web.
- Blacklist and Malware: listing on spam blacklists and on Google Safe Browsing.
`
