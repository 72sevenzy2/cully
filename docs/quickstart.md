---
title: Quickstart
description: Install Cully, start its memory stack and try it in your coding agent.
---

# Quickstart

Cully saves useful notes for your connected agents and helps you review suggestions while you work. This setup runs on your laptop. It needs Docker with the Compose plugin, Python 3 and Claude Code, Codex or Cursor.

## 1. Install Cully

For Codex:

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex
```

Use `--agent claude` or `--agent cursor` for another agent. The installer sets up the agent and starts the advisor daemon in the background automatically. If your shell cannot find `cully`, use the binary path printed by the installer or add that directory to `PATH`.

## 2. Start Cully memory

Make sure Docker is running, then run:

```sh
cully setup codex
```

Use the same agent name you installed. Setup starts PostgreSQL, Mem0, the private data API and Cully MCP in Docker, generates private service credentials, and connects your agent. Wait for `Cully MCP is configured at ...`, then restart the agent. In Codex, review and trust the installed Cully hooks in `/hooks` when Codex asks. No repository checkout or OAuth setup is needed. See [self-hosting](/hosting#one-command-full-stack) for details.

## 3. Continue work in another session

Work on a real task in your connected agent. Cully's installed session instructions ask the agent to look up relevant project notes before substantial work and save one concise summary when it finishes. Start a new session or switch between Claude Code, Codex and Cursor with the same Cully MCP server; the agent can pick up the task, decisions, result and remaining steps from those notes. You can also ask what happened earlier.

The agent starts with a bounded `cully_context` lookup and uses `cully_get` only when it needs a full note. It saves the result with `cully_log` through its own MCP connection. Mem0 handles semantic recall; PostgreSQL keeps the source note. The advisor does not upload transcripts. If the MCP server is unavailable, the agent can keep working but cannot save shared continuity until it reconnects. See the [memory guide](/memory) and [session optimization guide](/session-optimization).

## 4. Check suggestions

In a terminal, run:

```sh
cully status
cully suggestions
```

You can also use Cully inside your agent:

| Agent | In-session control |
| --- | --- |
| Claude Code | Look at the Cully status line or run `/cully suggestions`. |
| Codex | Run `/prompts:cully suggestions` or ask Codex to check suggestions. |
| Cursor | Run the project `/cully suggestions` command or ask Cursor to check suggestions. |

If Cully lists a numbered improvement, preview it with `cully apply 1 --dry-run`, using the number shown on your machine. Review the preview before applying. `cully status` also shows whether the advisor daemon is running; if startup failed, run `cully agent setup codex` to retry. The [advisor guide](/advisor) explains each agent's integration.
