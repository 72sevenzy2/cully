---
title: Connect an agent
description: Connect Claude Code, Codex or Cursor to Cully and its memory tools.
---

# Connect an agent

The [quickstart](/quickstart) starts Cully and connects your first agent. To connect another agent or use an existing Cully server, give that agent the MCP URL printed by setup.

```sh
cully agent setup codex --mcp-url http://127.0.0.1:8080/mcp
```

Replace `codex` with `claude` or `cursor` and use the URL printed by your server. Restart your agent after setup. For an OAuth-enabled HTTPS server, append `--oauth` and sign in using the steps below.

| Agent | Local Cully controls | OAuth sign-in, when enabled |
| --- | --- | --- |
| Claude Code | Status line and `/cully suggestions` | Use `/mcp`. |
| Codex | `/prompts:cully suggestions` or ask Codex to check suggestions | Run `codex mcp login cully`. |
| Cursor | Project `/cully suggestions` or ask Cursor to check suggestions | Use Cursor's MCP settings. |

The [local advisor guide](/advisor) shows what each integration can display. The [memory guide](/memory) explains what the MCP tools save and retrieve.

## If you already installed the advisor

Use `cully mcp add --agent codex --url URL` to add only the MCP connection. Add `--oauth` when the server requires sign-in. Setup preserves unrelated agent configuration and does not replace a Cully connection that points at another URL. Run `cully status` to inspect the local integration.

For detailed client configuration, see the [client reference](https://github.com/mcp-runtime/cully/blob/main/clients/README.md).
