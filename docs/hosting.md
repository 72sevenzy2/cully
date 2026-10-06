# Host Cully with Docker

Cully currently supports Docker Compose for self-hosting on a local machine or VM. The local advisor runs on the user's machine. The MCP service talks to a private data API. PostgreSQL stores authoritative source records; Mem0 uses its own pgvector database for semantic indexing and recall. See [architecture](/architecture) and [Mem0](/mem0).

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

Install Docker with the Compose plugin and Python 3. From the repository checkout, run:

```sh
cd deploy/self-hosted
./setup.sh codex
```

On first run, setup generates the PostgreSQL passwords, MCP-to-data token, data-to-Mem0 API key, Mem0 signing secret and stable no-OAuth owner in `~/.cully/config.json`. That file is mode `600` inside a mode `700` directory. Later runs reuse the values. If a previous `.env` already contains credentials, setup imports them before generating missing values; it does not rotate a database password behind an existing volume. `.env` is created from `.env.example` for nonsecret deployment options. Keep both files private if you add provider credentials to `.env`. The script exports saved credentials to Docker Compose; it never prints them.

[`compose.yaml`](https://github.com/mcp-runtime/cully/blob/main/deploy/self-hosted/compose.yaml) defines the full stack: Cully PostgreSQL, a separate Mem0 pgvector/PostgreSQL database, Mem0 REST, the data API, MCP, and optional Caddy and MCP Auth. `setup.sh` creates the Cully schema, then starts the services. MCP binds only to loopback by default. If the Cully CLI is already installed, `codex` also configures its local integration, shared skill and MCP connection. Use `claude` or `cursor`, or omit the agent to start only Docker services. Restart the agent after setup.

OAuth applies only to MCP. For OAuth, set the real `CULLY_AUTH_ISSUER`, `CULLY_AUTH_RESOURCE`, `CULLY_JWKS_URL` and HTTPS host values in `.env`. Then run `./setup.sh --oauth existing` for an existing authorization server. To run MCP Auth and Caddy in this same Compose stack, prepare the connector JSON, signing key and upstream client secret described in [OAuth deployment](/oauth), then run `./setup.sh --oauth mcp-auth`. Add `--mcp-url https://mcp.example.com/mcp codex` to configure an agent at the same time; the URL must equal `CULLY_AUTH_RESOURCE`. The database passwords and service tokens still come from `config.json` in either mode.

The full stack builds Mem0 from a [pinned upstream source commit](https://github.com/mem0ai/mem0/tree/c93420c49a6b14c3d446bdb156d96811908fd90a/server) with [FastEmbed's local BGE small model](https://qdrant.github.io/fastembed/examples/Supported_Models/). The model runs on the host CPU; indexing does not require an embedding API key. Cully calls Mem0 with `infer=false`, so it does not ask an LLM to extract facts. The Mem0 API and both databases stay on private Compose networks. The only published MCP port binds to loopback.

## Service credentials

PostgreSQL uses its native password authentication. The data API receives a PostgreSQL URL with that password; Mem0's separate PostgreSQL database uses another generated password. The MCP process authenticates to the data API with `CULLY_DATA_API_TOKEN`, while the data API authenticates to Mem0 with `CULLY_MEM0_API_KEY`. No database or Mem0 port is published. These service credentials do not depend on MCP OAuth. Back up `~/.cully/config.json` alongside the Docker volumes; changing the database passwords without updating the existing volumes will break connections.

## OAuth installation

The same MCP binary supports OAuth. The [OAuth guide](/oauth) has commands for an existing authorization server and for MCP Auth connected to your organization's identity provider. The examples add Caddy for HTTPS. Register the exact public MCP resource and grant `tools:read` and `tools:write`. Clients sign in only in OAuth mode.

## Example personal deployment on MCP Runtime

The [personal deployment guide](/personal-deployment) describes an example release-triggered stack with PostgreSQL and Mem0 volumes. Its `.mcp/servers.yaml` uses OAuth for public ingress; its issuer, JWKS URL and audience must match the authorization server. Keep a no-OAuth deployment on a private route.

Only the MCP URL belongs in agent configuration. The private API token, database credentials and Mem0 key stay with the operator. [Connect an agent](/agents) explains client setup.

Client formats and OAuth sign-in behavior also follow the official [Codex MCP documentation](https://developers.openai.com/codex/mcp), [Claude Code MCP documentation](https://code.claude.com/docs/en/mcp) and [Cursor MCP documentation](https://cursor.com/docs/mcp).
