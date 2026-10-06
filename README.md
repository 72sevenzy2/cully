# Cully

**Your companion for better work and everyday life.**

[Website](https://cully.net) · [Docs](https://docs.cully.net) (site source and build in this repository)

Cully remembers what you do and how you work. It helps you guide coding agents, manage projects, and spot ways to improve your workflow. When you ask it to remember personal things, it can help with your goals, routines and everyday life too.

| Component | What it does |
| --- | --- |
| `cully` | Local advisor, agent setup, session status and suggestions |
| `cully-mcp` | OAuth-protected shared memory tools |
| `cully-data` | Private PostgreSQL/pgvector API and self-hosted Mem0 indexing |

Works with Claude Code, Codex and Cursor. It brings local agent guidance and shared memory into one Go repository. PostgreSQL stores the original records; self-hosted Mem0 adds semantic recall.

## Get started

From this checkout, with Go 1.25 or newer:

```sh
go build -o ./build/cully ./cmd/cully
./build/cully install
./build/cully status
./build/cully mcp add --agent codex
```

Configure shared memory separately using the [installation guide](docs/installation.md). The future deployment target is the existing Buddy VM; repository changes do not activate the new service endpoint.

## Documentation

- [Quickstart](docs/quickstart.md)
- [Install and connect agents](docs/installation.md)
- [Agent-specific setup](docs/agents.md)
- [Local advisor](docs/advisor.md)
- [Shared memory and tools](docs/memory.md)
- [Architecture](docs/architecture.md)
- [Configuration](docs/configuration.md)
- [Self-hosted Mem0](docs/mem0.md)
- [Hosted and self-hosted setup](docs/hosting.md)
- [VM migration](docs/migration.md)
- [Website, docs and DNS](docs/website.md)
- [Development and tests](docs/development.md)
- [Roadmap](docs/roadmap.md)

Licensed under [Apache 2.0](LICENSE).

## Website and docs build

The marketing site is in `site/`; the documentation site is built from `docs/` with VitePress. Run `npm ci && npm run site:build` to generate static files in `dist/site` and `dist/docs`. See [website hosting](docs/website.md) for Caddy routes and deployment notes.
