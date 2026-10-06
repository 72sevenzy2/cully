---
title: Connect an agent
description: Set up Cully's local advisor and optional shared memory in Claude Code, Codex, or Cursor.
---

# Connect an agent

Cully has two setup commands. `cully setup` adds the local advisor integration. `cully mcp add` adds one remote shared-memory connection. You can use the local advisor on its own.

You can do both setup steps together with `cully setup codex --mcp-url URL`, or pass `--agent codex --mcp-url URL` to the [release installer](/installation#release-installer). Add `--oauth` only if that endpoint uses OAuth. The Cully skill then guides session checks and calls the configured `cully_*` memory tools when useful.

## Claude Code

```sh
cully setup claude
cully mcp add --agent claude --url https://your-server.example/mcp
```

Restart Claude Code. The local integration includes a status line, session hooks, a `/cully` command, and a skill. If your MCP server enables OAuth, add `--oauth` to `cully mcp add`, then open `/mcp` to sign in. The URL is for your own deployed service until Cully's hosted endpoint is available.

## Codex

```sh
cully setup codex
cully mcp add --agent codex --url https://your-server.example/mcp
```

Restart Codex. The local integration uses native status fields, a `/prompts:cully` prompt, an AGENTS.md pointer, and a skill. For an OAuth-protected MCP server, add `--oauth` to `cully mcp add` and then run `codex mcp login cully`. The default no-OAuth server connects directly.

## Cursor

```sh
cully setup cursor
cully mcp add --agent cursor --url https://your-server.example/mcp
```

Restart Cursor. If the MCP server enables OAuth, add `--oauth` to `cully mcp add` and sign in from its MCP settings. The local integration provides a project `/cully` command and skill.

## What setup changes

`cully setup` configures the selected local agent and starts the advisor daemon. `cully mcp add` registers one user-level MCP connection. It does not deploy a server, sign you in automatically, or replace a Cully entry that points at another URL. Use `cully status` to inspect local integration state.

For operator requirements, see [self-hosting](/hosting) and [OAuth deployment](/oauth). For all client configuration details, see the [client reference](https://github.com/mcp-runtime/cully/blob/main/clients/README.md).
