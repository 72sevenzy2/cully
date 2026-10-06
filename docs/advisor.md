# Local advisor and session controls

Cully helps you guide coding agents, review session pressure and apply useful
workflow improvements. Shared memory uses Cully MCP, PostgreSQL and Mem0.

## Setup

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex
cully status
```

Use `--agent claude` or `--agent cursor` for another local integration. To add
shared memory, run `cully setup codex --mcp-url URL`; see [agent setup](agents.md)
for sign-in when the server uses OAuth.

| Agent | Local integration |
| --- | --- |
| Claude Code | Command-backed status line, Stop/SessionEnd hooks and `/cully` |
| Codex | Native status fields, `/prompts:cully`, AGENTS.md pointer and skill |
| Cursor | Project `/cully` command and skill |

The installer starts the advisor daemon. Claude supplies rich live hook payloads;
Codex and Cursor use their available native surfaces. Starting the daemon does
not create live hooks in clients that do not expose them. Re-run `cully setup`
to restart a stopped advisor. Inspect its state using `cully status`.

## Commands

| Command | Purpose |
| --- | --- |
| `cully setup [agent]` | Configure local integrations and start the advisor |
| `cully uninstall [agent]` | Remove managed local integration settings |
| `cully status [directory]` | Session warnings, agent setup and advisor state |
| `cully suggestions` | Review numbered improvements and informational notes |
| `cully apply <n> --dry-run` | Preview an improvement |
| `cully apply <n>` | Apply a selected improvement after confirmation |
| `cully mcp add --agent <agent> --url URL` | Configure one MCP connection for an existing agent setup |
| `cully version` | Print the CLI version |

Apply can update project instructions and skills, add relevant project MCP
integrations or run accepted installation commands. Suggestions are advisory.
Review the dry run before applying a change. Existing user-owned agent settings
and unrelated MCP connections are preserved.

Hooks, workers and daemon execution use an internal entry point installed by
Cully. They are not public user commands.

## Controls

| Variable | Effect |
| --- | --- |
| `CULLY_ANALYZE_DISABLE=1` | Disable advisor analysis; keep the status line |
| `CULLY_ANALYZE_PROMPTS=0` | Omit recent prompt text from analyzer signals |
| `CULLY_DEBUG=1` | Enable local debug logs |
| `CULLY_DISPLAY` | `minimal`, `full` or `debug` |
| `CULLY_COST_INDEX` | `eco`, `normal` or `perf` |
| `CULLY_ALERT_CHIME=1` | Terminal bell at critical context pressure |
| `CLAUDE_CONFIG_DIR` | Custom Claude configuration directory |
| `CODEX_HOME` | Custom Codex configuration directory |
| `CURSOR_CONFIG_DIR` | Custom Cursor configuration directory |

Local snapshots, suggestion state and diagnostic counters support the advisor.
Durable personal/project records are managed through the [shared memory tools](memory.md).

Cully is licensed under [Apache 2.0](https://github.com/mcp-runtime/cully/blob/main/LICENSE).
