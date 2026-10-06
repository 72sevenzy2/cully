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

Works with Claude Code, Codex and Cursor. It brings local session guidance, workflow improvement and memory into one Go repository. The standard self-hosted stack runs PostgreSQL for authoritative records and Mem0 for semantic indexing and recall.

## Get started

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex
```

Then run `cully setup codex` to start Cully's memory stack on your laptop and connect Codex. See the [quickstart](https://docs.cully.net/quickstart) for prerequisites and other agents.

Licensed under [Apache 2.0](LICENSE).
