# Website and documentation hosting

Cully's website is **https://cully.net** and its docs are **https://docs.cully.net**. Both are built into one Docker image. Once the VM is bootstrapped and deployment enabled, changes to website/docs inputs on `main` publish both sites automatically.

## Build

Use Node.js 20 or newer:

```sh
npm ci
npm run site:build
```

The build produces `dist/site` for the product website and `dist/docs` for the docs. To preview the docs locally, run `npm run site:preview`. You can serve `dist/site` with any static file server. Do not serve the repository root: it includes source files and deployment notes.

The docs source is Markdown in `docs/`. `docs/.vitepress/config.mts` defines the public navigation. Every Markdown guide under `docs/` is built, including architecture, migration, and website hosting. Product and documentation links assume the two production domains. Local previews can follow the corresponding links in the sidebar.

## Publish on the existing VM

`Dockerfile.website` builds both sites and serves them with an unprivileged Nginx container. `deploy/compose.website.yaml` starts only `cully-web` on the existing `workspace_workspace` network. No public container port is required: the VM's existing Docker Caddy container routes both domains to `cully-web:8080`. Nginx resolves clean docs URLs such as `/quickstart` to their generated HTML files. Buddy, workspace and Keycloak keep their existing routes.

The website workflow builds and smoke-tests the image, validates Caddy/Compose, and publishes `ghcr.io/mcp-runtime/cully-web:<commit>`. Deployment uses a dedicated restricted SSH key; its forced command can only run the website controller. The Actions job's temporary registry token pulls the image without leaving a registry credential on the VM. The controller waits for container health and restores the previous image if startup fails. Public HTTPS checks verify the docs revision marker and a deep docs link. The apex revision is also checked whenever its A records point only to the VM, so stale registrar parking records do not undo a healthy deployment.

### GitHub configuration

| Setting | Type | Purpose |
| --- | --- | --- |
| `CULLY_WEB_SSH_KEY` | Secret | Dedicated website deployment private key |
| `CULLY_WEB_KNOWN_HOSTS` | Secret | Pinned VM SSH host key |
| `CULLY_WEB_HOST` | Variable | VM IP, currently `103.181.176.61` |
| `CULLY_WEB_SSH_USER` | Variable | `root`, restricted to the deployment controller |
| `CULLY_WEB_SSH_PORT` | Variable | `22` |
| `CULLY_WEB_DEPLOY_ENABLED` | Variable | Set `true` after bootstrap and routing are ready |

The `website-production` environment records deployments. `GITHUB_TOKEN` is provided by Actions for GHCR publishing/pulling; no personal registry token is required. PR builds validate sites without deploying. Push deployment and manual dispatch use the dedicated website workflow.

### One-time VM bootstrap

Copy `deploy/` and the dedicated public key to a private temporary directory on the VM. Run `sh deploy/bootstrap-website.sh /path/to/deployment-key.pub` as root. This installs the website Compose file, controller and forced SSH command; it does not change memory services.

Add the [Caddy fragment](https://github.com/mcp-runtime/cully/blob/main/deploy/Caddyfile.website.fragment) to `/opt/workspace/Caddyfile`, preserving existing routes. Validate with `docker exec workspace-caddy caddy validate --config /etc/caddy/Caddyfile`, then reload with `docker exec workspace-caddy caddy reload --config /etc/caddy/Caddyfile`. Caddy runs inside Docker and obtains/renews Let's Encrypt certificates automatically. Its existing persistent `/data` mount retains certificates across restarts. A separate Certbot container is unnecessary.

Set `CULLY_WEB_DEPLOY_ENABLED=true` and dispatch `website-deploy.yml` for the first deployment. Future changes under `site/`, `docs/`, deployment files, the website Dockerfile or package inputs deploy automatically from `main`. The currently active/previous revisions are recorded at `/opt/cully-web/state.json`.

Point the apex and `docs` DNS records at the public web host, allow HTTPS certificate issuance, then verify both domains and a deep docs link. On the currently planned VM (`103.181.176.61`), `workspace.mcpruntime.org` resolved to that address on 2026-10-06. Recheck the address before changing DNS.

Keep only `103.181.176.61` in the apex A records; remove the parking addresses `3.33.130.190` and `15.197.148.33`. Verify A/AAAA records and reachability on ports 80/443 before certificate issuance. The VM identifies itself as `devbox-2`; the local `devbox2` and `cully-vm` aliases use its configured SSH key.

The website domains do not change the MCP resource URL. The planned hosted MCP endpoint is `https://mcp.mcpruntime.org/cully/mcp`; its service cutover has separate requirements in the [VM migration guide](migration.md).

## Release installer copy

The website defaults to the source build while current GitHub releases have no CLI archives. After the next release publishes the `cully_<os>_<arch>.tar.gz` files and `checksums.txt`, update the website install panel and docs to recommend the one-line installer, then rebuild and publish the static files.
