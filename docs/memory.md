# Shared memory

Every record belongs to an owner. In default no-OAuth mode, the operator fixes one owner with `CULLY_MCP_OWNER`; in OAuth mode, the verified token subject becomes the owner. The same owner can retrieve records across connected agents and devices. `personal` and `company` are sections within that owner's memory; company records are not automatically shared with an organization.

## Tools

| Tool | Purpose |
| --- | --- |
| `cully_log` | Save a structured personal or project record |
| `cully_search` | Full-text search over authoritative PostgreSQL records |
| `cully_recall` | Semantic recall through configured self-hosted Mem0 |
| `cully_recent` | Recent entries, optionally filtered by project or section |
| `cully_get` | Retrieve one owned entry |
| `cully_update` | Correct selected fields |
| `cully_delete` | Delete one owned source entry and queue projection cleanup |
| `cully_projects` | List projects with recent activity |

Results use an envelope containing `entry`, `entries`, `projects` or `deleted`, depending on the operation. Read tools require `tools:read`; mutation tools require `tools:write`.

## Record fields

Logging requires a summary, assistant and section. Assistants identify Codex, Claude, Cursor, ChatGPT or other. Entry types are work, issue, learning and decision; work is the default. Optional fields include project URL, category, approach, outcome, issue, learning, next steps and tags.

Project URLs normalize to `https://github.com/owner/repo`; SSH GitHub remotes are accepted. Personal records need no project URL. Categories include career, fitness, relationship, finance, food, water, reading, mood, check-in and other.

Text fields allow up to 8,000 characters; records allow up to 20 tags of 64 characters. Credentials are rejected in note text. Embeddings are managed inside Mem0, never supplied through Cully tool inputs.

`occurred_at` accepts RFC3339 timestamps with an explicit offset. If omitted, logging uses the current instant. Returned timestamps use Asia/Kolkata's `+05:30` offset; PostgreSQL stores instants as `timestamptz`.

Updates change only supplied fields. An empty optional text field clears it; an empty summary is rejected. Get/update of an absent or differently owned record returns no entry; deletion reports whether an owned record was removed.

## Search and recall

`cully_search` requires query text. It supports project, section, category, entry type and since filters. Search/recent limits are bounded to 50; projects is bounded to 100. It remains available when Mem0 is unavailable.

`cully_recall` requires query text and configured Mem0. Indexing is asynchronous: recently edited records may not appear in semantic recall until their projection is updated. Recall checks the live source owner, filters and timestamp before returning a record. Text search and recent activity remain available during Mem0 outages.

See [self-hosted Mem0](mem0.md) for configuration, indexing and deletion behavior.

## Local versus shared capture

Shared memory is written through `cully_log`; the Cully skill tells agents when useful work warrants a record. Local session diagnostics do not become memory records automatically. PostgreSQL stores durable personal/project records; self-hosted Mem0 does semantic indexing and recall when configured.
