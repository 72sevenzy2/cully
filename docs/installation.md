# Installation

## Local CLI

Build from the Cully checkout using Go 1.25 or newer:

```sh
go build -o ./build/cully ./cmd/cully
./build/cully install
```

The installer detects configured agents, adds their supported session integrations and installs the combined Cully skill. Target an agent explicitly with `cully install claude`, `cully install codex`, `cully install cursor`, or `cully install all`. Restart the agent after configuration changes.

Use `cully systems` to inspect integrations, `cully list` to review suggestions and `cully apply 1 --dry-run` to preview a change. The full command list is in [session controls](flightdeck/README.md).

When the first Cully binary release is published, the repository's `install.sh` will download the matching macOS/Linux CLI archive. Historical Flightdeck and Buddy releases are not Cully CLI builds; until then, use the source build above.

## Shared memory connection

The planned endpoint is `https://mcp.mcpruntime.org/cully/mcp`. It becomes available after the future deployment and OAuth resource registration. Configure the remote server and complete OAuth using the [per-agent client instructions](../clients/README.md).

Read tools require `tools:read`; logging, editing and deletion require `tools:write`. The issuer must grant those scopes for the new resource.

The combined skill is also available at `skills/cully/SKILL.md`. Local controls work without a remote connection; shared memory requires the deployed MCP server. Agents do not need the private data API token or Mem0 API key.

## Existing installations

The product uses the new `cully` command, `cully_*` tools and `CULLY_*` configuration. No old aliases are provided. At cutover, remove recognized old integrations with their original uninstallers, then install Cully and reconnect the new MCP server. Preserve user-owned agent configuration. Follow [the VM migration guide](migration.md) before changing running services.
