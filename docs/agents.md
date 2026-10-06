---
title: Connect an agent
description: Set up Cully's local advisor and optional shared memory in Claude Code, Codex, or Cursor.
---

# Connect an agent

When your MCP server is running, one command installs the Cully skill and local advisor integration **and** registers the server with your agent:

```sh
cully setup codex --mcp-url https://mcp.example.com/mcp
```

Replace `codex` with `claude` or `cursor`. Add `--oauth` if that MCP endpoint requires sign-in. The skill guides local session checks and calls the configured `cully_*` memory tools when useful. `cully setup` does not deploy a server; [self-hosting](/hosting) explains how to start one.

## Claude Code

```sh
cully setup claude --mcp-url https://your-server.example/mcp
```

Restart Claude Code. The local integration includes a status line, session hooks, a `/cully` command, and a skill. If the server uses OAuth, append `--oauth` to the setup command and use `/mcp` to sign in.

## Codex

```sh
cully setup codex --mcp-url https://your-server.example/mcp
```

Restart Codex. The local integration uses native status fields, a `/prompts:cully` prompt, an AGENTS.md pointer, and a skill. For an OAuth-protected server, append `--oauth` to the setup command, then run `codex mcp login cully`.

## Cursor

```sh
cully setup cursor --mcp-url https://your-server.example/mcp
```

Restart Cursor. If the server uses OAuth, append `--oauth` to the setup command and sign in from Cursor's MCP settings. The local integration provides a project `/cully` command and skill.

## What setup changes

`cully setup AGENT` without `--mcp-url` installs only the local advisor and skill. If that integration is already installed and you only need to add an MCP connection, use `cully mcp add --agent AGENT --url URL` (plus `--oauth` when needed). Both commands preserve unrelated agent settings; neither signs you in automatically or replaces a Cully entry that points at another URL. Use `cully status` to inspect local integration state.

For operator requirements, see [self-hosting](/hosting) and [OAuth deployment](/oauth). For all client configuration details, see the [client reference](https://github.com/mcp-runtime/cully/blob/main/clients/README.md).
