# Cully

**Your companion for better work and everyday life.**

Cully remembers what you do and how you work. It helps you guide coding agents, manage projects, and spot ways to improve your workflow. When you ask it to remember personal things, it can help with your goals, routines and everyday life too.

| Component | What it does |
| --- | --- |
| `cully` | Local status line, session controls, suggestions and compact session memory |
| `cully-mcp` | OAuth-protected shared memory tools |
| `cully-data` | Private PostgreSQL/pgvector API and self-hosted Mem0 indexing |

Works with Claude Code, Codex and Cursor. It combines Agent Flightdeck and Buddy in one Go repository, preserving both histories. PostgreSQL stores the original records; self-hosted Mem0 adds semantic recall.

## Get started

From this checkout, with Go 1.25 or newer:

```sh
go build -o ./build/cully ./cmd/cully
./build/cully install
./build/cully systems
```

Configure shared memory separately using the [installation guide](docs/installation.md). The future deployment target is the existing Buddy VM; repository changes do not activate the new service endpoint.

## Documentation

- [Install and connect agents](docs/installation.md)
- [Session controls](docs/flightdeck/README.md)
- [Shared memory and tools](docs/memory.md)
- [Architecture](docs/architecture.md)
- [Configuration](docs/configuration.md)
- [Self-hosted Mem0](docs/mem0.md)
- [VM migration](docs/migration.md)
- [Development and tests](docs/development.md)
- [Roadmap](docs/roadmap.md)
- [Source history and release archive](docs/archive/flightdeck/README.md)
