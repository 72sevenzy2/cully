# Cully

Cully combines live coding-agent session controls with shared personal and project memory. It brings Agent Flightdeck and Buddy together in one Go repository, preserving both Git histories.

The local `cully` command shows session instruments, suggests useful controls, discovers skills and MCP integrations, and records compact session activity. The remote memory service stores substantive work, decisions, issues, lessons and personal notes across agents and devices. Self-hosted Mem0 adds semantic recall; PostgreSQL remains the source of truth.

## Components

| Binary | Purpose |
| --- | --- |
| `cully` | Local CLI, status line, agent hooks, suggestions, explicit apply, daemon and compact session memory |
| `cully-mcp` | OAuth-protected MCP memory service using the official Go SDK |
| `cully-data` | Private HTTP data API, PostgreSQL/pgvector access, migrations and Mem0 projection worker |

Cully's own services are Go binaries. Self-hosted Mem0 runs separately using its upstream runtime. Agent models interpret memory; Cully does not need a central language model. Mem0 projection uses `infer=false` to preserve source records and avoid additional fact-extraction model calls; embeddings still require the configured Mem0 embedder.

## Local session controls

```sh
go build -o ./build/cully ./cmd/cully
./build/cully install
./build/cully systems
./build/cully list
./build/cully apply 1 --dry-run
./build/cully memory --json
```

Claude receives the rich status line and Stop/SessionEnd hooks. Codex receives native status fields and `/prompts:cully`; Cursor receives a project `/cully` command. The installer preserves user-owned agent configuration. Full command and environment documentation is in [the session-control guide](docs/flightdeck/README.md).

Release archives and `install.sh` use the new Cully names. The source build above works before the first Cully binary release is published. No `cockpit` or `buddy_*` aliases are provided.

## Shared memory

Tools: `cully_log`, `cully_search`, `cully_recent`, `cully_get`, `cully_update`, `cully_delete`, `cully_projects`, and `cully_recall` for self-hosted Mem0.

Entries belong to an authenticated OAuth subject and the `personal` or `company` section. Company is a section label, not shared team access. Records retain project URL, category, assistant, summary, approach, outcome, issue, learning, next steps and tags. Full-text search works without embeddings; optional 1536-dimensional vectors support hybrid search. Timestamps require explicit offsets and are returned in Asia/Kolkata time.

Mem0 writes are queued transactionally with source changes. Failed projections retry; updates replace the source projection and deletes remove it. Semantic recall hydrates current owner-scoped PostgreSQL records, suppressing deleted or stale projections. `cully-data reindex` queues existing records when enabling Mem0. The API key and owner mapping belong to the data service, not agent clients.

The CLI's local transcript summaries and remote memory are separate capture paths in this release. The shared skill tells agents when to log substantive work remotely. Automatic local-to-remote session sync remains a follow-up in [the architecture plan](docs/architecture.md).

## Future VM deployment

The intended deployment target is the VM currently running Buddy. Keep its PostgreSQL volume and records. Run Cully's private Go data API and connect it to self-hosted Mem0 there; the OAuth MCP workload can remain in MCP Runtime. Deployment is manual and the new `/cully/mcp` endpoint is not activated by a repository rename.

Read [the migration runbook](docs/migration.md) before using Compose, updating Caddy or reconnecting clients. [Client configuration](clients/README.md) documents the intended Cully endpoint. [Mem0 setup](docs/mem0.md) describes the self-hosted REST contract.

## Development

```sh
go build ./...
go vet ./...
go test ./... -race -count=1
```

CI runs PostgreSQL/pgvector integration tests, including data preservation during the table rename, owner isolation, text/vector/hybrid search and durable Mem0 indexing. Local database tests require `CULLY_TEST_DATABASE_URL` pointing to a disposable pgvector database. They create and remove generated test schemas; never point them at production.

The architecture and phased optimization plan is in [docs/architecture.md](docs/architecture.md). The existing session-control implementation is kept together under `internal/cully` initially; the new memory services use separate domain, transport and store packages.

## Source attribution

Agent Flightdeck was transferred to the mcp-runtime organization and imported without squashing history. Its MIT license is retained in [licenses/flightdeck-MIT.txt](licenses/flightdeck-MIT.txt). Buddy history is retained as the destination repository's history; this consolidation does not assign a new license to previously unlicensed source.
