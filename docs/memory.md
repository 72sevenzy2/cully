---
title: Use Cully memory
description: Save useful notes and find them later from connected coding agents.
---

# Use Cully memory

After [starting Cully on your laptop](/hosting) or [connecting to a team server](/agents), your agent can save and find notes through Cully MCP tools.

Try asking your agent to save a short work summary, then ask it to find that summary in a later session. The Cully skill guides the agent on when a record is useful. Session transcripts are not uploaded automatically; a note is saved when the agent calls `cully_log`.

## What you can save

A record has a summary, the assistant that wrote it and a `personal` or `company` section. Work records can include a GitHub project URL, what changed, the result, a lesson and next steps. Personal records can cover goals or routines without a project URL. Avoid putting credentials or full transcripts in notes.

Both sections belong to the same owner. `company` is a label for organizing your notes; it does not share them with coworkers. On a default private server, the fixed owner configured by setup owns every record. With OAuth, each signed-in user's verified identity is the owner.

## Find and change notes

| Tool | Use it to |
| --- | --- |
| `cully_log` | Save a structured note. |
| `cully_recent` | See recent notes. |
| `cully_search` | Find text in saved notes. |
| `cully_recall` | Find semantically related notes through Mem0. |
| `cully_get` | Read one note by ID. |
| `cully_update` | Correct a note. |
| `cully_delete` | Remove a note. |
| `cully_projects` | List projects with recent activity. |

PostgreSQL holds the source notes. `cully_search` and `cully_recent` read those records directly. Mem0 is part of the standard Cully stack and provides semantic candidates for `cully_recall`; Cully checks them against live, owned records before returning them. A new or edited note may take a little time to appear in recall, while text search remains available.

## Record details

`cully_log` requires a summary, assistant and section. Entry types are `work`, `issue`, `learning` and `decision`; `work` is the default. Optional fields include project URL, category, approach, outcome, issue, learning, next steps and tags. GitHub project URLs normalize to `https://github.com/owner/repo`.

Text fields allow up to 8,000 characters and records allow up to 20 tags of 64 characters. Search can filter by project, section, category, entry type and time. Changes through `cully_update` affect only supplied fields. See [Mem0 recall](/mem0) for its indexing behavior.
