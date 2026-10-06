# Hosted and self-hosted Cully

The deployment model changes who operates the server, not the client protocol.

| Component | Cully hosted service | Self-hosted service |
| --- | --- | --- |
| Local CLI/advisor | User's machine | User's machine |
| `cully-mcp` | Cully operator | Self-hosting operator |
| `cully-data` and PostgreSQL | Cully operator | Self-hosting operator |
| Self-hosted Mem0 runtime | Cully operator | Self-hosting operator |

## Coding-agent setup

```sh
cully install
cully mcp add --agent codex
# Or connect this agent to your own deployment:
cully mcp add --agent codex --url https://my-host.example/mcp
```

Repeat MCP setup separately for each chosen agent. The default URL is
`https://mcp.mcpruntime.org/cully/mcp`; this endpoint requires the planned Cully
cutover before it is available. Setup writes the connection, then prints client
OAuth instructions. It does not deploy the service or perform sign-in.

Only the MCP URL belongs in agent configuration. The data API service token,
database credentials and Mem0 key stay with the operator. The local daemon
supports session analysis; it is not a local copy of the remote memory servers.

## Operator setup

Configure `cully-data` with `CULLY_DATABASE_URL`, `CULLY_DATA_API_TOKEN`, and the
private self-hosted Mem0 URL/key. PostgreSQL is authoritative; Mem0 adds semantic
recall when enabled. Deploy and persist Mem0 separately using [its guide](mem0.md).

Configure `cully-mcp` with the matching `CULLY_DATA_API_TOKEN`,
`CULLY_DATA_API_URL`, OAuth issuer/JWKS and its exact public resource URL. An
operator may place data and MCP on the same host or use a private authenticated
network between them. Agents never connect directly to the data API.

For the existing Buddy VM, follow [migration.md](migration.md) and retain its
actual database volume. A new self-hosted installation must provision a database,
volume, private network, compatible OAuth issuer and HTTPS routing. The supplied
Compose file is an existing-volume data-service deployment; it does not provision
an OAuth issuer or the public MCP workload. Build the public service using
`Dockerfile` and deploy it through your container platform; MCP Runtime metadata
is in `.mcp/servers.yaml`. See [configuration](configuration.md) for every variable.

Before connecting clients, run the explicit schema migration and verify health,
OAuth metadata, read/write scopes, owner isolation and Mem0 indexing. Registration
in a coding agent is not evidence that the server has been deployed.

Client formats and sign-in steps follow the official
[Codex MCP documentation](https://developers.openai.com/codex/mcp),
[Claude Code MCP documentation](https://code.claude.com/docs/en/mcp) and
[Cursor MCP documentation](https://cursor.com/docs/mcp).
