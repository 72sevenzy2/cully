# Installation

## Install and set up an agent

Run the repository installer and choose the agent you use:

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex
```

Use `--agent claude` or `--agent cursor` instead. The installer installs the Cully CLI, local advisor, and Cully skill. Restart your agent afterward. If you already have a Cully MCP server, connect it in the same command:

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex --mcp-url https://mcp.example.com/mcp --oauth
```

Omit `--oauth` when your MCP server does not require sign-in. The agent needs only its MCP URL and whether OAuth is enabled. The operator keeps database and service credentials.

If your shell cannot find `cully` afterward, use the binary path printed by the installer or add that directory to `PATH`.

### If installation is slow

The default installer downloads a prebuilt CLI, shows download progress, and limits the binary download to two minutes. A stalled or failed download exits with a retry message; it does not automatically compile Cully. Check access to GitHub release downloads if this step fails.

To build from source explicitly, install Go 1.26 or newer and run:

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --from-source --agent codex
```

Source builds can take several minutes on the first run while Go downloads its toolchain and dependencies. The installer prints build progress. `CULLY_VERSION` selects the source tag; the default source build uses `main`.

## Start a self-hosted Cully server

With Docker Compose and Python 3 installed, the Cully CLI can start the full stack and connect your agent:

```sh
cully setup codex
```

Setup downloads the matching Cully release's Docker files into `~/.cully/self-hosted/releases/`, creates private service credentials, starts PostgreSQL, Mem0, the data API and MCP, and registers the MCP URL and Cully skill for Codex. No checkout is needed. It keeps editable deployment settings in `~/.cully/self-hosted/config/.env` and generated credentials in `~/.cully/config.json`. Use `claude` or `cursor` instead of `codex`, or omit the agent to start only the services. OAuth is off by default. For a team deployment, prepare the [MCP Auth connector and identity provider](/oauth), then run `cully setup --oauth`. See [self-hosting](/hosting) for prerequisites and configuration.

## After installation

Use `cully status` to inspect the local integration. `cully agent setup codex --mcp-url URL` can add the skill and MCP connection together later. If the skill is already installed and you only need a new MCP connection, use `cully mcp add --agent codex --url URL` (and `--oauth` when needed). See [connect an agent](/agents) for sign-in steps.

`CULLY_VERSION` selects a release tag when running the installer. Inspect the [installer source](https://github.com/mcp-runtime/cully/blob/main/install.sh) before running it.
