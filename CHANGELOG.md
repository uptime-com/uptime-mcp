# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/2.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.18.0] - 2026-11-18

### Added

- `search_issues` finds existing GitHub issues for this server, and `report_issue` drafts a bug report or feature request, or a comment on a matching issue, for the user to submit.
- The server instructions ask agents to search for an existing issue before reporting a problem, to comment there rather than open a duplicate, and to replace sensitive data in the draft with placeholders.

### Changed

- In stdio mode, `-client-id` (`UPTIME_OAUTH_CLIENT_ID`) defaults to `uptime-mcp`, so the browser login starts without one.

### Removed

- **Breaking:** `UPTIME_BEARER_TOKEN` is no longer read; stdio mode signs in through the browser.
- **Breaking:** HTTP mode no longer reads a token from the `token` query parameter; send it in the `Authorization: Bearer` header.
- **Breaking:** account API keys are no longer accepted; every token must be an Uptime.com OAuth2 access token.
- **Breaking:** the Helm chart's `auth.existingSecret` values are removed.

### Fixed

- `list_checks` with `is_paused: false` returns only checks that are not paused.
- `list_alerts` and `list_outages` report how many results the page holds, rather than "Found 0 results".
- `list_tags` and the tag resource print a tag's color with a single `#`.
- `get_alert`, `list_alerts` and the alert resource show each location reporting the check down, with its server and output.
- `get_check_stats` reports response time in milliseconds rather than printing the API's seconds as milliseconds.

## [0.17.3] - 2026-08-31

### Fixed

- An OAuth2 access token and an account API key are both accepted in every token source, whatever their shape.

## [0.17.2] - 2026-07-25

### Fixed

- An account API key passed as the bearer token is accepted by the Uptime.com API.

## [0.17.1] - 2026-07-01

### Fixed

- HTTP mode no longer answers `404 session not found` when requests of one session reach different replicas.

## [0.17.0] - 2026-07-01

### Added

- `-api-url` (`UPTIME_API_URL`) and `-oauth-url` (`UPTIME_OAUTH_URL`) set the API base URL and the OAuth2 authorization server separately from `-uptime-url`.

## [0.16.0] - 2026-06-15

### Added

- First public release.

[Unreleased]: https://github.com/uptime-com/uptime-mcp/compare/v0.18.0...HEAD
[0.18.0]: https://github.com/uptime-com/uptime-mcp/compare/v0.17.3...v0.18.0
[0.17.3]: https://github.com/uptime-com/uptime-mcp/compare/v0.17.2...v0.17.3
[0.17.2]: https://github.com/uptime-com/uptime-mcp/compare/v0.17.1...v0.17.2
[0.17.1]: https://github.com/uptime-com/uptime-mcp/compare/v0.17.0...v0.17.1
[0.17.0]: https://github.com/uptime-com/uptime-mcp/compare/v0.16.0...v0.17.0
[0.16.0]: https://github.com/uptime-com/uptime-mcp/releases/tag/v0.16.0
