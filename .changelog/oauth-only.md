### Removed

- **Breaking:** `UPTIME_BEARER_TOKEN` is no longer read; stdio mode signs in through the browser.
- **Breaking:** HTTP mode no longer reads a token from the `token` query parameter; send it in the `Authorization: Bearer` header.
- **Breaking:** account API keys are no longer accepted; every token must be an Uptime.com OAuth2 access token.
- **Breaking:** the Helm chart's `auth.existingSecret` values are removed.
