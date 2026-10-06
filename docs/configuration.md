# Configuration

## MCP access mode

`cully-mcp` defaults to no OAuth. Set `CULLY_MCP_OWNER` to a stable, single-user owner (1–512 bytes, without surrounding whitespace or control characters). Every tool call uses it; the caller cannot choose another owner. The listener defaults to `127.0.0.1` in this mode. Keep access on a loopback, private network or trusted tunnel. No issuer or JWKS request is made. Docker `setup.sh` generates this owner automatically.

Set `CULLY_MCP_AUTH_MODE=oauth` or start `cully-mcp --oauth` for OAuth. In that mode, unset `CULLY_MCP_OWNER` and explicitly set issuer, exact public resource audience and JWKS URL. The server verifies RS256 signature, issuer, audience, expiration, subject and read/write scopes. The `--oauth` flag overrides `CULLY_MCP_AUTH_MODE=none`.

## Go memory services

| Variable | Used by | Meaning |
| --- | --- | --- |
| `CULLY_DATABASE_URL` | Data service | PostgreSQL connection URL; required |
| `CULLY_DB_MAX_CONNS` | Data service | Pool maximum, 2–100; default 8 |
| `CULLY_DATA_API_TOKEN` | Both services | Shared private API bearer token; required for serving |
| `CULLY_DATA_API_URL` | MCP service | Private API base URL; required |
| `CULLY_MEM0_URL` | Data service | Self-hosted REST base URL; optional with matching key |
| `CULLY_MEM0_API_KEY` | Data service | Authenticates data API requests to Mem0; required when URL is set. This is not an embedding-provider key |
| `CULLY_HOST` | Both services | Bind address; MCP defaults to `127.0.0.1` without OAuth, `0.0.0.0` with OAuth; data service defaults to `0.0.0.0` |
| `CULLY_PORT` | Both services | Default 8080 for MCP, 8083 for data |
| `CULLY_MCP_AUTH_MODE` | MCP service | `none` (default) or `oauth` |
| `CULLY_MCP_OWNER` | MCP service | Required fixed owner in no-OAuth mode; forbidden in OAuth mode |
| `CULLY_AUTH_ISSUER` | OAuth MCP service | Required exact token issuer |
| `CULLY_AUTH_RESOURCE` | OAuth MCP service | Required exact public MCP resource URL and token audience |
| `CULLY_JWKS_URL` | OAuth MCP service | Required signing-key endpoint; configure explicitly because issuer paths vary |

MCP Runtime can inject `MCP_AUTH_ISSUER`, `MCP_AUTH_RESOURCE` and `MCP_PATH`. Platform issuer/resource values take precedence over `CULLY_AUTH_ISSUER` and `CULLY_AUTH_RESOURCE` in OAuth mode. `MCP_PATH` selects the internal transport path, default `/mcp`; the public resource may include an external route prefix. None of these variables implicitly enable OAuth. Docker `setup.sh` derives the resource URL from `CULLY_MCP_HOST` and, when it runs MCP Auth, derives the issuer and JWKS URL from `CULLY_AUTH_HOST`. Set an explicit JWKS URL only when an existing authorization server uses a different endpoint. See [MCP OAuth](/oauth) for examples.

Keep the data API token identical on both sides. The MCP workload does not receive database credentials or a Mem0 key. The data service accepts forwarded owner identity only from authenticated service calls.

## Service authentication boundaries

| Connection | Credential | Where it belongs |
| --- | --- | --- |
| Agent to Cully MCP | OAuth bearer when enabled; otherwise a private single-user MCP endpoint | Agent and MCP service |
| Cully MCP to data API | `CULLY_DATA_API_TOKEN`, independent of MCP OAuth | MCP deployment environment (from the Actions secret) and data service environment |
| Data API to Mem0 | `CULLY_MEM0_API_KEY` | Data service and Mem0 only |
| Data API to PostgreSQL | Database credentials in `CULLY_DATABASE_URL` | Data service and PostgreSQL only |

Keep the data API, Mem0 and databases on private networks. The MCP service derives the record owner from the verified OAuth subject or the configured single-user owner, then sends it to the data API. The data API trusts that owner only after checking its service token, so protect that token as access to all owners' records.

For the personal release deployment, set this token in the VM `.env` and the `CULLY_DATA_API_TOKEN` GitHub Actions secret. The workflow injects it into the hosted MCP workload environment through MCP Runtime's deploy API. Mem0's API key stays on the VM and is never sent to the hosted MCP workload.

## Docker Compose self-hosting

The supported [Docker self-hosted stack](/hosting#one-command-full-stack) uses plain PostgreSQL for Cully records and a separate pgvector database for Mem0. `./setup.sh` generates the passwords and service tokens in `~/.cully/config.json` and exports them to Compose. PostgreSQL uses password authentication; its credentials are independent of MCP OAuth. The optional Caddy and MCP Auth services are selected by `--oauth existing|mcp-auth`.

The release-triggered [personal stack](/personal-deployment) uses [`deploy/personal-data.env.example`](https://github.com/mcp-runtime/cully/blob/main/deploy/personal-data.env.example) and creates project-scoped volumes: plain PostgreSQL for Cully and pgvector/PostgreSQL for Mem0. Keep its database password, data API token, Mem0 API key and JWT secret on the VM. Mem0's local embedding model needs no provider key. The `workspace_workspace` network must already exist for the private Caddy data route.

## Local CLI

The [local advisor guide](advisor.md#controls) lists local `CULLY_*` controls for display and advisor behavior. Agent-owned directories still honor `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `CURSOR_CONFIG_DIR`.

Only the documented `CULLY_*` settings are supported.
