# Cully

Live instruments and control suggestions for long-running coding-agent sessions.

Cully installs the `cully` command. In Claude Code it adds a compact
status line and `Stop` hook that suggest the next useful control before a session
gets expensive, repetitive, or hard to steer. Across Claude Code, Codex, and
Cursor it discovers agent-specific project surfaces and can propagate accepted
rules or skills into the right files.

## Why use it

- See branch, PR state, model, effort, context pressure, token churn, rate-limit
  usage, and session cost while you work.
- Get timely suggestions for `/compact`, `/clear`, model changes, skills,
  subagents, MCP, graphify, and workflow tools.
- Accept a suggestion in one step: numbered rows in the status bar map to
  `cully apply <n>`, which updates project config only after you confirm.
- Keep Claude Code, Codex, and Cursor project guidance aligned from the same
  accepted control.
- Save compact hourly memory of coding-agent sessions for retrieval by other
  tools without storing raw transcripts.

## Agent support

| Agent | Support |
|---|---|
| Claude Code | Rich live status line, `Stop`/`SessionEnd` hooks, `/cully`, MCP discovery, skills, subagents |
| Codex | Native `[tui].status_line` fields, `/prompts:cully` saved prompt, `AGENTS.md`, shared skill, agent discovery, shared `.mcp.json` discovery |
| Cursor | `/cully` project command, shared Cully skill, Cursor MCP config discovery |

