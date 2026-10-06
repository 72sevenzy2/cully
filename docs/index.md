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

**Local advisor.** The installer or `cully agent setup` configures supported integrations on your machine. `cully status` shows session state, and `cully suggestions` offers changes you can preview before applying. It works without a server. See the [local advisor guide](/advisor).

**Shared memory.** `cully setup codex` starts your own Docker stack and registers its MCP server with Codex. For an already-running server, `cully agent setup codex --mcp-url URL` installs the skill and registers the endpoint. Cully tools can save and search your own personal or project records. PostgreSQL holds the source records; optional self-hosted Mem0 handles semantic indexing and recall. A single-user server needs no login by default; operators can enable OAuth for per-user access.

## Choose a path

| For yourself | For your team |
| --- | --- |
| [Install Cully](/quickstart) for local session guidance. When you want shared memory, [start your own Docker stack](/hosting) and connect your agent. | [Plan a team deployment](/team-deployment) on Docker Compose, Kubernetes or your container platform. Connect [MCP Auth](/oauth) to your organization's identity provider so each person signs in. |

For the service contract, see [configuration](/configuration) and [architecture](/architecture). [MCP Runtime](https://mcpruntime.org) is a platform where you can deploy an MCP server. See [website and docs hosting](/website) for this project's static sites.
