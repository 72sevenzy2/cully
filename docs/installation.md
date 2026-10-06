---
title: Install Cully
description: Install Cully and start its standard memory stack on your laptop.
---

# Install Cully

Choose the coding agent you use. For Codex:

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex
```

Use `--agent claude` or `--agent cursor` for another agent. The installer adds the Cully CLI, advisor and skill, and starts the advisor daemon in the background. It prints `Advisor started` when that succeeds. If your shell cannot find `cully` afterward, use the binary path printed by the installer or add that directory to `PATH`.

## Start Cully memory

Install and start Docker with the Compose plugin, and install Python 3. Then run:

```sh
cully setup codex
```

Use the same agent name you installed. Setup starts Cully MCP, the data API, PostgreSQL and Mem0, then connects your agent and starts the advisor daemon if it is not already running. Restart the agent after setup. Follow the [quickstart](/quickstart) to save and find your first note.

## Connect to a server already running

If a company runs Cully for you, use its MCP URL when installing:

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex --mcp-url https://mcp.example.com/mcp --oauth
```

Use the URL and sign-in instructions your company provides. Leave off `--oauth` for a private single-user endpoint. You can add the server later with `cully agent setup codex --mcp-url URL`. See [connect an agent](/agents).

## Build the CLI from source

If a prebuilt release cannot be downloaded, check GitHub release access or install from source with Go 1.26 or newer:

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --from-source --agent codex
```

A source build can take several minutes while Go downloads its toolchain and dependencies. You can [inspect the installer](https://github.com/mcp-runtime/cully/blob/main/install.sh) before running it.

## If the download stalls

The default installer downloads a prebuilt CLI from GitHub Releases and shows progress. It stops a binary download after two minutes or when the connection stalls, then prints a retry message; it does not silently switch to a source build. Check access to GitHub Releases and retry, or choose `--from-source` explicitly. Set `CULLY_VERSION` to a release tag when you need a particular version; a source build otherwise uses the current main branch.

## Check and adjust the installation

Run `cully status` to see the advisor, agent integration, MCP connection and continuity hooks. If you installed the advisor first, `cully agent setup codex --mcp-url URL` adds the skill and server connection together. If only the server connection is missing, use `cully mcp add --agent codex --url URL`; add `--oauth` when that server requires sign-in. Cully keeps unrelated agent settings. For client-specific sign-in and commands, see [connect an agent](/agents).
