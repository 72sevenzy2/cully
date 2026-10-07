---
name: cully
description: Use Cully to inspect advisor status and suggestions, improve agent workflows, and save or recall useful project and personal context through its configured MCP server.
---

<!-- cully:skill:managed -->
# Cully

Use the installed `cully` CLI for local session guidance and the configured Cully MCP tools for durable records. `cully setup` installs integrations and starts the advisor alongside the local memory services; use `--mcp-url URL` for an existing server. The connected agent makes memory calls with its own MCP identity; the local advisor does not hold OAuth credentials. Local guidance works without a server. Do not assume a particular URL or login flow.

## Guide the current work

- Use `cully status` when session state, context pressure, route drift, tool faults, or agent integration matters. It also reports configured agents, MCP servers, skills, and graphify state. A running daemon confirms the background process is alive; it does not prove a particular agent's signals are arriving.
- Use `cully suggestions` to find improvements to project instructions, MCP setup, or skills. Preview a numbered suggestion with `cully apply <n> --dry-run` before changing files or agent settings. Apply a suggestion only when the user requests it or approves the concrete change.
- Explain the suggestions that actually appear for the current session. An empty list can mean no issue has been detected or that the agent has not supplied enough signals. Avoid interrupting unrelated work with routine status checks.

## Match advice to the agent

- Claude Code has Cully's native status line and richer live signals for context, repeated tool failures, search effort, and verification gaps.
- Codex shows Cully's live lower pane when launched with `cully codex`. The pane counts bounded tool events and can suggest checking repeated failures, narrowing searches, or verifying edits. Plain `codex` has Cully prompts and continuity hooks but no Cully status line. The pane and background daemon have distinct signal paths; check the pane after an actual tool call when troubleshooting Codex advice.
- Cursor uses its Cully command and continuity integration. Check its configured surface rather than assuming Claude's status line or Codex's pane exists there.
- When asked whether the advisor works, verify `cully version`, `cully status`, and the relevant agent surface. For Codex, check that the PostToolUse hook is active and that a real tool call changes the pane counter. If a hook needs trust review, review its command before enabling it. Describe any untested part plainly.

## Continue work across sessions

- At the start of a substantive task, automatically call `cully_context` with the current project and a focused task query, using its default three-note limit. This returns concise previews; call `cully_get` only for a record whose full details matter. If the server is older and lacks `cully_context`, use one `cully_search` or `cully_recent` call with `limit: 3`.
- If text search misses, try `cully_context` with `mode: semantic` or use `cully_recall` with `limit: 3`. Mem0 recall may lag a recent write. `cully_projects` helps discover active project records. Search again when the task or project changes; avoid repeated broad lookups. Do not ask the user to request memory manually.
- Before answering a question about earlier work, search the configured Cully MCP server with `cully_search`. Filter by project, section, or date when known; broaden a search if a narrow query misses relevant records.
- Ground claims about past work in the returned records. Include dates or record IDs when they help the user verify the answer, and say when no matching record was found.
- Use Git and live service checks for current branch, release, deployment, or process state; memory records are history, not a substitute for a fresh status check.

## Preserve what matters

- Before finishing substantive work, automatically use `cully_log` to save one concise record that answers: what task was this, what changed, how and why was it done, what worked, what failed or was missed, and what should happen next? Include only fields that add useful context. Record a reusable blocker or decision as it happens. Do not log every command, trivial turn, transient detail, or full transcript.
- If the installed Cully hook provides an opaque `session_ref`, include it in `cully_log`. This groups notes from one agent session without storing the native session ID or transcript. Do not invent a reference when no hook supplied one.
- Avoid duplicate records when work spans several turns or a stop hook asks you to check again. If an existing record needs correction, inspect it with `cully_get` and use `cully_update`.
- Set `assistant` to the agent in use and `section` to `personal` or `company`. Ask when the section is material and unclear. `company` is still scoped to the configured owner; it is not automatically shared with an organization. Set `entry_type` to `work`, `issue`, `learning`, or `decision` when useful.
- For a GitHub project, derive `project_url` from the Git remote and normalize it to `https://github.com/OWNER/REPO`. Omit it when the repository identity is unknown; do not guess from a directory name. Use a personal category only when it adds useful context.
- Use `cully_delete` only when the user asks to remove that record.
- Never save credentials, private keys, raw transcripts, or unrelated personal details. Report a write only after the tool confirms it. If MCP is unavailable, explain that the shared record was not saved and continue the user's task.

PostgreSQL holds source records. Mem0 is part of Cully's standard memory stack and performs semantic indexing and recall; a successful write can appear there after a short delay. The private data API and Mem0 key are operator settings, never agent configuration.
