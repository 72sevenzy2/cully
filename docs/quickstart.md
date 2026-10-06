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

Use the same agent name you installed. Setup starts PostgreSQL, Mem0, the private data API and Cully MCP in Docker, generates private service credentials, and connects your agent. Wait for `Cully MCP is configured at ...`, then restart the agent. No repository checkout or OAuth setup is needed. See [self-hosting](/hosting#one-command-full-stack) for details.

## 3. Save and find a note

Ask your agent to save a short summary of useful work with Cully, then ask it to find that summary. The agent uses `cully_log` to save it and `cully_search` or `cully_recall` to find it. Mem0 handles semantic recall; PostgreSQL keeps the source note. Cully does not upload full session transcripts automatically. See the [memory guide](/memory).

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
