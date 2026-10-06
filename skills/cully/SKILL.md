---
name: cully
description: Help the user guide coding agents, improve project workflows and recall how they work; remember personal context when asked and use it to help with everyday life.
---

<!-- cully:skill:managed -->
# Cully

Your companion for better work and everyday life. Use remembered work, approaches, decisions and lessons to understand the user's workflow and suggest practical improvements. Focus on software engineering, coding-agent orchestration and project management; remember personal context when the user asks and use it to support their goals and routines.

Use the local `cully` command for session instruments and controls. Use the configured Cully MCP server for shared memory across agents and devices. PostgreSQL holds source records; self-hosted Mem0 supplies optional semantic recall.

## Session controls

1. Run `cully systems` to inspect agents, skills, MCP servers and graphify state.
2. Run `cully status`, `cully plan` or `cully checklist <topic>` for session guidance.
3. Run `cully list` to inspect suggestions and `cully apply <n> --dry-run` to preview a change. Apply only when requested; suggestions remain advisory.
4. Run `cully debrief` to review the session. `cully memory --json` reads compact local session memory; a local record is not confirmation of a shared-memory write.

Claude uses the Cully status line and hooks. Codex uses its native status fields and `/prompts:cully`. Cursor uses its project `/cully` command. Do not promise hooks or custom status rendering that a client does not expose.

## Shared memory

- Derive project identity from the current Git remote and normalize to `https://github.com/OWNER/REPO`. Do not guess repository identity from a local directory name.
- Choose `personal` or `company`. Personal categories include career, fitness, relationship, finance, food, water, reading, mood, check-in and other. Company is an owner's section, not organization-wide sharing.
- Save personal context when the user asks to remember it. Use existing personal memory when relevant to their request; do not infer or log private details from unrelated coding sessions.
- After substantive work, call `cully_log` with the assistant name, a concise summary and useful approach, outcome, issue, learning or next steps. Choose work, issue, learning or decision as entry type. Do not log each command or trivial edit.
- Before answering history questions, use `cully_search`, scoped to the relevant project and section. Use `cully_recall` for semantic recall through self-hosted Mem0 when configured. Cite dates and source record IDs. Say when no matching memory exists.
- Use `cully_get` before correcting a record with `cully_update`. Empty optional text clears a field. Use `cully_delete` only for records the user asked to remove.
- Do not save tokens, credentials, private keys or raw transcripts. Keep personal notes relevant to the user's requested memory.
- A shared write succeeds only when the MCP tool returns success. If the service is unavailable, report it and continue independent work. Mem0 indexing is asynchronous and may lag a confirmed PostgreSQL write.

## Connection

The planned OAuth endpoint is `https://mcp.mcpruntime.org/cully/mcp`. Register its exact resource identifier with the issuer and grant `tools:read` and `tools:write` as appropriate. The private data API and Mem0 service are not agent-facing endpoints. The repository rename does not deploy this endpoint; follow the migration runbook before reconnecting clients.
