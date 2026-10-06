---
title: Quickstart
description: Install Cully and inspect your first local session.
---

# Quickstart

Get the local advisor running first. It does not need a Cully account or remote service.

## 1. Install Cully for your agent

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex
```

Use `--agent claude` or `--agent cursor` for another agent. The installer adds the local advisor and Cully skill. Restart your agent after setup. See [installation](/installation) for options.

## 2. See guidance inside your agent

Restart your coding agent, then check Cully while you work:

- **Claude Code:** watch the live Cully status line and run `/cully suggestions`.
- **Codex:** run `/prompts:cully suggestions`, or ask the agent to run `cully suggestions`.
- **Cursor:** run the project `/cully suggestions` command, or ask the agent to check Cully suggestions.

The installed skill helps the agent use Cully's local guidance. Claude supplies
live hook signals; Codex and Cursor use their supported commands, skills, and
native session displays. See [in-session guidance](/advisor#see-suggestions-as-you-work).

You can also check from your terminal:

```sh
cully status
cully suggestions
```

Cully starts the advisor and shows session status for the selected agent.

If your shell cannot find `cully`, use the binary path printed by the installer or add that directory to `PATH`.

## 3. Preview a suggestion

When `cully suggestions` lists a numbered improvement, inspect it before applying:

```sh
cully apply 1 --dry-run
```

Local guidance is optional and advisory. Cully preserves unrelated, user-owned agent configuration. The [local advisor guide](/advisor) explains the commands and controls.

## Add shared memory when your server is ready

For a self-hosted Docker deployment, run `cully setup codex`. It generates private credentials in `~/.cully/config.json`, starts MCP, the data API, PostgreSQL and Mem0, then registers the MCP URL and Cully skill. No checkout is needed. See [self-hosting](/hosting#one-command-full-stack).

If your server is already running, connect it to the agent you set up:

```sh
cully agent setup codex --mcp-url http://127.0.0.1:8080/mcp
```

The default single-user server needs no login. If the operator enables OAuth, use a public HTTPS URL and add `--oauth` to the setup command; then run `codex mcp login cully`. See [agent setup](/agents) for Claude Code and Cursor. The local advisor also works without an MCP server.
