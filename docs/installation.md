# Installation

## Install and set up an agent

Run the repository installer and choose the agent you use:

```sh
curl -fsSL https://raw.githubusercontent.com/mcp-runtime/cully/main/install.sh | bash -s -- --agent codex
```

Use `--agent claude` or `--agent cursor` instead. The installer installs the Cully CLI, local advisor, and Cully skill. Restart your agent afterward. If you already have a Cully MCP server, connect it in the same command:

```sh
curl -fsSL https://raw.githubusercontent.com/mcp-runtime/cully/main/install.sh | bash -s -- --agent codex --mcp-url https://mcp.example.com/mcp --oauth
```

Omit `--oauth` when your MCP server does not require sign-in. The installer and `cully setup` configure the agent; they do not deploy the server. The agent needs only its MCP URL and whether OAuth is enabled. The operator keeps database and service credentials.

If your shell cannot find `cully` afterward, use the binary path printed by the installer or add that directory to `PATH`.

## Start a self-hosted Cully server

From a Cully checkout, run the Docker setup after installing the CLI:

```sh
cd deploy/self-hosted
./setup.sh codex
```

This creates private service credentials, starts PostgreSQL, Mem0, the data API and MCP, then installs the skill and registers the new MCP URL for Codex. For OAuth with your organization's identity provider, prepare the [MCP Auth connector](/oauth#self-hosted-docker-with-mcp-auth), then run `./setup.sh --oauth mcp-auth codex`. The setup script derives the MCP URL from your configured hostname; no separate agent URL flag is needed. See [Docker self-hosting](/hosting) for prerequisites.

## After installation

Use `cully status` to inspect the local integration. `cully setup codex --mcp-url URL` can add the skill and MCP connection together later. If the skill is already installed and you only need a new MCP connection, use `cully mcp add --agent codex --url URL` (and `--oauth` when needed). See [connect an agent](/agents) for sign-in steps.

`CULLY_VERSION` selects a release tag when running the installer. Inspect the [installer source](https://github.com/mcp-runtime/cully/blob/main/install.sh) before running it.
