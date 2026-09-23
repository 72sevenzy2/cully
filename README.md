# Buddy

Buddy is a personal memory and work journal for coding agents. This repository
starts with the local SQLite capture CLI and its Buddy skill. The next milestone
is a shared, authenticated MCP service so Codex, Claude, Cursor, and phone-based
agent clients can record and retrieve work notes.

## Current local CLI

```sh
python3 buddy.py ask --mode quick
python3 buddy.py log career "Finished the API migration; next: document rollout"
python3 buddy.py search "API migration"
python3 buddy.py summary
```

SQLite data is created locally at `memory/buddy_memory.db` by default. That
directory is intentionally excluded from Git. Use `--memory /path/to/file.db`
to select another local database. Buddy never starts or hosts a language model.

## Shared Buddy direction

The shared service will store work records with a normalized GitHub project URL
as the project key, `mcp` as the theme, the coding assistant, a concise summary,
approach, outcome, and next steps. Agents use their own model to turn retrieved
records into plain-English answers. The VM hosts storage and the MCP/API service;
it does not run inference.

The MCP service will provide authenticated, scoped create, read, update, delete,
and search operations. Until that service is deployed, the local CLI and skill
are the working implementation; do not configure clients with a pretend remote
endpoint. Destructive operations will require explicit caller intent. Never log
credentials, tokens, private keys, or full conversation transcripts.

## Client skill

`SKILL.md` contains the Buddy workflow instructions. `openai.yaml` contains
Codex skill metadata. The shared workflow will derive the project key from the
Git remote, normalize SSH and HTTPS forms to `https://github.com/OWNER/REPO`,
and append a brief work record at the end of substantive coding sessions.

## Storage and integrations

The planned remote store is PostgreSQL with full-text search and optional
pgvector support. Search and CRUD run on the service; language-model reasoning
stays with the connected agent. Agent Flightdeck integration is deferred until
the Buddy API and data model have matured.

## Local data safety

Do not commit personal databases, environment files, API tokens, OAuth secrets,
or exported memory. `.gitignore` excludes these by default.
