---
name: buddy
description: Use during coding work to record substantive changes, blockers, and lessons in the shared Buddy MCP memory, and when asked to search or manage prior project notes.
---

# Buddy

Buddy is shared memory for the user's Personal and Company work across coding agents. Use the configured remote Buddy MCP server for reads and writes; it is the source of truth on every device.

## Shared work memory

- Derive `project_url` from the current Git remote (prefer `origin`) and normalize it to `https://github.com/OWNER/REPO`. Do not guess a project when the directory is not a GitHub checkout; ask once or omit the log if the user does not want to identify it.
- Set theme to `mcp` (the service currently applies this default).
- Every new entry must set `section` to `personal` or `company`. Ask a short clarification if the user's intent is unclear. Use a category for personal notes (`career`, `fitness`, `relationship`, `finance`, `food`, `water`, `reading`, `mood`, `check-in`, or `other`); personal entries do not need a project URL. Company work belongs in `company`; the user says it is done on `pune-devel-04`, with company/project details pending. Keep the two sections separate in search and summaries unless the user asks for both.
- At the end of a substantive coding session, write one concise `buddy_log` entry with the assistant name, what changed, the approach, result, encountered issue and resolution, learned insight, and next step when available. Include only fields that add useful context.
- Also log a meaningful blocker or reusable discovery when it happens; do not create a record for every command, trivial edit, or conversation turn.
- Never store credentials, tokens, private keys, or full conversation transcripts. Personal details may be saved in the `personal` section when the user asks; keep them concise and relevant to the requested memory.
- Before answering questions about prior project work, call `buddy_search`; filter by normalized project URL when relevant. Cite dates and record IDs or summaries from results, and say when no matching memory was found.
- Use `buddy_get` to inspect a record before changing it. Use `buddy_update` for corrections. Use `buddy_delete` only when the user explicitly asks to remove that specific record.
- Do not claim a write succeeded unless `buddy_log` or `buddy_update` returns success. If Buddy is unavailable, report that and continue the user's coding work without silently substituting a local-only log.

## Record shape

Choose `entry_type` from `work`, `issue`, `learning`, or `decision`. Useful records answer: what happened, why this approach was chosen, what worked or failed, and what should happen next. Keep summaries independently understandable because the agent may retrieve one entry without surrounding conversation.
