---
name: buddy
description: Use to record, find, update, and recall coding-agent work across GitHub projects through the shared Buddy MCP server, including approaches, outcomes, blockers, and lessons learned.
---

# Buddy

Buddy is shared work memory for the user's coding agents. Use the configured remote Buddy MCP server for cross-device work logs. Use the local CLI only when the user specifically asks to work with local personal-life logs.

## Shared work memory

- Derive `project_url` from the current Git remote (prefer `origin`) and normalize it to `https://github.com/OWNER/REPO`. Do not guess a project when the directory is not a GitHub checkout; ask once or omit the log if the user does not want to identify it.
- Set theme to `mcp` (the service currently applies this default).
- At the end of a substantive coding session, write one concise `buddy_log` entry with the assistant name, what changed, the approach, result, encountered issue and resolution, learned insight, and next step when available. Include only fields that add useful context.
- Also log a meaningful blocker or reusable discovery when it happens; do not create a record for every command, trivial edit, or conversation turn.
- Never store credentials, tokens, private keys, personal data unrelated to the project, or full conversation transcripts. Keep enough technical detail to make the reasoning and lesson reusable.
- Before answering questions about prior project work, call `buddy_search`; filter by normalized project URL when relevant. Cite dates and record IDs or summaries from results, and say when no matching memory was found.
- Use `buddy_get` to inspect a record before changing it. Use `buddy_update` for corrections. Use `buddy_delete` only when the user explicitly asks to remove that specific record.
- Do not claim a write succeeded unless `buddy_log` or `buddy_update` returns success. If Buddy is unavailable, report that and continue the user's coding work without silently substituting a local-only log.

## Record shape

Choose `entry_type` from `work`, `issue`, `learning`, or `decision`. Useful records answer: what happened, why this approach was chosen, what worked or failed, and what should happen next. Keep summaries independently understandable because the agent may retrieve one entry without surrounding conversation.

## Local personal-life CLI

Only use the existing `buddy.py` local SQLite commands when the user explicitly asks about personal-life tracking. They store data on that machine and do not sync to the remote MCP server. See `README.md` for the CLI usage.
