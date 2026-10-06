# Host Cully

Cully's local advisor runs on your machine. The optional MCP service talks to a private data API. PostgreSQL stores authoritative source records; self-hosted Mem0 does the semantic indexing and recall work when configured. See [architecture](/architecture) and [Mem0](/mem0).

| Component | Planned Cully hosted service | Self-hosted service |
| --- | --- | --- |
| Local CLI and advisor | User's machine | User's machine |
| `cully-mcp` | Cully operator | Self-hosting operator |
| `cully-data` and PostgreSQL | Cully operator | Self-hosting operator |
| Self-hosted Mem0 runtime | Cully operator | Self-hosting operator |

## Choose MCP access

| Mode | Start | Owner | Network |
| --- | --- | --- | --- |
| No OAuth (default) | `cully-mcp` or `CULLY_MCP_AUTH_MODE=none` | Required `CULLY_MCP_OWNER`; every call uses this fixed owner | Loopback, private network or trusted tunnel. Anyone who reaches the endpoint can use its tools. |
| OAuth | `cully-mcp --oauth` or `CULLY_MCP_AUTH_MODE=oauth` | Verified token subject | HTTPS route, with issuer, audience, signature, expiration and read/write scopes checked. |

No-OAuth mode makes no identity-provider or JWKS requests. Cully never accepts an owner from a tool input. Keep a stable `CULLY_MCP_OWNER` to access the same records over time. Do not expose no-OAuth mode on an untrusted public route. OAuth is provider-neutral: MCP Auth with a connector is one option, while another compatible authorization server can work too. See [OAuth deployment](/oauth).

## One-command full stack

For a fresh self-hosted instance, copy [`deploy/self-hosted/.env.example`](https://github.com/mcp-runtime/cully/blob/main/deploy/self-hosted/.env.example) to `.env` and fill in the database password and URL, private API token, stable `CULLY_MCP_OWNER`, Mem0 database password, JWT secret, API key, and embedding-provider key. Keep `.env` private. From `deploy/self-hosted`, run:

```sh
./start.sh codex
```

The script starts Cully's PostgreSQL database, a separate persistent Mem0 database, the self-hosted Mem0 REST server, Cully's data API, and Cully MCP. It runs the Cully schema migration first. If the Cully CLI is already installed, the `codex` argument also installs its local integration and shared Cully skill and registers the loopback MCP URL. Use `claude` or `cursor` instead, or omit the argument to start services only. Restart the agent after setup. This is a fresh-install stack; do not point it at an existing personal database volume.

For OAuth, also fill in `CULLY_AUTH_ISSUER`, `CULLY_AUTH_RESOURCE`, `CULLY_JWKS_URL`, `CULLY_MCP_HOST`, and the HTTPS routing values in `.env`. The server cannot infer an issuer or signing keys from the client flag. With an existing authorization server, run `./start.sh --oauth existing`; with a new MCP Auth broker and Keycloak connector, prepare its connector JSON, signing key, and Keycloak client secret as described in [OAuth deployment](/oauth), then run `./start.sh --oauth mcp-auth`. To register an agent in that same command, append `--mcp-url https://mcp.example.com/mcp codex`; the URL must equal `CULLY_AUTH_RESOURCE` in `.env`.

The full stack builds Mem0 from a [pinned upstream source commit](https://github.com/mem0ai/mem0/tree/c93420c49a6b14c3d446bdb156d96811908fd90a/server) because its published image is stale. It uses an OpenAI embedding key by default; authored memory text is sent to that embedding provider. Mem0 fact extraction is disabled. You can configure another Mem0 embedder separately. The Mem0 API and both databases stay on private Compose networks. The only published MCP port binds to loopback.

## Minimal single-user Compose installation

The files in [`deploy/self-hosted`](https://github.com/mcp-runtime/cully/tree/main/deploy/self-hosted) build the MCP and data binaries, create a PostgreSQL volume and bind MCP to `127.0.0.1:8080`. From the repository root:

```sh
cd deploy/self-hosted
cp .env.example .env
# Edit .env: database password, URL-encoded database password in CULLY_DATABASE_URL,
# private API token, and stable CULLY_MCP_OWNER.
chmod 600 .env
docker compose up -d db
docker compose --profile ops run --rm migrate
docker compose up -d --build data-api mcp
curl -fsS http://127.0.0.1:8080/healthz
```

Connect a local agent with `cully install codex --mcp-url http://127.0.0.1:8080/mcp`, which installs the skill and registers MCP together. A remote agent needs a private tunnel or trusted network route. This minimal example does not start Mem0: set `CULLY_MEM0_URL` and `CULLY_MEM0_API_KEY` only after provisioning the [self-hosted Mem0 REST service](/mem0), or use `./start.sh` for the full stack. `cully_recall` requires Mem0; logging, text search and recent records work without it.

## OAuth installation

The same MCP binary supports OAuth. The [OAuth guide](/oauth) has commands for an existing authorization server and for a new MCP Auth broker with a Keycloak connector. The examples add Caddy for HTTPS. Register the exact public MCP resource and grant `tools:read` and `tools:write`. Clients sign in only in OAuth mode.

## Existing personal deployment

The [personal deployment guide](/personal-deployment) describes a release-triggered stack on the existing VM with fresh, isolated PostgreSQL and Mem0 volumes. It leaves any prior database volume alone until records are deliberately migrated. The personal `.mcp/servers.yaml` stays explicitly in OAuth mode for its public ingress; its issuer, JWKS URL and audience must match the authorization server. Do not reuse its public route in no-OAuth mode.

Only the MCP URL belongs in agent configuration. The private API token, database credentials and Mem0 key stay with the operator. [Connect an agent](/agents) explains client setup.

Client formats and OAuth sign-in behavior also follow the official [Codex MCP documentation](https://developers.openai.com/codex/mcp), [Claude Code MCP documentation](https://code.claude.com/docs/en/mcp) and [Cursor MCP documentation](https://cursor.com/docs/mcp).
