# Personal Cully deployment

Your personal Cully deployment has a separate release flow. A `v*` tag first runs GoReleaser. After that job succeeds, `personal-deploy.yml` builds and publishes `ghcr.io/mcp-runtime/cully-data:<tag>`. If `CULLY_PERSONAL_DEPLOY_ENABLED=true`, it deploys the data service on the VM and then updates the MCP workload on MCP Runtime. Website and docs changes do not trigger this flow.

The data service has **not yet been cut over** on the VM. Keep the deployment gate disabled until the database, network, secrets, OAuth and Caddy route are ready. Publishing a release image does not change the database or your running services.

## Prepare the VM once

Follow the [migration runbook](migration.md) to inventory and back up the existing PostgreSQL volume, apply the Cully schema migration explicitly, verify the preserved owner records, and prepare the `cully-data-api-token` Runtime secret. The release controller deploys only `data-api`; it never creates or replaces PostgreSQL or Mem0.

Create `/opt/cully-personal-data/.env` on the VM with mode `600`. Set `CULLY_DATABASE_URL`, `CULLY_DATA_API_TOKEN`, optional Mem0 settings, and `CULLY_DATA_NETWORK` to the existing Docker network that reaches PostgreSQL. The data service also joins `workspace_workspace` so Caddy can reach `data-api:8083`. The database URL must use a hostname reachable on the data network. The [configuration reference](configuration.md) describes the runtime values.

Create a dedicated Ed25519 deploy key and configure the public key with `sh deploy/bootstrap-personal-data.sh /path/to/key.pub` as root. This installs `deploy/compose.personal-data.yaml` and a restricted SSH controller. Its forced command accepts only deploy and rollback requests for release-tagged Cully data images. Keep the private key in `CULLY_PERSONAL_SSH_KEY`, and pin the VM host key in `CULLY_PERSONAL_KNOWN_HOSTS`.

Apply the [Caddy data route](https://github.com/mcp-runtime/cully/blob/main/deploy/Caddyfile.fragment) to the existing `workspace.mcpruntime.org` site, keeping the MCP Runtime egress allowlist current. Prepare OAuth registration for `https://mcp.mcpruntime.org/cully/mcp` and verify the platform token can update the `cully` workload. Stop any old data API container that would share the `data-api` network name before cutover.

| Setting | Type | Purpose |
| --- | --- | --- |
| `CULLY_PERSONAL_DEPLOY_ENABLED` | Repository variable | Enable release deployment only after cutover is ready |
| `CULLY_PERSONAL_HOST` | Repository variable | VM address |
| `CULLY_PERSONAL_SSH_USER` | Repository variable | `root`, restricted to the data controller |
| `CULLY_PERSONAL_SSH_PORT` | Repository variable | SSH port |
| `CULLY_PERSONAL_SSH_KEY` | Secret | Dedicated personal deployment private key |
| `CULLY_PERSONAL_KNOWN_HOSTS` | Secret | Pinned SSH host key |
| `MCP_PLATFORM_API_TOKEN` | Secret | MCP Runtime deployment token |

The workflow uses the `personal-cully` GitHub environment. After cutover, set the gate and dispatch `personal-deploy.yml` with a release tag created after this workflow was added. Later releases deploy automatically after GoReleaser succeeds. The VM controller records the current and previous data tags in `/opt/cully-personal-data/state.json`; it restores the previous data image if startup fails. If the MCP platform update fails after the data service deploys, the workflow requests a data rollback. Database migrations remain an explicit operator step with a verified backup before any release that changes schema.
