#!/bin/sh
# Cully installer — downloads a prebuilt binary, then self-registers it for
# detected coding agents. Use --from-source to explicitly build with Go.
#
#   curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex --mcp-url http://127.0.0.1:8080/mcp
#
# Env overrides: CULLY_VERSION (e.g. v0.1.0), CLAUDE_CONFIG_DIR, CODEX_HOME,
# CURSOR_CONFIG_DIR. Add --oauth only when the MCP server requires OAuth.
set -eu

REPO="mcp-runtime/cully"
CLAUDE_DIR="${CLAUDE_CONFIG_DIR:-$HOME/.claude}"
BIN_DIR="$CLAUDE_DIR/bin"

agent=""
mcp_url=""
oauth=false
from_source=false
while [ "$#" -gt 0 ]; do
  case "$1" in
    --agent)
      [ "$#" -ge 2 ] || { echo '--agent requires a value' >&2; exit 2; }
      agent="$2"; shift 2 ;;
    --mcp-url)
      [ "$#" -ge 2 ] || { echo '--mcp-url requires a value' >&2; exit 2; }
      mcp_url="$2"; shift 2 ;;
    --oauth) oauth=true; shift ;;
    --from-source) from_source=true; shift ;;
    *) echo "unknown installer option: $1" >&2; exit 2 ;;
  esac
done
[ "$oauth" = false ] || [ -n "$mcp_url" ] || { echo '--oauth requires --mcp-url' >&2; exit 2; }
case "$agent" in
  ""|claude|codex|cursor|all) ;;
  *) echo 'choose --agent claude, codex, cursor, or all' >&2; exit 2 ;;
esac

die() { printf '\033[31mx\033[0m %s\n' "$1" >&2; exit 1; }
say() { printf '\033[36m==>\033[0m %s\n' "$1"; }

command -v curl >/dev/null 2>&1 || die "curl is required."
command -v tar  >/dev/null 2>&1 || die "tar is required."

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
raw_os="$os"
raw_arch="$arch"
case "$arch" in
  x86_64|amd64)  arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) die "unsupported arch: $raw_arch (supported: amd64, arm64)" ;;
esac
case "$os" in
  darwin|linux) ;;
  *) die "unsupported OS: $raw_os (supported: darwin, linux)" ;;
esac

ver="${CULLY_VERSION:-latest}"
asset="cully_${os}_${arch}.tar.gz"
if [ "$ver" = "latest" ]; then
  url="https://github.com/$REPO/releases/latest/download/$asset"
else
  url="https://github.com/$REPO/releases/download/$ver/$asset"
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
say "Detected platform: $raw_os/$raw_arch -> $os/$arch"
if [ "$from_source" = true ]; then
  command -v go >/dev/null 2>&1 || die "Source installation requires Go 1.26 or newer."
  ref="$ver"
  [ "$ref" != latest ] || ref=main
  say "Building Cully from $ref; downloading the Go toolchain and dependencies can take several minutes"
  GOBIN="$tmp" go install -v "github.com/mcp-runtime/cully/cmd/cully@$ref" || die "source install failed for $ref"
else
  say "Downloading $asset ($ver)"
  if curl -fL --progress-bar --connect-timeout 10 --max-time 120 --speed-limit 1024 --speed-time 30 "$url" -o "$tmp/c.tar.gz"; then
    sums_url="$(dirname "$url")/checksums.txt"
    say "Verifying download"
    if curl -fsSL --connect-timeout 10 --max-time 30 "$sums_url" -o "$tmp/checksums.txt"; then
      expected="$(awk -v asset="$asset" '$2 == asset {print $1; exit}' "$tmp/checksums.txt")"
      if [ -n "$expected" ]; then
        if command -v sha256sum >/dev/null 2>&1; then
          actual="$(sha256sum "$tmp/c.tar.gz" | awk '{print $1}')"
        else
          actual="$(shasum -a 256 "$tmp/c.tar.gz" | awk '{print $1}')"
        fi
        [ "$actual" = "$expected" ] || die "checksum mismatch for $asset (expected $expected, got $actual) — aborting install"
      else
        say "warning: no checksum entry for $asset — skipping verification"
      fi
    else
      say "warning: could not fetch checksums.txt — skipping verification"
    fi

    tar -xzf "$tmp/c.tar.gz" -C "$tmp" || die "extract failed"
    [ -f "$tmp/cully" ] || die "archive did not contain the cully binary"
  else
    die "Could not download $asset ($ver). Check your connection and the release at https://github.com/$REPO/releases, then retry. To build with Go instead, rerun with: sh -s -- --from-source"
  fi
fi
[ -f "$tmp/cully" ] || die "installer did not produce the cully binary"

tmp_bin="$tmp/cully"
old_ver=""
if [ -x "$BIN_DIR/cully" ]; then
  old_ver="$("$BIN_DIR/cully" version 2>/dev/null || true)"
  new_ver="$("$tmp_bin" version 2>/dev/null || true)"
  if [ -n "$old_ver" ] && [ "$old_ver" = "$new_ver" ]; then
    say "Already on $new_ver"
  fi
  # A running advisor daemon holds the OLD binary's code in memory; installing
  # over it silently leaves the stale version running. Stop it first so
  # `cully agent setup` (below) starts the new binary fresh.
  say "Stopping any running advisor daemon before upgrade"
  "$BIN_DIR/cully" daemon stop >/dev/null 2>&1 || true
fi

mkdir -p "$BIN_DIR"
install -m 0755 "$tmp_bin" "$BIN_DIR/cully"
# Clear macOS Gatekeeper quarantine on the downloaded binary.
[ "$os" = "darwin" ] && xattr -d com.apple.quarantine "$BIN_DIR/cully" 2>/dev/null || true

say "Installed binary -> $BIN_DIR/cully ($("$BIN_DIR/cully" version 2>/dev/null || echo "$ver"))"
if ! command -v cully >/dev/null 2>&1; then
  say "For later CLI commands, add $BIN_DIR to PATH or use $BIN_DIR/cully"
fi
set --
[ -z "$agent" ] || set -- "$@" "$agent"
[ -z "$mcp_url" ] || set -- "$@" --mcp-url "$mcp_url"
[ "$oauth" = false ] || set -- "$@" --oauth
say "Registering coding agents${mcp_url:+ and MCP endpoint}"
if "$BIN_DIR/cully" help | grep -q 'cully agent setup'; then
  "$BIN_DIR/cully" agent setup "$@"
else
  "$BIN_DIR/cully" setup "$@"
fi
