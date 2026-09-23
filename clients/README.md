# Connect an agent to Buddy

Buddy's Streamable HTTP endpoint is `https://workspace.mcpruntime.org/buddy/mcp`.
Clients use the existing MCP OAuth issuer `https://auth.mcpruntime.org/mcp-auth`.
OAuth discovery depends on the resource identifier being allowed by that issuer.
The resource requested by Buddy is the full endpoint URL above.

## Codex

Add the remote MCP server in `~/.codex/config.toml`:

```toml
[mcp_servers.buddy]
url = "https://workspace.mcpruntime.org/buddy/mcp"
```

Then run `codex mcp login buddy` and complete the browser OAuth flow.

## Claude Code

```sh
claude mcp add --transport http buddy https://workspace.mcpruntime.org/buddy/mcp
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
      "url": "https://workspace.mcpruntime.org/buddy/mcp"
    }
  }
}
```

Restart Cursor and complete OAuth for Buddy when prompted.

## OAuth resource registration

The authorization service must issue tokens for exactly
`https://workspace.mcpruntime.org/buddy/mcp`. Buddy's resource is registered
with the current authorization issuer. If the public Buddy URL changes, update
the issuer allowlist and service audience together. Keep PostgreSQL private;
only expose the HTTPS MCP endpoint through Caddy.
