#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$script_dir"

say() { printf '\n==> %s\n' "$1"; }

mode=none
agent=""
endpoint=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --oauth)
      [ "$#" -ge 2 ] || { echo '--oauth requires mcp-auth' >&2; exit 2; }
      mode="$2"; shift 2 ;;
    claude|codex|cursor)
      [ -z "$agent" ] || { echo 'Choose one agent.' >&2; exit 2; }
      agent="$1"; shift ;;
    *) echo 'Usage: ./setup.sh [--oauth mcp-auth] [claude|codex|cursor]' >&2; exit 2 ;;
  esac
done
case "$mode" in
  none|mcp-auth) ;;
  *) echo 'Choose --oauth mcp-auth.' >&2; exit 2 ;;
esac
command -v docker >/dev/null 2>&1 || { echo 'Docker Compose is required for self-hosting.' >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo 'Docker Compose is required for self-hosting.' >&2; exit 1; }
cli=${CULLY_CLI_BINARY:-cully}
if ! command -v "$cli" >/dev/null 2>&1; then
  if [ -x "${CLAUDE_CONFIG_DIR:-$HOME/.claude}/bin/cully" ]; then
    cli="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/bin/cully"
  elif [ -x "${CODEX_HOME:-$HOME/.codex}/bin/cully" ]; then
    cli="${CODEX_HOME:-$HOME/.codex}/bin/cully"
  elif [ -x "${CURSOR_CONFIG_DIR:-$HOME/.cursor}/bin/cully" ]; then
    cli="${CURSOR_CONFIG_DIR:-$HOME/.cursor}/bin/cully"
  elif [ -x ../../build/cully ]; then
    cli=../../build/cully
  else
    echo 'Install the Cully CLI first, then rerun setup.' >&2
    exit 1
  fi
fi
say 'Generating or reusing private database passwords, API keys and owner ID (values are not printed)'
"$cli" _internal self-hosted-credentials ensure
if [ ! -f .env ]; then
  cp .env.example .env
  chmod 600 .env
fi
exports=$("$cli" _internal self-hosted-credentials export)
eval "$exports"
if [ "$mode" = mcp-auth ]; then
  [ -f connectors.json ] || { echo 'Create connectors.json for your identity provider first.' >&2; exit 2; }
  [ -f .secrets/signing-key.pem ] || { echo 'Create .secrets/signing-key.pem first.' >&2; exit 2; }
fi
if [ "$mode" != none ]; then
  say 'Validating OAuth hostnames and identity-provider connector'
  oauth_exports=$("$cli" _internal self-hosted-credentials oauth-env "$mode")
  eval "$oauth_exports"
  endpoint=$CULLY_AUTH_RESOURCE
fi
compose() {
  case "$mode" in
    none) CULLY_MCP_AUTH_MODE=none docker compose --env-file .env -f compose.yaml "$@" ;;
    mcp-auth) CULLY_MCP_AUTH_MODE=oauth CULLY_MCP_OWNER='' docker compose --env-file .env -f compose.yaml --profile oauth --profile mcp-auth "$@" ;;
  esac
}

say 'Checking Docker Compose configuration'
compose config --quiet
say 'Pulling images if needed and starting PostgreSQL databases'
compose up -d --wait db mem0-db
say 'Applying the Cully database schema'
compose --profile ops run --rm migrate
say 'The Mem0 image includes the local embedding model; its first build downloads and caches that model'
case "$mode" in
  none)
    say 'Building and starting Cully MCP, the private data API and Mem0'
    compose up -d --build --wait data-api mem0 mcp ;;
  mcp-auth)
    say 'Building and starting Cully MCP, the private data API, Mem0, MCP Auth and Caddy'
    compose up -d --build --wait data-api mem0 mcp mcp-auth caddy ;;
esac

if [ "$mode" = none ]; then
  endpoint="http://$(compose port mcp 8080)/mcp"
fi
say 'Local Cully services are ready'
echo "Cully MCP is configured at $endpoint"
if [ -n "$agent" ]; then
  say "Connecting $agent and configuring the Cully skill, hooks and controls; starting the advisor daemon"
  if [ "$mode" = none ]; then
    "$cli" agent setup "$agent" --mcp-url "$endpoint"
  else
    "$cli" agent setup "$agent" --mcp-url "$endpoint" --oauth
  fi
  echo "Setup complete. Restart $agent to load the Cully skill, hooks and MCP tools."
else
  echo 'Run cully agent setup <agent> --mcp-url URL (add --oauth if enabled) to connect an agent.'
fi
