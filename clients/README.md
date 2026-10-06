> These are the planned Cully endpoints. Follow docs/migration.md before reconnecting existing clients; the repository rename does not activate them.

# Connect an agent to Cully

Cully's Streamable HTTP endpoint is `https://mcp.mcpruntime.org/cully/mcp`, served
by MCP Runtime. Clients use the MCP OAuth issuer
`https://auth.mcpruntime.org/mcp-auth`. The OAuth resource identifier is the full
endpoint URL above, and the issuer must allow it before clients can sign in.
Cully's MCPServer uses `auth.mode: oauth`, so the Runtime operator routes both the
MCP endpoint and its protected-resource metadata path. The data API is on the
Cully VM at `https://workspace.mcpruntime.org/cully-data` and is used only by the
MCP workload.

Install the latest shared Cully skill in the current user's Codex, Claude Code,
and Cursor directories by running `cully install` from the
Cully checkout.

## Codex

Add the remote MCP server in `~/.codex/config.toml`:

```toml
[mcp_servers.cully]
url = "https://mcp.mcpruntime.org/cully/mcp"
```

Then run `codex mcp login cully` and complete the browser OAuth flow.

## Claude Code

```sh
claude mcp add --transport http cully https://mcp.mcpruntime.org/cully/mcp
claude mcp get cully
```

Open `/mcp` in Claude Code and complete OAuth when prompted. For Claude mobile,
add the remote connector through Claude web first, then enable it in the mobile
client if the account exposes that connector there.

## Cursor

Add this server to `~/.cursor/mcp.json` (or project `.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "cully": {
      "url": "https://mcp.mcpruntime.org/cully/mcp"
    }
  }
}
```

Restart Cursor and complete OAuth for Cully when prompted.

## OAuth resource registration

The authorization service must issue tokens for exactly
`https://mcp.mcpruntime.org/cully/mcp`. Register Cully's resource with the authorization issuer before cutover, and grant
`tools:read` plus `tools:write` when mutation tools are needed. If the public Cully URL changes, update
the issuer allowlist and service audience together. Keep PostgreSQL private;
only expose the HTTPS MCP endpoint through Caddy.

## Protected resource host mismatch

If a client reports that Cully's protected resource uses
`workspace.mcpruntime.org` while the configured endpoint uses
`mcp.mcpruntime.org`, inspect the public metadata response:

```sh
curl -fsS https://mcp.mcpruntime.org/.well-known/oauth-protected-resource/cully/mcp | jq -r .resource
```

It should return `https://mcp.mcpruntime.org/cully/mcp`. Keep the server's
`auth.audience` and the authorization server's resource registration set to that
full endpoint. If the running MCPServer still advertises the workspace host,
redeploy Cully from `.mcp/servers.yaml` so the stored `spec.auth.audience` is
updated, then remove and add the Cully MCP connection again to refresh cached
OAuth discovery. Keep `https://workspace.mcpruntime.org/cully-data` as the
workload's data API URL; it is separate from the client-facing OAuth resource.
