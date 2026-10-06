#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$script_dir"

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
command -v python3 >/dev/null 2>&1 || { echo 'Python 3 is required for local credential setup.' >&2; exit 1; }
command -v docker >/dev/null 2>&1 || { echo 'Docker Compose is required for self-hosting.' >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo 'Docker Compose is required for self-hosting.' >&2; exit 1; }
python3 ./credentials.py ensure
if [ ! -f .env ]; then
  cp .env.example .env
  chmod 600 .env
fi
exports=$(python3 ./credentials.py export)
eval "$exports"
if [ "$mode" = mcp-auth ]; then
  [ -f connectors.json ] || { echo 'Create connectors.json for your identity provider first.' >&2; exit 2; }
  [ -f .secrets/signing-key.pem ] || { echo 'Create .secrets/signing-key.pem first.' >&2; exit 2; }
fi
if [ "$mode" != none ]; then
  oauth_exports=$(python3 ./credentials.py oauth-env "$mode")
  eval "$oauth_exports"
  endpoint=$CULLY_AUTH_RESOURCE
fi
if [ -n "$agent" ]; then
  cli=cully
  if ! command -v "$cli" >/dev/null 2>&1; then
    if [ -x "${CLAUDE_CONFIG_DIR:-$HOME/.claude}/bin/cully" ]; then
      cli="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/bin/cully"
    elif [ -x ../../build/cully ]; then
      cli=../../build/cully
    else
      echo 'Install the Cully CLI first, then rerun this command with the agent name.' >&2
      exit 1
    fi
  fi
fi

compose() {
  case "$mode" in
    none) CULLY_MCP_AUTH_MODE=none docker compose --env-file .env -f compose.yaml "$@" ;;
    mcp-auth) CULLY_MCP_AUTH_MODE=oauth CULLY_MCP_OWNER='' docker compose --env-file .env -f compose.yaml --profile oauth --profile mcp-auth "$@" ;;
  esac
}

compose config --quiet
compose up -d --wait db mem0-db
compose --profile ops run --rm migrate
case "$mode" in
  none) compose up -d --build --wait data-api mem0 mcp ;;
  mcp-auth) compose up -d --build --wait data-api mem0 mcp mcp-auth caddy ;;
esac

if [ "$mode" = none ]; then
  endpoint="http://$(compose port mcp 8080)/mcp"
fi
echo "Cully MCP is configured at $endpoint"
if [ -n "$agent" ]; then
  if [ "$mode" = none ]; then
    "$cli" agent setup "$agent" --mcp-url "$endpoint"
  else
    "$cli" agent setup "$agent" --mcp-url "$endpoint" --oauth
  fi
  echo "Restart $agent to load the Cully skill and MCP tools."
else
  echo 'Run cully agent setup <agent> --mcp-url URL (add --oauth if enabled) to connect an agent.'
fi
