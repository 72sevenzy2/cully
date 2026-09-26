# Connect an agent to Buddy

Buddy's Streamable HTTP endpoint is `https://mcp.mcpruntime.org/buddy/mcp`, served
by MCP Runtime. Clients use the MCP OAuth issuer
`https://auth.mcpruntime.org/mcp-auth`. The OAuth resource identifier is the full
endpoint URL above, and the issuer must allow it before clients can sign in.
Buddy's MCPServer uses `auth.mode: oauth`, so the Runtime operator routes both the
MCP endpoint and its protected-resource metadata path. The data API is on the
Buddy VM at `https://workspace.mcpruntime.org/buddy-data` and is used only by the
MCP workload.

Install the latest shared Buddy skill in the current user's Codex, Claude Code,
and Cursor directories by running `python3 scripts/install_skill.py` from the
Buddy checkout.

## Codex

Add the remote MCP server in `~/.codex/config.toml`:

```toml
[mcp_servers.buddy]
url = "https://mcp.mcpruntime.org/buddy/mcp"
```

Then run `codex mcp login buddy` and complete the browser OAuth flow.

## Claude Code

```sh
claude mcp add --transport http buddy https://mcp.mcpruntime.org/buddy/mcp
claude mcp get buddy
```

Open `/mcp` in Claude Code and complete OAuth when prompted. For Claude mobile,
add the remote connector through Claude web first, then enable it in the mobile
client if the account exposes that connector there.

## Cursor

Add this server to `~/.cursor/mcp.json` (or project `.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "buddy": {
      "url": "https://mcp.mcpruntime.org/buddy/mcp"
    }
  }
}
```

Restart Cursor and complete OAuth for Buddy when prompted.

## OAuth resource registration

The authorization service must issue tokens for exactly
`https://mcp.mcpruntime.org/buddy/mcp`. Buddy's resource is registered
with the current authorization issuer. If the public Buddy URL changes, update
the issuer allowlist and service audience together. Keep PostgreSQL private;
only expose the HTTPS MCP endpoint through Caddy.

## Protected resource host mismatch

If a client reports that Buddy's protected resource uses
`workspace.mcpruntime.org` while the configured endpoint uses
`mcp.mcpruntime.org`, inspect the public metadata response:

```sh
curl -fsS https://mcp.mcpruntime.org/.well-known/oauth-protected-resource/buddy/mcp | jq -r .resource
```

It should return `https://mcp.mcpruntime.org/buddy/mcp`. Keep the server's
`auth.audience` and the authorization server's resource registration set to that
full endpoint. If the running MCPServer still advertises the workspace host,
redeploy Buddy from `.mcp/servers.yaml` so the stored `spec.auth.audience` is
updated, then remove and add the Buddy MCP connection again to refresh cached
OAuth discovery. Keep `https://workspace.mcpruntime.org/buddy-data` as the
workload's data API URL; it is separate from the client-facing OAuth resource.
