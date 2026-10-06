---
title: Host Cully on your laptop
description: Start Cully's private memory service with Docker Compose and connect your coding agent.
---

# Host Cully on your laptop

Cully's memory service runs on your laptop for the default single-user setup. Your connected agents use it to save and find notes across sessions. You do not need OAuth, a domain name or an identity provider for this setup.

## One-command full stack

1. Install and start Docker with the Compose plugin. Install Python 3 and the [Cully CLI](/installation) if you have not already.
2. Run setup for your agent:

   ```sh
   cully setup codex
   ```

   Use `claude` or `cursor` instead of `codex`. Run `cully setup` without an agent name if you only want to start the services.
3. Wait for `Cully MCP is configured at ...` to appear, then restart your agent. Setup registers the MCP connection, installs the Cully skill and starts the advisor daemon if it is not already running.
4. Work on a substantive task. Your connected agent is prompted to find relevant notes and save a concise work summary for later sessions. The [memory guide](/memory) explains what gets saved.

The command downloads the matching Cully release's Docker files, starts PostgreSQL, Mem0, the private data API and the MCP server, and creates the database schema. No repository checkout is needed. It generates service credentials and a stable single-user owner in `~/.cully/config.json`; keep that file private and back it up with your Docker volumes. Editable stack settings live in `~/.cully/self-hosted/config/.env`. You do not need to edit either file for the default laptop setup.

The default MCP URL is printed by setup and binds to `127.0.0.1`. Only programs on that machine can reach it through that address. Anyone who can reach a no-OAuth endpoint can use its memory tools, so keep this mode on loopback or a trusted private network.

## Connect another agent on the same machine

Use the MCP URL printed by setup:

```sh
cully agent setup claude --mcp-url http://127.0.0.1:8080/mcp
```

Replace the example URL with the one printed on your machine. You can use `codex` or `cursor` instead of `claude`. Restart that agent after setup. See [connect an agent](/agents) for each client's commands.

## When you need sign-in

OAuth is optional for a company team deployment. If your ops team exposes Cully over HTTPS to multiple people, each person can sign in and access their own records. That setup needs an authorization server, identity provider, public hostnames and TLS. Follow the [team deployment guide](/team-deployment) and [OAuth setup](/oauth). A `personal` or `company` section in memory does not share records between people; ownership still follows the signed-in user.

## Run the services another way

The [Compose file](https://github.com/mcp-runtime/cully/blob/main/deploy/self-hosted/compose.yaml) is the reference for the containers and private networks. PostgreSQL holds source records. Mem0 uses a separate pgvector database for semantic recall; its local embedding model does not need an embedding API key. The data API, Mem0 and both databases stay private. The [team deployment guide](/team-deployment) covers other container platforms, and the [configuration reference](/configuration) lists settings for a manual deployment.
