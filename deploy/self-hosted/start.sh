#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$script_dir"

if [ ! -f .env ]; then
  echo 'Copy .env.example to .env and set the required values first.' >&2
  exit 1
fi
mode=none
agent=""
endpoint=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --oauth)
      [ "$#" -ge 2 ] || { echo '--oauth requires existing or mcp-auth' >&2; exit 2; }
      mode="$2"; shift 2 ;;
    --mcp-url)
      [ "$#" -ge 2 ] || { echo '--mcp-url requires a URL' >&2; exit 2; }
      endpoint="$2"; shift 2 ;;
    claude|codex|cursor)
      [ -z "$agent" ] || { echo 'Choose one agent.' >&2; exit 2; }
      agent="$1"; shift ;;
    *) echo 'Usage: ./start.sh [--oauth existing|mcp-auth] [--mcp-url URL] [claude|codex|cursor]' >&2; exit 2 ;;
  esac
done
case "$mode" in
  none|existing|mcp-auth) ;;
  *) echo 'Choose --oauth existing or --oauth mcp-auth.' >&2; exit 2 ;;
esac
if [ "$mode" != none ] && [ -n "$agent" ] && [ -z "$endpoint" ]; then
  echo 'Agent setup with OAuth requires --mcp-url set to the exact public CULLY_AUTH_RESOURCE.' >&2
  exit 2
fi
if [ "$mode" = mcp-auth ]; then
  [ -f connectors.keycloak.json ] || { echo 'Create connectors.keycloak.json from the example first.' >&2; exit 2; }
  [ -f .secrets/signing-key.pem ] || { echo 'Create .secrets/signing-key.pem first.' >&2; exit 2; }
fi
if [ -n "$agent" ]; then
  if ! command -v cully >/dev/null 2>&1; then
    echo 'Install the Cully CLI first, then rerun this command with the agent name.' >&2
    exit 1
  fi
fi

compose() {
  case "$mode" in
    none) docker compose --env-file .env -f compose.yaml -f compose.mem0.yaml "$@" ;;
    existing) CULLY_CADDYFILE=Caddyfile.existing-auth docker compose --env-file .env -f compose.yaml -f compose.mem0.yaml -f compose.oauth.yaml "$@" ;;
    mcp-auth) CULLY_CADDYFILE=Caddyfile.new-auth docker compose --env-file .env -f compose.yaml -f compose.mem0.yaml -f compose.oauth.yaml -f compose.mcp-auth.yaml "$@" ;;
  esac
}

compose config --quiet
compose up -d --wait db mem0-db
compose --profile ops run --rm migrate
case "$mode" in
  none) compose up -d --build --wait data-api mem0 mcp ;;
  existing) compose up -d --build --wait data-api mem0 mcp caddy ;;
  mcp-auth) compose up -d --build --wait data-api mem0 mcp mcp-auth caddy ;;
esac

if [ "$mode" = none ]; then
  endpoint="http://$(compose port mcp 8080)/mcp"
fi
if [ -n "$endpoint" ]; then
  echo "Cully MCP is configured at $endpoint"
else
  echo 'Cully services are running. Connect an agent to CULLY_AUTH_RESOURCE from .env.'
fi
if [ -n "$agent" ]; then
  if [ "$mode" = none ]; then
    cully install "$agent" --mcp-url "$endpoint"
  else
    cully install "$agent" --mcp-url "$endpoint" --oauth
  fi
  echo "Restart $agent to load the Cully skill and MCP tools."
else
  echo 'Run cully install <agent> --mcp-url URL (add --oauth if enabled) to connect an agent.'
fi
