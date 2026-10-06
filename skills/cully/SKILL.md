---
name: cully
description: Use Cully to inspect coding sessions, improve agent workflows, and save or recall useful work and personal context through a configured Cully MCP server.
---

<!-- cully:skill:managed -->
# Cully

Use the installed `cully` CLI for local session guidance and the configured Cully MCP tools for durable records. These are independent: local guidance works without a server, and the MCP connection may use a private single-owner endpoint or OAuth. Do not assume a particular URL, login flow, or Mem0 setup.

## Guide the current work

- Use `cully status` when session state, context pressure, route drift, tool faults, or agent integration matters. It also reports configured agents, MCP servers, skills, and graphify state.
- Use `cully suggestions` to find improvements to project instructions, MCP setup, or skills. Preview a numbered suggestion with `cully apply <n> --dry-run` before changing files or agent settings. Apply a suggestion only when the user requests it or approves the concrete change.
- Explain advice in terms of the user's current task. Avoid interrupting unrelated work with routine status checks.

## Find useful context

- Before answering a question about earlier work, search the configured Cully MCP server with `cully_search`. Filter by project, section, or date when known; broaden a search if a narrow query misses relevant records.
- Use `cully_recall` when semantic similarity would help and Mem0 is configured. It may lag a recent write. `cully_recent` helps when chronology matters; `cully_projects` helps discover active project records; `cully_get` retrieves a specific record.
- Ground claims about past work in the returned records. Include dates or record IDs when they help the user verify the answer, and say when no matching record was found.

## Preserve what matters

- Use `cully_log` after substantive work to keep a concise decision, outcome, lesson, or unresolved issue that is likely to help later. Record personal context when the user asks or it directly supports their request. Do not log every command, transient detail, or full transcript.
- Set `assistant` to the agent in use and `section` to `personal` or `company`. Ask when the section is material and unclear. `company` is still scoped to the configured owner; it is not automatically shared with an organization. Set `entry_type` to `work`, `issue`, `learning`, or `decision` when useful.
- For a GitHub project, derive `project_url` from the Git remote and normalize it to `https://github.com/OWNER/REPO`. Omit it when the repository identity is unknown; do not guess from a directory name. Use a personal category only when it adds useful context.
- Use `cully_get` before correcting a record with `cully_update`. Use `cully_delete` only when the user asks to remove that record.
- Never save credentials, private keys, raw transcripts, or unrelated personal details. Report a write only after the tool confirms it. If MCP is unavailable, explain that the shared record was not saved and continue the user's task.

PostgreSQL holds source records. When enabled, self-hosted Mem0 performs semantic indexing and recall; a successful write can appear there after a short delay. The private data API and Mem0 key are operator settings, never agent configuration.
