# Configuration

## MCP access mode

`cully-mcp` defaults to no OAuth. Set `CULLY_MCP_OWNER` to a stable, single-user owner (1–512 bytes, without surrounding whitespace or control characters). Every tool call uses it; the caller cannot choose another owner. The listener defaults to `127.0.0.1` in this mode. Keep access on a loopback, private network or trusted tunnel. No issuer or JWKS request is made.

Set `CULLY_MCP_AUTH_MODE=oauth` or start `cully-mcp --oauth` for OAuth. In that mode, unset `CULLY_MCP_OWNER` and explicitly set issuer, exact public resource audience and JWKS URL. The server verifies RS256 signature, issuer, audience, expiration, subject and read/write scopes. The `--oauth` flag overrides `CULLY_MCP_AUTH_MODE=none`.

## Go memory services

| Variable | Used by | Meaning |
| --- | --- | --- |
| `CULLY_DATABASE_URL` | Data service | PostgreSQL connection URL; required |
| `CULLY_DB_MAX_CONNS` | Data service | Pool maximum, 2–100; default 8 |
| `CULLY_DATA_API_TOKEN` | Both services | Shared private API bearer token; required for serving |
| `CULLY_DATA_API_URL` | MCP service | Private API base URL; required |
| `CULLY_MEM0_URL` | Data service | Self-hosted REST base URL; optional with matching key |
| `CULLY_MEM0_API_KEY` | Data service | Mem0 service API key; required when URL is set |
| `CULLY_HOST` | Both services | Bind address; MCP defaults to `127.0.0.1` without OAuth, `0.0.0.0` with OAuth; data service defaults to `0.0.0.0` |
| `CULLY_PORT` | Both services | Default 8080 for MCP, 8083 for data |
| `CULLY_MCP_AUTH_MODE` | MCP service | `none` (default) or `oauth` |
| `CULLY_MCP_OWNER` | MCP service | Required fixed owner in no-OAuth mode; forbidden in OAuth mode |
| `CULLY_AUTH_ISSUER` | OAuth MCP service | Required exact token issuer |
| `CULLY_AUTH_RESOURCE` | OAuth MCP service | Required exact public MCP resource URL and token audience |
| `CULLY_JWKS_URL` | OAuth MCP service | Required signing-key endpoint; configure explicitly because issuer paths vary |

MCP Runtime can inject `MCP_AUTH_ISSUER`, `MCP_AUTH_RESOURCE` and `MCP_PATH`. Platform issuer/resource values take precedence over `CULLY_AUTH_ISSUER` and `CULLY_AUTH_RESOURCE` in OAuth mode. `MCP_PATH` selects the internal transport path, default `/mcp`; the public resource may include an external route prefix. None of these variables implicitly enable OAuth. See [OAuth deployment](/oauth) for MCP Auth and existing-server examples.

Keep the data API token identical on both sides. The MCP workload does not receive database credentials or a Mem0 key. The data service accepts forwarded owner identity only from authenticated service calls.

## VM Compose

`.env.example` lists `CULLY_DB_NAME`, `CULLY_DB_USER`, `CULLY_DB_PASSWORD` and `CULLY_DB_VOLUME` for the existing database. These values describe actual storage assets; changing their names does not migrate their contents. `CULLY_DB_VOLUME` must identify the existing external volume. Construct a correctly escaped database URL separately.

The root Compose workspace network is external and must already exist. The [fresh self-hosted Compose example](/hosting#minimal-single-user-compose-installation) creates its own network and volume. The [full-stack command](/hosting#one-command-full-stack) also starts a private Mem0 service and its persistent database. A separately operated Mem0 service must be reachable from the data service's network.

The release-triggered [personal stack](/personal-deployment) uses [`deploy/personal-data.env.example`](https://github.com/mcp-runtime/cully/blob/main/deploy/personal-data.env.example) and creates fresh, project-scoped PostgreSQL and Mem0 volumes. Its `CULLY_DATABASE_URL` points at the new `db` service; it does not use `CULLY_DB_VOLUME` or the old database network. Keep its database password, data API token, Mem0 API key, JWT secret and embedding-provider key on the VM. The `workspace_workspace` network must already exist for the private Caddy data route.

## Local CLI

The [local advisor guide](advisor.md#controls) lists local `CULLY_*` controls for display and advisor behavior. Agent-owned directories still honor `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `CURSOR_CONFIG_DIR`.

No `BUDDY_*` or `COCKPIT_*` environment aliases are maintained.