Claude Code is the only client with a command-backed rich status-line and hook
payload API, so it receives the full Flightdeck renderer and live advisor hooks.
Codex receives its native built-in status fields plus a saved cully prompt;
Cursor receives a native project `/cully` command. Both clients can use the
shared skill and session-memory integrations, but neither currently exposes a
public API for embedding Flightdeck's custom ANSI renderer in its own status bar.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/mcp-runtime/cully/main/install.sh | bash
```

The installer auto-detects installed/configured coding agents and registers
Cully for the ones it finds.

The installer downloads the matching macOS or Linux release binary, installs it
to `~/.claude/bin/cully`, then runs `cully install`. For Claude Code it
merges the `statusLine` plus hooks into `~/.claude/settings.json`; for Codex it
writes an `AGENTS.md` pointer, a saved prompt at `CODEX_HOME/prompts/cully.md`,
and native status fields in `CODEX_HOME/config.toml`; for Cursor it writes the
shared skill and `.cursor/commands/cully.md`. Existing Claude settings, Codex
prompts/status fields, and Cursor commands are preserved when they are
user-owned. Restart each client after installation so it reloads commands and
configuration.

Build from source:

```bash
go install github.com/mcp-runtime/cully/cmd/cully@latest
cully install
```

`cully install` auto-detects present agents. To force a specific target:

```bash
cully install codex
cully install cursor
cully install all
```

## What it shows

- **Claude status line:** project, git state, model, effort, context fill, token
  churn, cache/output tokens, rate limits, and cost.
- **Codex status line:** the native model/reasoning, directory, branch, context,
  rate-limit, and fast-mode fields configured by `cully install codex`.
- **Codex/Cursor cully command:** run the same `cully systems`, `status`,
  `list`, `checklist`, `plan`, `debrief`, and daemon controls inside the agent.
- **Session advisor:** a background `haiku` check that surfaces the highest-value
  next controls for the current session.
- **Session memory:** the daemon scans changed Claude/Codex/Cursor session files
  hourly and writes compact JSONL summaries of what the user asked, when it
  happened, which tools ran, and which files were touched.
- **Tool awareness:** suggestions can reference Claude Code commands, shared
  Cully skills, installed Claude/Codex skills and agents, MCP
  resources, graphify state, and audited third-party tool gaps.
- **Non-blocking runtime:** analysis runs detached, so your turn does not wait on
  the advisor.

## Accepting suggestions

After a few turns, the advisor writes suggestions to the status bar.
**Notes** are informational only. **Numbered fixes** can be wired into your
project with `cully apply`:

```bash
cully list
cully apply 1
cully apply 2 --yes
cully apply 3 --dry-run
```

When you accept a fix, Cully may:

- Append the accepted rule to `.cully/skills/cully/SKILL.md`.
- Merge MCP servers into `.mcp.json`.
- Write project skills to `.cully/skills/<name>/SKILL.md`.
- Run safelisted install commands such as `brew`, `npm`, or `npx -y`.

Restart Claude Code or run `/hooks` after MCP servers are added so they load.

## Background memory

The daemon runs a memory scan once an hour. It does not persist raw transcripts;
it stores compact records in `~/.claude/cully-logs/memory.jsonl` and tracks
processed files in `memory-state.json` so unchanged sessions are skipped.

```bash
cully memory             # latest compact summaries
cully memory --json      # JSONL for another system to consume
cully memory --scan      # scan now, then print memory
cully memory auth --json # query summaries by text
```

## Commands

| Command | Purpose |
|---|---|
| `cully install` | Auto-detect present coding agents and register their integrations |
| `cully install codex` | Add the `AGENTS.md` pointer, shared skill, native Codex status fields, and `/prompts:cully` prompt |
| `cully install cursor` | Add the shared skill and native Cursor `/cully` project command |
| `cully install all` | Install all detected integrations, or explicitly target Claude, Codex, and Cursor |
| `cully uninstall` | Remove Claude Code cully settings and transient state |
| `cully uninstall codex` | Remove the managed Codex pointer, prompt, and status fields |
| `cully uninstall cursor` | Remove the managed Cursor `/cully` command; keep the shared skill |
| `cully statusline` | Render the cully status line for Claude Code, Codex, Cursor, or another agent |
| `cully analyze` | Run the `Stop` hook analyzer |
| `cully list` | Show numbered suggestions |
| `cully apply N` | Accept suggestion N - updates agent instructions, MCP, skills |
| `cully systems` | Synoptic view of hooks, agents, MCP, skills, graphify |
| `cully checklist <topic>` | Procedure for `context`, `budget`, `search`, and related topics |
| `cully plan` | Session route, cost index, deviation |
| `cully status` | Deferred items |
| `cully debrief [session]` | Post-session summary |
| `cully memory [query]` | Retrieve compact background session memory |
| `cully memory --json` | Emit memory as JSONL for other systems |
| `cully memory --scan` | Scan sessions immediately before retrieval |
| `cully daemon start` | Start persistent advisor daemon |
| `cully daemon stop` | Stop advisor daemon |
| `cully daemon status` | Show daemon state and queue depth |
| `cully version` | Print the installed version |

## Controls

| Variable | Effect |
|---|---|
| `CULLY_ANALYZE_DISABLE=1` | Disable advisor analysis; keep the status line |
| `CULLY_ANALYZE_PROMPTS=0` | Omit recent prompt text from analyzer signals |
| `CULLY_DEBUG=1` | Write debug logs to `~/.claude/cully-logs/.cully-debug.log` |
| `CULLY_DISPLAY` | `minimal`, `full` (default), or `debug` |
| `CULLY_COST_INDEX` | `eco`, `normal` (default), or `perf` |
| `CULLY_ALERT_CHIME=1` | Terminal bell when context crosses 90% |
| `CULLY_MEMORY_DISABLE=1` | Disable hourly background memory scans |
| `CULLY_MEMORY_INTERVAL` | Override scan interval, e.g. `30m` or `3600` |
| `CULLY_CLAUDE_SESSION_DIR` | Override Claude transcript scan root |
| `CULLY_CODEX_SESSION_DIR` | Override Codex session scan root |
| `CULLY_CURSOR_SESSION_DIR` | Override Cursor session scan root |
| `CLAUDE_CONFIG_DIR` | Use a different Claude config directory |
| `CODEX_HOME` | Use a different Codex config directory |
| `CURSOR_CONFIG_DIR` | Use a different Cursor config directory |
| `CULLY_VERSION` | Pin installer downloads to a release tag |

## Develop

```bash
go build ./...
go test ./... -race
```

Release by pushing a tag such as `v0.1.0`; GitHub Actions builds the prebuilt
macOS and Linux binaries.

## License

Cully is licensed under [Apache 2.0](../../LICENSE).
