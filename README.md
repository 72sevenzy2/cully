<p align="center">
  <img src="assets/brand/cully-logo.png" alt="Cully" width="420">
</p>

# Cully

**Your companion for better work and everyday life.**

[Website](https://cully.net) · [Documentation](https://docs.cully.net)

Cully remembers what you do and how you work. It helps you guide coding agents, manage projects, and spot ways to improve your workflow. When you ask it to remember personal things, it can help with your goals, routines and everyday life too.

| Component | What it does |
| --- | --- |
| `cully` | Local advisor, agent setup, session status and suggestions |
| `cully-mcp` | Shared memory tools; private single-owner mode by default, optional OAuth |
| `cully-data` | Private PostgreSQL API and Mem0 projection worker |

Works with Claude Code, Codex and Cursor. It brings local session guidance, workflow improvement and shared memory into one Go repository. PostgreSQL stores authoritative records; self-hosted Mem0 does semantic indexing and recall when configured.

## Get started

From this checkout, with Go 1.26 or newer:

```sh
go build -o ./build/cully ./cmd/cully
./build/cully install
./build/cully status
./build/cully mcp add --agent codex
```

Licensed under [Apache 2.0](LICENSE).
