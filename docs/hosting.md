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

## Fresh single-user Compose installation

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

Connect a local agent with `cully mcp add --agent codex --url http://127.0.0.1:8080/mcp`. A remote agent needs a private tunnel or trusted network route. The example does not ship a Mem0 server: set `CULLY_MEM0_URL` and `CULLY_MEM0_API_KEY` only after provisioning the [self-hosted Mem0 REST service](/mem0). `cully_recall` requires Mem0; logging, text search and recent records work without it.

## OAuth installation

The same MCP binary supports OAuth. The [OAuth guide](/oauth) has commands for an existing authorization server and for a new MCP Auth broker with a Keycloak connector. The examples add Caddy for HTTPS. Register the exact public MCP resource and grant `tools:read` and `tools:write`. Clients sign in only in OAuth mode.

## Existing personal deployment

The root `compose.yaml` and [personal deployment guide](/personal-deployment) target an existing VM database volume and network. They are separate from the fresh installation above. The personal `.mcp/servers.yaml` stays explicitly in OAuth mode for its public ingress; its JWKS URL and audience must match the authorization server. Do not reuse its public route in no-OAuth mode.

Only the MCP URL belongs in agent configuration. The private API token, database credentials and Mem0 key stay with the operator. [Connect an agent](/agents) explains client setup.

Client formats and OAuth sign-in behavior also follow the official [Codex MCP documentation](https://developers.openai.com/codex/mcp), [Claude Code MCP documentation](https://code.claude.com/docs/en/mcp) and [Cursor MCP documentation](https://cursor.com/docs/mcp).
