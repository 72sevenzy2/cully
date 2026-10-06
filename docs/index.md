---
title: Cully documentation
description: Install Cully, connect your coding agents, and understand local guidance and shared memory.
---

# Make every session count

Cully helps you work with coding agents across projects and devices. The local CLI watches the session signals your agent exposes, shows useful warnings, and suggests improvements you can review. The optional MCP service stores durable records that the same owner can retrieve from another connected agent. Self-hosted Mem0 handles semantic indexing and recall when configured.

| I want to… | Start here |
| --- | --- |
| Try the local CLI | [Quickstart](/quickstart) |
| Set up Claude Code, Codex, or Cursor | [Connect an agent](/agents) |
| Save and find durable notes | [Shared memory](/memory) |
| Run my own memory service | [Self-hosting](/hosting) |

## Two parts that work together

**Local advisor.** `cully setup` configures supported integrations on your machine. `cully status` shows session state, and `cully suggestions` offers changes you can preview before applying. It works without a server. See the [local advisor guide](/advisor).

**Shared memory.** `cully mcp add` registers a remote MCP connection for one agent. Cully tools can save and search your own personal or project records. PostgreSQL holds the source records; optional self-hosted Mem0 handles semantic indexing and recall. A single-user server needs no login by default; operators can enable OAuth for per-user access.

::: info Hosted service status
The Cully hosted MCP endpoint is planned but has not yet been cut over. Local advisor features work now. To use shared memory today, deploy the server yourself and pass its URL to `cully mcp add --url`.
:::

## Choose a path

1. [Install and inspect your local setup](/quickstart).
2. [Connect the coding agent you use](/agents).
3. [Explore the memory tools](/memory) or [run your own service](/hosting).

Browse the [architecture](/architecture), [configuration](/configuration), [OAuth deployment](/oauth), [personal deployment](/personal-deployment), and [website hosting](/website) guides when you operate Cully.
