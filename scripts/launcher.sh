#!/usr/bin/env bash
set -euo pipefail

# Resolve the plugin root directory (parent of scripts/)
PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"

# --- Platform detection ---

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)      echo "ERROR: unsupported OS: $OS" >&2; exit 1 ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)       ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)            echo "ERROR: unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# --- Version detection ---

if [ -n "${UPTIME_MCP_VERSION:-}" ]; then
  VERSION="$UPTIME_MCP_VERSION"
else
  VERSION="$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$PLUGIN_ROOT/.claude-plugin/plugin.json")"
  if [ -z "$VERSION" ]; then
    echo "ERROR: could not determine version from plugin.json" >&2
    exit 1
  fi
fi

# --- Cache management ---

CACHE_DIR="${PLUGIN_ROOT}/.cache/${VERSION}"
BINARY="${CACHE_DIR}/uptime-mcp"

if [ -x "$BINARY" ]; then
  exec "$BINARY" -transport=stdio "$@"
fi

# --- Download ---

if ! command -v curl >/dev/null 2>&1; then
  echo "ERROR: curl is required but not found in PATH" >&2
  exit 1
fi

ARCHIVE="uptime-mcp_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/uptime-com/uptime-mcp/releases/download/v${VERSION}/${ARCHIVE}"

mkdir -p "$CACHE_DIR"

TMPFILE="$(mktemp "${CACHE_DIR}/${ARCHIVE}.XXXXXX")"
trap 'rm -f "$TMPFILE"' EXIT

echo "Downloading uptime-mcp v${VERSION} (${OS}/${ARCH})..." >&2
if ! curl -fsSL -o "$TMPFILE" "$URL"; then
  echo "ERROR: failed to download ${URL}" >&2
  echo "Check that version v${VERSION} exists and has a release for ${OS}/${ARCH}" >&2
  exit 1
fi

tar xzf "$TMPFILE" -C "$CACHE_DIR" uptime-mcp
chmod +x "$BINARY"
rm -f "$TMPFILE"
trap - EXIT

echo "Cached uptime-mcp v${VERSION} at ${BINARY}" >&2

exec "$BINARY" -transport=stdio "$@"
