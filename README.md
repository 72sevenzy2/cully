# Buddy

Buddy is shared work memory for coding agents. It provides a remote,
OAuth-protected MCP service and a Buddy skill that tells agents how to capture
and retrieve work across projects and devices. The service stores data and
performs search; the connected agent's model interprets it and writes answers.
The VM does not run a language model.

## Remote MCP service

Buddy uses PostgreSQL with full-text search and pgvector. The service exposes
`buddy_log`, `buddy_search`, `buddy_recent`, `buddy_get`, `buddy_update`,
`buddy_delete`, and `buddy_projects`. Entries use normalized GitHub repository
URLs as project keys and `mcp` as the theme. Each entry can retain the assistant,
summary, approach, outcome, issue, learning, next steps, and tags.

The app runs on the Buddy VM and Caddy routes the existing
`workspace.mcpruntime.org` HTTPS host to `/buddy/mcp`; the Buddy MCP's internal
HTTP path remains `/mcp`. Deployment uses Docker Compose and requires
`BUDDY_DB_PASSWORD`. Copy
`.env.example` to `.env`, set a unique long password, then run:

```sh
docker compose up -d --build
```

Do not expose PostgreSQL publicly. Put the MCP endpoint behind HTTPS and the
configured OAuth authorization server. The server URL defaults to
`https://workspace.mcpruntime.org/buddy/mcp`; its OAuth resource identifier
must be registered with the authorization server before agents can connect.

## Agent skill

Install `SKILL.md` as the `buddy` skill in each coding agent. Configure that
agent's remote MCP client to use the HTTPS endpoint above with OAuth. The skill
normalizes the current Git remote, records substantive work and useful
blockers/lessons, and searches shared memory before answering history questions.
See `clients/README.md` for per-agent configuration and setup.

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

The service uses PostgreSQL full-text search immediately. It also installs
pgvector and creates an HNSW cosine index. A client can pass a 1536-dimensional
vector to `buddy_log`/`buddy_update` and `query_embedding` to `buddy_search`.
Buddy stores and searches vectors but does not generate them or run a model.
Without client-supplied vectors, search uses PostgreSQL full-text search. Agent
Flightdeck integration is deferred until the Buddy API and data model mature.

## Local data safety

Do not commit personal databases, environment files, API tokens, OAuth secrets,
or exported memory. `.gitignore` excludes these by default.
