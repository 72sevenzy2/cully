# Personal Cully deployment

Your personal Cully deployment has a separate release flow. A `v*` tag first runs GoReleaser. After that succeeds, `personal-deploy.yml` builds and publishes tagged Cully data and pinned self-hosted Mem0 images. The restricted VM controller starts two PostgreSQL databases, Mem0 and the data API, runs the Cully schema migration, and then the workflow updates Cully MCP on MCP Runtime. The deployment job fails if `CULLY_PERSONAL_DEPLOY_ENABLED` is unset. Website and docs changes do not trigger this release flow.

Keep the deployment gate disabled until the VM network, secrets, OAuth and Caddy route are ready. A release with the gate disabled fails its deployment job explicitly. This Compose project uses named volumes: plain PostgreSQL for Cully's authoritative records and a separate pgvector database for Mem0.

## Prepare the VM once

Prepare the OAuth issuer, resource and JWKS settings. The release controller creates `cully-db`, `mem0-db` and `mem0-history` volumes in its own Compose project.

Create `/opt/cully-personal-data/.env` on the VM with mode `600`, using [`deploy/personal-data.env.example`](https://github.com/mcp-runtime/cully/blob/main/deploy/personal-data.env.example). Set distinct database passwords, the URL-encoded Cully database URL, data API token, Mem0 API key and JWT secret. The Mem0 image runs a local open source embedding model, so no embedding-provider key is needed. The data API joins `workspace_workspace` so Caddy can reach `cully-data-api:8083`; neither database nor Mem0 publishes a port. The [configuration reference](configuration.md) describes the runtime values.

## Platform OAuth and identity provider

The personal public MCP uses MCP Runtime's existing `https://auth.mcpruntime.org/mcp-auth` authorization server. MCP Runtime hosts MCP Auth; its selected Keycloak/OIDC connector owns the upstream identity-provider issuer, confidential client ID and secret, callback URI, user claim mapping, and supported MCP scopes. Those provider settings belong to the **platform's MCP Auth connector**, not Cully's `.mcp/servers.yaml` or the VM data stack. Historical Buddy MCP metadata likewise supplied only the MCP Auth issuer and its own resource audience; it did not carry Keycloak client configuration.

For Cully, preserve the platform's working identity-provider connector and add `tools:write` to its `mcp_scopes` alongside `tools:read`. The connector's upstream `scopes` such as `openid` and `profile` are separate from these MCP tool scopes. The Keycloak client must allow the exact `https://auth.mcpruntime.org/mcp-auth/identity/callback` redirect. Keep its client secret in the platform's connector secret, never in Cully metadata or GitHub Actions. The [MCP Runtime OAuth guide](https://docs.mcpruntime.org/mcp-oauth/) describes the connector file and platform setup.

[Cully's Runtime manifest](https://github.com/mcp-runtime/cully/blob/main/.mcp/servers.yaml) specifies the public resource `https://mcp.mcpruntime.org/cully/mcp` and the platform issuer. MCP Runtime reconciles that resource into MCP Auth and injects the resource and issuer into the standalone MCP container. Cully verifies the resulting token against the platform JWKS at `https://auth.mcpruntime.org/mcp-auth/.well-known/jwks.json`. The release workflow checks discovery for both `tools:read` and `tools:write`; if discovery advertises only read access, update the selected platform connector before releasing.

Set `CULLY_DATA_API_TOKEN` as a Cully GitHub Actions secret with the same value as the VM data service. The release workflow injects it into its temporary `.mcp/servers.yaml` after building the image, then deploys the MCP workload through the platform API. MCP OAuth identifies the user; the data API service token authenticates the MCP workload independently. Mem0 uses its separate `CULLY_MEM0_API_KEY`. The Runtime API stores a literal workload environment variable in its server spec, so people with platform server-spec read access can view this personal service token. Rotate it if that access changes.

Create a dedicated Ed25519 deploy key and configure the public key with `sh deploy/bootstrap-personal-data.sh /path/to/key.pub` as root. This installs the full personal Compose file, Mem0 database initialization SQL, and a restricted SSH controller. Its forced command accepts only deploy and rollback requests for tagged Cully service images. Keep the private key in `CULLY_PERSONAL_SSH_KEY`, and pin the VM host key in `CULLY_PERSONAL_KNOWN_HOSTS`.

Apply the [Caddy data route](https://github.com/mcp-runtime/cully/blob/main/deploy/Caddyfile.fragment) to the `workspace.mcpruntime.org` site, keeping the MCP Runtime egress allowlist current. Prepare OAuth registration for `https://mcp.mcpruntime.org/cully/mcp` and verify the platform token can update the `cully` workload. The data service uses the dedicated `cully-data-api` network name.

| Setting | Type | Purpose |
| --- | --- | --- |
| `CULLY_PERSONAL_DEPLOY_ENABLED` | Repository variable | Enable release deployment after VM and OAuth setup is ready |
| `CULLY_PERSONAL_HOST` | Repository variable | VM address |
| `CULLY_PERSONAL_SSH_USER` | Repository variable | `root`, restricted to the data controller |
| `CULLY_PERSONAL_SSH_PORT` | Repository variable | SSH port |
| `CULLY_PERSONAL_SSH_KEY` | Secret | Dedicated personal deployment private key |
| `CULLY_PERSONAL_KNOWN_HOSTS` | Secret | Pinned SSH host key; defaults to the existing `CULLY_WEB_KNOWN_HOSTS` secret for the same VM |
| `MCP_PLATFORM_API_TOKEN` | Secret | MCP Runtime deployment token |
| `CULLY_DATA_API_TOKEN` | Secret | Private MCP-to-data API service token, shared with the VM data service |

The workflow checks the live MCP Auth issuer and JWKS before building service images. It connects the pinned `mcp-runtime` binary to `https://platform.mcpruntime.org` using `MCP_PLATFORM_API_TOKEN`, validates metadata, builds, pushes and deploys the Cully MCP image after the VM data stack passes health checks. It then queries the platform for the deployed service and checks both tool scopes, Cully's public protected-resource metadata and the unauthenticated challenge. A full interactive sign-in and authorized tool call still require an agent after release. The workflow uses the `personal-cully` GitHub environment. After VM setup, set the gate and dispatch `personal-deploy.yml` with a release tag created after this workflow was added. Later releases deploy automatically after GoReleaser succeeds. The VM controller records current and previous image tags in `/opt/cully-personal-data/state.json`; it restores the prior Cully service images if startup fails. If MCP deployment fails, the workflow requests a service rollback; a later OAuth verification failure leaves the new data stack running for diagnosis. Compose never removes database volumes during rollback. Take a verified backup before any release that changes schema; image rollback does not reverse a database migration.
