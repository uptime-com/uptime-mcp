# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/2.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.18.0] - 2026-10-09

### Added

- In stdio mode without `-client-id`, the browser login registers its own OAuth2 client with the authorization server.

### Deprecated

- `-client-id` (`UPTIME_OAUTH_CLIENT_ID`) and `-client-secret` (`UPTIME_OAUTH_CLIENT_SECRET`) are deprecated and will be removed in a later version.
- The Helm chart's `config.clientId` value is deprecated.

### Fixed

- `list_checks` with `is_paused: false` returns only checks that are not paused.

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
