# Personal Cully deployment

Your personal Cully deployment has a separate release flow. A `v*` tag first runs GoReleaser. After that succeeds, `personal-deploy.yml` builds and publishes tagged Cully data and pinned self-hosted Mem0 images. If `CULLY_PERSONAL_DEPLOY_ENABLED=true`, the restricted VM controller starts two PostgreSQL databases, Mem0 and the data API, runs the Cully schema migration, and then the workflow updates Cully MCP on MCP Runtime. Website and docs changes do not trigger this release flow.

The personal stack has **not yet been cut over** on the VM. Keep the deployment gate disabled until its network, secrets, OAuth and Caddy route are ready. Publishing release images does not change the VM or running services. This Compose project uses new named volumes: plain PostgreSQL for Cully's authoritative records and a separate pgvector database for Mem0. It does not mount or migrate an existing personal database. Historical record import is a separate operation that Cully does not provide.

## Prepare the VM once

Follow the [VM cutover runbook](migration.md) to inventory current routes and prepare recovery. Prepare the `cully-data-api-token` Runtime secret and its OAuth issuer, resource and JWKS settings. The release controller creates fresh `cully-db`, `mem0-db` and `mem0-history` volumes in its own Compose project.

Create `/opt/cully-personal-data/.env` on the VM with mode `600`, using [`deploy/personal-data.env.example`](https://github.com/mcp-runtime/cully/blob/main/deploy/personal-data.env.example). Set distinct database passwords, the URL-encoded Cully database URL, data API token, Mem0 API key and JWT secret, and an embedding-provider key. Mem0 uses OpenAI embeddings in this example, so authored summaries are sent to that provider. The data API joins `workspace_workspace` so Caddy can reach `data-api:8083`; neither database nor Mem0 publishes a port. The [configuration reference](configuration.md) describes the runtime values.

Create a dedicated Ed25519 deploy key and configure the public key with `sh deploy/bootstrap-personal-data.sh /path/to/key.pub` as root. This installs the full personal Compose file, Mem0 database initialization SQL, and a restricted SSH controller. Its forced command accepts only deploy and rollback requests for tagged Cully service images. Keep the private key in `CULLY_PERSONAL_SSH_KEY`, and pin the VM host key in `CULLY_PERSONAL_KNOWN_HOSTS`. Rerun this bootstrap before enabling the release gate if an older data-only controller was installed.

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

The workflow uses the `personal-cully` GitHub environment. After cutover, set the gate and dispatch `personal-deploy.yml` with a release tag created after this workflow was added. Later releases deploy automatically after GoReleaser succeeds. The VM controller records current and previous image tags in `/opt/cully-personal-data/state.json`; it restores the prior service images if startup fails. If the MCP platform update fails, the workflow requests a service rollback. Compose never removes database volumes during rollback. Take a verified backup before any release that changes schema; image rollback does not reverse a database migration.
