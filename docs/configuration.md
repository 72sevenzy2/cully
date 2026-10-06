---
title: Configuration reference
description: Find settings for the local advisor, Docker setup or a manual Cully deployment.
---

# Configuration reference

Most people do not need to edit configuration files. The [quickstart](/quickstart) installs Cully and starts its memory stack with generated service credentials. Use this page when you need to change a setting or run the services yourself.

## Local advisor

Run `cully status` to check the integration. The [advisor controls](/advisor#controls) cover display and analysis options such as `CULLY_DISPLAY` and `CULLY_ANALYZE_DISABLE`. Cully honors `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `CURSOR_CONFIG_DIR` when your agent uses a custom configuration directory. Agent setup preserves unrelated user-owned settings.

## Docker setup files

| File | When you need it |
| --- | --- |
| `~/.cully/self-hosted/config/.env` | Editable deployment options. The default laptop setup needs no edits. Add public hostnames and a client secret here for [OAuth setup](/oauth). |
| `~/.cully/config.json` | Generated database passwords, service tokens and single-user owner. Keep it private and back it up with Docker volumes. |
| `~/.cully/self-hosted/config/connectors.json` | Identity-provider connector settings for MCP Auth. |
| `~/.cully/self-hosted/config/.secrets/signing-key.pem` | MCP Auth signing key for OAuth. |

Run `cully setup --prepare` to create editable stack files before starting containers. Run `cully setup codex` for the default private stack, or follow the [OAuth guide](/oauth) before adding `--oauth`.

## Manual MCP service settings

These variables are for deploying `cully-mcp` and `cully-data` yourself. Docker setup supplies them automatically.

| Variable | Service | Purpose |
| --- | --- | --- |
| `CULLY_DATA_API_URL` | MCP | Private data API base URL; required. |
| `CULLY_DATA_API_TOKEN` | MCP and data API | Same private bearer token on both services; required for serving. |
| `CULLY_DATABASE_URL` | Data API | PostgreSQL connection URL; required. |
| `CULLY_MEM0_URL` | Data API | Self-hosted Mem0 base URL for semantic recall. The standard stack sets this. |
| `CULLY_MEM0_API_KEY` | Data API | Mem0 API key. Set it with `CULLY_MEM0_URL` for semantic recall. |
| `CULLY_DB_MAX_CONNS` | Data API | Database pool limit; default 8, allowed 2–100. |
| `CULLY_HOST`, `CULLY_PORT` | Both | Bind address and port. MCP defaults to port 8080; data API to 8083. |

Keep the data API, Mem0 and databases on private networks. The MCP service receives the data API token; it does not need database credentials or the Mem0 key.

### MCP access mode

For one person on a private endpoint, set a stable `CULLY_MCP_OWNER` (1–512 bytes, no surrounding whitespace or control characters). Every request uses that owner. `CULLY_MCP_AUTH_MODE=none` is the default, and the MCP listener defaults to `127.0.0.1` in this mode. Keep it on loopback or a trusted private network.

For per-user sign-in, set `CULLY_MCP_AUTH_MODE=oauth` or start `cully-mcp --oauth`. Unset `CULLY_MCP_OWNER`, and set `CULLY_AUTH_ISSUER`, `CULLY_AUTH_RESOURCE` (the exact public MCP URL) and `CULLY_JWKS_URL`. MCP defaults to binding `0.0.0.0` in this mode; put HTTPS in front of it. The provided [OAuth setup](/oauth) derives these values from hostnames.

On MCP Runtime, `MCP_AUTH_ISSUER` and `MCP_AUTH_RESOURCE` take precedence over the matching `CULLY_AUTH_*` values. `MCP_PATH` changes the internal MCP transport path; it defaults to `/mcp`. These values do not enable OAuth by themselves.
