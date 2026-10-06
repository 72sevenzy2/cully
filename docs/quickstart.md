---
title: Quickstart
description: Install Cully and inspect your first local session.
---

# Quickstart

Get the local advisor running first. It does not need a Cully account or remote service.

## 1. Install the CLI

For now, build from source with Go 1.26 or newer:

```sh
git clone https://github.com/mcp-runtime/cully.git
cd cully
go build -o ./build/cully ./cmd/cully
./build/cully setup
```

The one-line [release installer](/installation#release-installer) is already in the repository and will work when the next release includes CLI archives. It supports macOS and Linux on amd64 and arm64.

## 2. Check your setup

```sh
./build/cully status
./build/cully suggestions
```

Cully detects supported agents, installs its skill and local integration, and starts the advisor. Restart your agent after installation. To target one agent, use `./build/cully setup codex`, `claude`, or `cursor`.

## 3. Preview a suggestion

When `cully suggestions` lists a numbered improvement, inspect it before applying:

```sh
./build/cully apply 1 --dry-run
```

Local guidance is optional and advisory. Cully preserves unrelated, user-owned agent configuration. The [local advisor guide](/advisor) explains the commands and controls.

## Add shared memory when your server is ready

For a self-hosted Docker deployment, run `./setup.sh codex` from `deploy/self-hosted`. It generates private credentials in `~/.cully/config.json` and starts MCP, the data API, PostgreSQL and Mem0 through Docker Compose. With the Cully CLI installed, it also installs the shared skill and registers the local MCP URL. See [hosting](/hosting#one-command-full-stack).

If your server is already running, connect one agent and install the shared skill together:

```sh
./build/cully setup codex --mcp-url http://127.0.0.1:8080/mcp
```

The default single-user server needs no login. If the operator enables OAuth, use a public HTTPS URL and add `--oauth` to the setup command; then run `codex mcp login cully`. See [agent setup](/agents) for Claude Code and Cursor. The default hosted MCP endpoint is planned and is not yet available for new connections. The local advisor works without it.
