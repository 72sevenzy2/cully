# Host Cully

Cully provides a one-command Docker Compose stack for self-hosting on a local machine or VM. Its services are containers configured through environment variables, so an ops team can run them with Docker, Kubernetes or its existing container platform. The local advisor runs on each user's machine. The MCP service talks to a private data API. PostgreSQL stores authoritative source records; Mem0 uses its own pgvector database for semantic indexing and recall. See [architecture](/architecture) and [Mem0](/mem0).

The Compose stack runs all services together. For a team deployment on another platform, use the [service and authentication contract](/team-deployment). [MCP Runtime](https://mcpruntime.org) is a separate platform for deploying MCP servers; see its [publishing guide](https://docs.mcpruntime.org/publish-mcp-server/).

## Choose MCP access

| Mode | Start | Owner | Network |
| --- | --- | --- | --- |
| No OAuth (default) | `cully-mcp` or `CULLY_MCP_AUTH_MODE=none` | Required `CULLY_MCP_OWNER`; every call uses this fixed owner | Loopback or private network. Anyone who reaches the endpoint can use its tools. |
| OAuth | `cully-mcp --oauth` or `CULLY_MCP_AUTH_MODE=oauth` | Verified token subject | HTTPS route, with issuer, audience, signature, expiration and read/write scopes checked. |

No-OAuth mode makes no identity-provider or JWKS requests. Cully never accepts an owner from a tool input. Keep a stable `CULLY_MCP_OWNER` to access the same records over time. Do not expose no-OAuth mode on an untrusted public route. OAuth is provider-neutral: MCP Auth with a connector is one option, while another compatible authorization server can work too. See [OAuth deployment](/oauth).

## One-command full stack

Install Docker with the Compose plugin, Python 3 and the [Cully CLI](/installation). Then run:

```sh
cully setup codex
```

The CLI downloads the matching release's Docker files into `~/.cully/self-hosted/releases/`; no checkout is needed. On first run, setup generates the PostgreSQL passwords, MCP-to-data token, data-to-Mem0 API key, Mem0 signing secret and stable no-OAuth owner in `~/.cully/config.json`. That file is mode `600` inside a mode `700` directory. Later runs reuse the values; they do not rotate a database password behind an existing volume. Nonsecret deployment options go in `~/.cully/self-hosted/config/.env`. Keep that file private if you add provider credentials to it. Setup exports saved credentials to Docker Compose without printing them.

[`compose.yaml`](https://github.com/mcp-runtime/cully/blob/main/deploy/self-hosted/compose.yaml) defines the full stack: Cully PostgreSQL, a separate Mem0 pgvector/PostgreSQL database, Mem0 REST, the data API, MCP, and optional Caddy and MCP Auth. Setup creates the Cully schema, then starts the services. MCP binds only to loopback by default. `codex` also configures the local integration, skill and MCP connection. Use `claude` or `cursor`, or omit the agent to start only Docker services. Restart the agent after setup.

OAuth applies only to MCP. Run `cully setup --prepare` to create editable config files without starting services. Set `CULLY_MCP_HOST`, `CULLY_AUTH_HOST` and the upstream client secret in `.env`, prepare the MCP Auth connector and signing key, then run `cully setup codex --oauth`. Setup derives the resource URL, issuer and JWKS URL from the hostnames. See [MCP OAuth](/oauth) for the connector setup.

The full stack builds Mem0 from a [pinned upstream source commit](https://github.com/mem0ai/mem0/tree/c93420c49a6b14c3d446bdb156d96811908fd90a/server) with [FastEmbed's local BGE small model](https://qdrant.github.io/fastembed/examples/Supported_Models/). The model runs on the host CPU; indexing does not require an embedding API key. Cully calls Mem0 with `infer=false`, so it does not ask an LLM to extract facts. The Mem0 API and both databases stay on private Compose networks. The only published MCP port binds to loopback.

## Service credentials

PostgreSQL uses its native password authentication. The data API receives a PostgreSQL URL with that password; Mem0's separate PostgreSQL database uses another generated password. The MCP process authenticates to the data API with `CULLY_DATA_API_TOKEN`, while the data API authenticates to Mem0 with `CULLY_MEM0_API_KEY`. No database or Mem0 port is published. These service credentials do not depend on MCP OAuth. Back up `~/.cully/config.json` alongside the Docker volumes; changing the database passwords without updating the existing volumes will break connections.

## OAuth installation

The same MCP binary supports OAuth. The [OAuth guide](/oauth) shows MCP Auth connected to your organization's identity provider. The Compose stack adds Caddy for HTTPS. Register the exact public MCP resource and grant `tools:read` and `tools:write`. Clients sign in only in OAuth mode.

## Team deployment

The [team deployment guide](/team-deployment) shows how MCP Auth can connect Cully's public MCP endpoint to an organization's identity provider. The MCP resource, issuer and JWKS URL must match the authorization server. Keep a no-OAuth deployment on a private route.

Only the MCP URL belongs in agent configuration. The private API token, database credentials and Mem0 key stay with the operator. [Connect an agent](/agents) explains client setup.

Client formats and OAuth sign-in behavior also follow the official [Codex MCP documentation](https://developers.openai.com/codex/mcp), [Claude Code MCP documentation](https://code.claude.com/docs/en/mcp) and [Cursor MCP documentation](https://cursor.com/docs/mcp).
