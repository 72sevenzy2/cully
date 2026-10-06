# Configuration

## Go memory services

| Variable | Used by | Meaning |
| --- | --- | --- |
| `CULLY_DATABASE_URL` | Data service | PostgreSQL connection URL; required |
| `CULLY_DB_MAX_CONNS` | Data service | Pool maximum, 2–100; default 8 |
| `CULLY_DATA_API_TOKEN` | Both services | Shared private API bearer token; required for serving |
| `CULLY_DATA_API_URL` | MCP service | Private API base URL; required |
| `CULLY_MEM0_URL` | Data service | Self-hosted REST base URL; optional with matching key |
| `CULLY_MEM0_API_KEY` | Data service | Mem0 service API key; required when URL is set |
| `CULLY_HOST` | Both services | Bind address; default `0.0.0.0` |
| `CULLY_PORT` | Both services | Default 8080 for MCP, 8083 for data |
| `CULLY_AUTH_ISSUER` | MCP service | OAuth issuer; default `https://auth.mcpruntime.org/mcp-auth` |
| `CULLY_AUTH_RESOURCE` | MCP service | Public resource; default `https://mcp.mcpruntime.org/cully/mcp` |
| `CULLY_JWKS_URL` | MCP service | Signing-key URL; defaults to the issuer's `/.well-known/jwks.json` |

MCP Runtime can inject platform variables `MCP_AUTH_ISSUER`, `MCP_AUTH_RESOURCE` and `MCP_PATH`. Platform issuer/resource values take precedence over Cully defaults. `MCP_PATH` selects the workload's internal transport path, default `/mcp`; the public OAuth resource still includes the external `/cully/mcp` route.

Keep the data API token identical on both sides. The MCP workload does not receive database credentials or a Mem0 key. The data service accepts forwarded owner identity only from authenticated service calls.

## VM Compose

`.env.example` lists `CULLY_DB_NAME`, `CULLY_DB_USER`, `CULLY_DB_PASSWORD` and `CULLY_DB_VOLUME` for the existing database. These values describe actual storage assets; changing their names does not migrate their contents. `CULLY_DB_VOLUME` must identify the existing external volume. Construct a correctly escaped database URL separately.

The Compose workspace network is external and must already exist. Self-hosted Mem0 must be reachable from the data service's network. Configuration alone does not start or reconfigure an existing Mem0 instance.

## Local CLI

The [session-control guide](flightdeck/README.md#controls) lists local `CULLY_*` controls for display and advisor behavior. Agent-owned directories still honor `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `CURSOR_CONFIG_DIR`.

No `BUDDY_*` or `COCKPIT_*` environment aliases are maintained.
