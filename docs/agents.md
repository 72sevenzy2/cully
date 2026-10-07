---
title: Connect an agent
description: Connect Claude Code, Codex or Cursor to Cully and its memory tools.
---

# Connect an agent

The [quickstart](/quickstart) starts Cully and connects your first agent. To connect another agent or use an existing Cully server, give that agent the MCP URL printed by setup.

<InstallCommand template="cully setup --agent {agent} --mcp-url http://127.0.0.1:8080/mcp" />

Pick your agent and use the URL printed by your server. Restart your agent after setup. In Codex, review and trust the Cully hooks in `/hooks` when prompted. For an OAuth-enabled HTTPS server, append `--oauth` and sign in using the steps below.

| Agent | Local Cully controls | OAuth sign-in, when enabled |
| --- | --- | --- |
| Claude Code | Status line and `/cully suggestions` | Use `/mcp`. |
| Codex | Run `cully codex` for a live status view, or use `/prompts:cully suggestions`. | Run `codex mcp login cully`. |
| Cursor | Project `/cully suggestions` or ask Cursor to check suggestions | Use Cursor's MCP settings. |

The [local advisor guide](/advisor) shows what each integration can display. The [memory guide](/memory) explains what the MCP tools save and retrieve.

### Claude Code

Run `cully setup --agent claude --mcp-url URL`, using the URL printed by your private stack or supplied by your team. Restart Claude Code. Cully installs a status line, session and stop hooks, a `/cully` command and a skill. If the server uses OAuth, add `--oauth` to setup and use `/mcp` to sign in.

Claude analysis jobs use the shared background worker through the Claude CLI. The worker consults configured Cully memory read-only and can research a concrete tool/documentation gap; the foreground session owns durable handoff writes.

### Codex

Run `cully setup --agent codex --mcp-url URL`, then restart Codex. Cully installs a session-start hook, an asynchronous tool-count hook, a `/prompts:cully` prompt, managed `AGENTS.md` guidance and the skill. Codex's native footer settings remain yours. Review Cully's hooks in `/hooks` when Codex asks you to trust them. With an OAuth server, add `--oauth` and run `codex mcp login cully`.

For the single live Cully display in Codex, start `cully codex` (pass any normal Codex options after `codex`). The default pane groups Model & activity, Project & usage and Cully MCP in three aligned columns at 120 columns or wider; narrower screens stack the groups. Two prioritized advisor comments and an Open advisor control follow the instruments. Open the full advisor and press Tab for all session instruments. The panel uses the terminal's native background, semantic colors and sparse icons; wrapped advice has a badge and hanging indent. The compact panel grows only enough for its bounded preview while leaving at least 12 Codex rows on normal screens and eight on short screens; terminals shorter than 12 rows give Codex the whole screen. Its reserved height is retained until a terminal resize, avoiding repeated conversation reflow as advice clears. The panel shows phase, project/branch, native model and reasoning effort, a context gauge with used/remaining percentages, input/output tokens, quota/fast-mode instruments, tool/search/edit/check/error counts, verification state, elapsed time, daemon health and controls. The preview wraps each of two comments across at most two rows; short panels prioritize context and the first comment. Press Ctrl+] / F6 to open all advice, select by mouse or Up/Down, and use Enter to preview an action. Exiting the wrapper stops its Codex child; the launcher does not create a detached tmux session.

The launcher requests native Codex footer instruments for that child process and displays them in Cully's status view. It reads only the rendered footer, without opening a transcript or saving screen text. Your Codex configuration file stays unchanged. Missing model/context values show waiting for the native footer; missing token/quota values show unavailable. Cost and cache usage are explicitly unavailable from Codex. Working-tree additions/deletions and tracked-file counts come from Git and include staged and unstaged tracked changes against HEAD, rather than Claude session change counters. Short panels prioritize context and the first recommendation, and show an enlarge-terminal hint when space permits. Setup installs an asynchronous `PostToolUse` hook that saves only bounded recent tool counters, not prompts, commands or tool output. Codex advice cautions at 25% context remaining or less and warns at 10% or less. It also flags repeated explicitly observed failures, broad searching and edits awaiting a check. Direct Bash commands and Codex `exec_command` calls support search/check classification; formatting alone does not satisfy verification. Orchestrated calls such as `functions.exec` remain generic unless their inner events are supplied separately, and arbitrary output text does not establish a tool failure. Missing tool signals show awaiting signals rather than implying a healthy workflow. See [advisor controls](/advisor) for the other agent surfaces.

The Codex wrapper also dispatches bounded session signals to the shared worker using an ephemeral Codex CLI process. Its findings are combined with immediate context/tool warnings; worker availability does not replace those local checks. The worker uses configured Cully read-only tools and a targeted web step only when evidence calls for research. Its availability/source status is separate from verification that live tools succeeded.

### Cursor

Run `cully setup --agent cursor --mcp-url URL`, then restart Cursor. Cully installs local continuity hooks, a project `/cully` command and the skill. With an OAuth server, add `--oauth` and sign in from Cursor's MCP settings. User-level hooks installed on your laptop do not run in Cursor cloud agents; connected cloud agents can still use the Cully MCP tools.

Cursor's continuity stop hook can dispatch coarse session metadata to the same shared worker through `cursor-agent` print/ask mode. It does not read transcripts or assume Claude/Codex context or quota instruments. The worker is instructed to recall relevant Cully notes read-only and research concrete gaps; durable handoff notes remain the foreground session's responsibility.

## If you already installed the advisor

Use `cully mcp add --agent codex --url URL` to add only the MCP connection. Add `--oauth` when the server requires sign-in. If you installed Cully before continuity hooks were available, rerun `cully setup --agent codex` (with your agent name and `--mcp-url URL` if using an existing server) to refresh the skill and hooks. Setup preserves unrelated agent configuration and does not replace a Cully connection that points at another URL. Run `cully status` to inspect the local integration.

Running `cully setup` starts the standard local memory stack and advisor, and installs integrations for detected agents. Use `--agent AGENT` to select one or `--agent all` for all supported agents. With `--mcp-url URL`, setup connects to that existing server and starts the local advisor without starting Docker services. Setup does not sign you into an OAuth server automatically. For server requirements, see [self-hosting](/hosting) and [team deployment](/team-deployment).

For detailed client configuration, see the [client reference](https://github.com/mcp-runtime/cully/blob/main/clients/README.md).
