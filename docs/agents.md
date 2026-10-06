---
title: Connect an agent
description: Set up Cully's local advisor and optional shared memory in Claude Code, Codex, or Cursor.
---

# Connect an agent

Cully has two setup commands. `cully install` adds the local advisor integration. `cully mcp add` adds one remote shared-memory connection. You can use the local advisor on its own.

## Claude Code

```sh
cully install claude
cully mcp add --agent claude --url https://your-server.example/mcp
```

Restart Claude Code. The local integration includes a status line, session hooks, a `/cully` command, and a skill. If you added MCP, open `/mcp` and complete OAuth sign-in. The URL is for your own deployed service until Cully's hosted endpoint is available.

## Codex

```sh
cully install codex
cully mcp add --agent codex --url https://your-server.example/mcp
codex mcp login cully
```

Restart Codex. The local integration uses native status fields, a `/prompts:cully` prompt, an AGENTS.md pointer, and a skill. Codex sign-in is needed only for the remote MCP connection.

## Cursor

```sh
cully install cursor
cully mcp add --agent cursor --url https://your-server.example/mcp
```

Restart Cursor, then complete sign-in from its MCP settings. The local integration provides a project `/cully` command and skill.

## What setup changes

`cully install` configures the selected local agent and starts the advisor daemon. `cully mcp add` registers one user-level MCP connection. It does not deploy a server, sign you in automatically, or replace a Cully entry that points at another URL. Use `cully status` to inspect local integration state.

For operator requirements, see [self-hosting](/hosting). For all client configuration details, see the [client reference](https://github.com/mcp-runtime/cully/blob/main/clients/README.md).
