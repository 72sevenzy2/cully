# Buddy

Buddy is shared Personal and Company memory for coding agents. It provides a remote,
OAuth-protected MCP service and a Buddy skill that tells agents how to capture
and retrieve work across projects and devices. The service stores data and
performs search; the connected agent's model interprets it and writes answers.
The VM does not run a language model.

## Remote MCP service

Buddy uses PostgreSQL with full-text search and pgvector. The service exposes
`buddy_log`, `buddy_search`, `buddy_recent`, `buddy_get`, `buddy_update`,
`buddy_delete`, and `buddy_projects`. Entries use normalized GitHub repository
URLs as project keys and `mcp` as the theme. Every entry belongs to the `personal`
or `company` section, which can be selected when logging, searching, or reviewing
recent activity. Personal entries can use categories such as career, fitness,
relationships, finance, food, water, reading, mood, or check-in; they do not need
a project URL. Company entries can be tied to a GitHub repository. Each entry
can retain the assistant, summary, approach, outcome, issue, learning, next
steps, and tags.

Entries are isolated by the authenticated OAuth subject; the same identity can
retrieve its entries from each connected agent.

The remote service is split into `buddy_mcp/settings.py` (configuration and
time-zone policy), `auth.py` (token verification and subject lookup),
`validation.py` (input checks), `repository.py` (authenticated HTTP client),
`postgres_repository.py` (owner-scoped SQL), `data_api.py` (VM-side HTTP API),
and `server.py` (MCP tool handlers and application startup).
`buddy_service.py` remains a compatibility entry point. Personal check-ins are
short and practical: capture a win or concern the user brings up, and one useful
next priority. Do not turn an ordinary log into a questionnaire.

The PostgreSQL database and the data API run on the Buddy VM. MCP Runtime hosts
the Buddy MCP server, which calls the VM API over HTTPS; it has no database
credentials or dependency on Runtime's internal services. Postgres is on the
Compose-only `private` network and is never published. Caddy routes
`/buddy-data/*` to the API and allows only the MCP Runtime node's egress IP;
the API also requires `BUDDY_DATA_API_TOKEN`. The allowlisted address must be
updated if that node's public egress IP changes. The API intentionally disables
interactive docs and emits no request access logs. Caddy also rewrites Buddy's
OAuth protected-resource metadata URL to an alias under `/buddy/mcp`, which is
the path Runtime routes to the Buddy server.

Deploy the VM-side database and API with Docker Compose. Copy
`.env.example` to `.env`, set independent random values for the database
password and API token, then run:

```sh
docker compose up -d --build
```

Deploy the MCP container in MCP Runtime with `BUDDY_DATA_API_URL` set to
`https://workspace.mcpruntime.org/buddy-data` and the same
`BUDDY_DATA_API_TOKEN` configured on the VM. Do not provide the MCP container
with `BUDDY_DATABASE_URL` or `BUDDY_DB_PASSWORD`. The MCP endpoint remains
behind HTTPS and the configured OAuth authorization server; its OAuth resource
identifier must be registered with the authorization server before agents can
connect.

## Timestamp behavior

Buddy returns `occurred_at`, `created_at`, `updated_at`, and project activity
times as ISO 8601 timestamps with the `+05:30` offset (India Standard Time,
Asia/Kolkata). Inputs to the remote MCP tools must include a timezone offset;
timestamps without one are rejected. PostgreSQL stores these values as
`timestamptz`, which preserves the instant rather than the submitted timezone
label. The service converts database values to IST when it returns them.

For `occurred_at`, pass an ISO timestamp with an offset, for example
`2026-09-23T15:00:00+05:30`.

## Agent skill

Install `SKILL.md` as the `buddy` skill in each coding agent. Configure that
agent's remote MCP client to use the HTTPS endpoint above with OAuth. The skill
normalizes the current Git remote, records substantive work and useful
blockers/lessons, and searches shared memory before answering history questions.
See `clients/README.md` for per-agent configuration and setup.

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
