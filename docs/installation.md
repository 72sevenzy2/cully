# Installation

Cully's local CLI can be built from source now. The repository also includes a release installer for macOS and Linux; it needs a release with CLI archives before the one-line command can succeed.

## Local CLI

Build from the Cully checkout using Go 1.25 or newer:

```sh
go build -o ./build/cully ./cmd/cully
./build/cully install
```

The installer detects configured agents, adds their supported session integrations and installs the combined Cully skill. The installer starts the local advisor daemon. Target an agent explicitly with `cully install claude`, `cully install codex`, `cully install cursor`, or `cully install all`. Restart the agent after configuration changes.

Use `cully status` to inspect integrations, `cully suggestions` to review suggestions and `cully apply 1 --dry-run` to preview a change. The full command list is in [local advisor](advisor.md).

## Release installer

After the next Cully release publishes CLI archives, install from the repository script:

```sh
curl -fsSL https://raw.githubusercontent.com/mcp-runtime/cully/main/install.sh | bash
```

The script selects a prebuilt archive for macOS or Linux on amd64 or arm64, installs the binary and runs `cully install` for detected agents. Set `CULLY_VERSION` to a tag to install a specific release. The current `v0.2.0` and `v0.2.1` GitHub releases have no downloadable CLI assets, so use the source build until the new release is published. Inspect the [installer source](https://github.com/mcp-runtime/cully/blob/main/install.sh) before piping it to a shell.

## Shared memory connection

The planned endpoint is `https://mcp.mcpruntime.org/cully/mcp`. It becomes available after the future deployment. Configure one coding agent at a time. For a server with OAuth enabled, add `--oauth` to print sign-in instructions:

```sh
cully mcp add --agent codex --url http://127.0.0.1:8080/mcp
cully mcp add --agent claude --url http://127.0.0.1:8080/mcp
cully mcp add --agent cursor --url https://my-server.example/mcp
# For a server that requires OAuth:
cully mcp add --agent codex --url https://my-server.example/mcp --oauth
```

With no `--agent`, setup proceeds only if exactly one agent is detected. There is no `--agent all`. `--url` selects a self-hosted or other operator-managed endpoint; the command's default URL selects the planned public Cully service, which is configured for OAuth and therefore needs `--oauth` for sign-in instructions. Setup adds a user-level connection and preserves unrelated client settings. An existing Cully entry with a different endpoint is preserved: edit that entry explicitly before switching.

No-OAuth servers connect directly. For OAuth servers, Codex signs in with `codex mcp login cully`; Claude Code uses `/mcp`; Cursor uses its MCP settings. See [connect an agent](agents.md), [OAuth deployment](oauth.md), or the [full client reference](https://github.com/mcp-runtime/cully/blob/main/clients/README.md).

`cully install` configures the local advisor, hooks, commands and skill. `cully mcp add` configures the remote connection. Neither command deploys `cully-mcp`, `cully-data`, PostgreSQL or Mem0. [Server operators](hosting.md) configure those components.

In OAuth mode, read tools require `tools:read`; logging, editing and deletion require `tools:write`. The issuer must grant those scopes for the configured resource. In no-OAuth mode, private network access is the boundary and all tools use the configured single owner.

The combined skill is also available at `skills/cully/SKILL.md`. Local controls work without a remote connection; shared memory requires the deployed MCP server. Agents do not need the private data API token or Mem0 API key.

## Existing installations

The product uses the new `cully` command, `cully_*` tools and `CULLY_*` configuration. No old aliases are provided. At cutover, remove recognized old integrations with their original uninstallers, then install Cully and reconnect the new MCP server. Preserve user-owned agent configuration. Follow [the VM migration guide](migration.md) before changing running services.
