---
title: Use the local advisor
description: Check your session and review Cully suggestions inside your coding agent.
---

# Use the local advisor

Cully checks the session signals your coding agent makes available and suggests practical improvements. It can point out a missing project instruction, a useful skill or an MCP connection. Suggestions are advisory: you review a change before applying it.

The installer runs agent setup, which starts the advisor daemon automatically. After [installing Cully](/installation), restart your agent and check suggestions:

| Agent | In the agent |
| --- | --- |
| Claude Code | Watch the Cully status line or run `/cully suggestions`. |
| Codex | Run `/prompts:cully suggestions` or ask Codex to run `cully suggestions`. |
| Cursor | Run the project `/cully suggestions` command or ask Cursor to check Cully suggestions. |

You can always run these in a terminal:

```sh
cully status
cully suggestions
```

`cully status` shows whether the daemon is running. If setup printed `Advisor unavailable`, run `cully agent setup codex` to retry with your agent name. Claude Code provides live hook signals and a Cully status line. Codex and Cursor use their own supported status displays, commands and skills, so the advice may differ between agents.

## Preview a suggestion

If Cully lists a numbered improvement, preview it with the number shown:

```sh
cully apply 1 --dry-run
```

Review the preview. To apply it, run `cully apply 1` and confirm the proposed change. Cully preserves unrelated agent settings and does not silently replace user-owned configuration.

## Commands

| Command | What it does |
| --- | --- |
| `cully status` | Shows session warnings and integration state. |
| `cully suggestions` | Lists improvements and informational notes. |
| `cully apply <n> --dry-run` | Previews a numbered improvement. |
| `cully apply <n>` | Applies one after confirmation. |
| `cully agent setup AGENT` | Installs or refreshes the local integration. |
| `cully uninstall AGENT` | Removes Cully-managed local integration settings. |

Cully's [memory tools](/memory) keep useful context across sessions. The local advisor can still show session guidance if the memory server is temporarily unavailable.

## Controls

Set these environment variables only when you want to change the defaults:

| Variable | Effect |
| --- | --- |
| `CULLY_ANALYZE_DISABLE=1` | Disable advisor analysis while keeping the status line. |
| `CULLY_ANALYZE_PROMPTS=0` | Exclude recent prompt text from analyzer signals. |
| `CULLY_DISPLAY` | Use `minimal`, `full` or `debug` display. |
| `CULLY_COST_INDEX` | Use `eco`, `normal` or `perf` cost estimate. |
| `CULLY_ALERT_CHIME=1` | Ring a terminal bell at critical context pressure. |
| `CULLY_DEBUG=1` | Enable local debug logs. |

Cully honors `CLAUDE_CONFIG_DIR`, `CODEX_HOME` and `CURSOR_CONFIG_DIR` for custom agent directories. Local snapshots and diagnostics support the advisor; they are not automatically saved as memory records.
