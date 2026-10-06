# Website and documentation hosting

Cully's website is **https://cully.net** and its docs are **https://docs.cully.net**. They have separate CI, images, containers, deployment workflows and rollback histories. Changes under `site/` deploy the website; changes under `docs/` deploy the docs.

## Build

Use Node.js 20 or newer:

```sh
npm ci
npm run site:build
```

The combined build produces `dist/site` and `dist/docs`. Run `npm run website:build` or `npm run docs:build` to build one site. To preview the docs locally, run `npm run site:preview`. You can serve `dist/site` with any static file server. Do not serve the repository root: it includes source files and deployment notes.

The docs source is Markdown in `docs/`. `docs/.vitepress/config.mts` defines the public navigation. Every Markdown guide under `docs/` is built, including architecture, migration, and website hosting. Product and documentation links assume the two production domains. Local previews can follow the corresponding links in the sidebar.

## Publish on the existing VM

`Dockerfile.website` packages only the website as `ghcr.io/mcp-runtime/cully-web:<commit>`. `Dockerfile.docs` packages only the docs as `ghcr.io/mcp-runtime/cully-docs:<commit>`. Their Compose files run `cully-web` and `cully-docs` separately on the existing `workspace_workspace` network. The VM's Docker Caddy routes `cully.net` to `cully-web:8080` and `docs.cully.net` to `cully-docs:8080`. Nginx resolves clean docs URLs such as `/quickstart` to generated HTML. No public container ports are required.

`website-deploy.yml` runs for website inputs; `docs-deploy.yml` runs for docs and package inputs. Each workflow builds and smoke-tests its own image, publishes it to GHCR, deploys only its own Compose project, and verifies its own public revision marker. Pull requests run the build checks without publishing. The shared restricted SSH controller keeps independent state and restores the previous image if startup or public verification fails. Its temporary registry token is not retained on the VM. The apex website check runs when its A records point only to the VM.

### GitHub configuration

| Setting | Type | Purpose |
| --- | --- | --- |
| `CULLY_WEB_SSH_KEY` | Secret | Dedicated website deployment private key |
| `CULLY_WEB_KNOWN_HOSTS` | Secret | Pinned VM SSH host key |
| `CULLY_WEB_HOST` | Variable | VM IP, currently `103.181.176.61` |
| `CULLY_WEB_SSH_USER` | Variable | `root`, restricted to the deployment controller |
| `CULLY_WEB_SSH_PORT` | Variable | `22` |
| `CULLY_WEB_DEPLOY_ENABLED` | Variable | Set `true` after bootstrap and routing are ready |
| `CULLY_SITE_DEPLOY_V2_ENABLED` | Variable | Set `true` after the two-site controller is installed |
| `CULLY_DOCS_ROUTE_V2_ENABLED` | Variable | Set `true` after Caddy points docs to `cully-docs` so public verification runs |

The `website-production` environment records both site deployments, while separate concurrency groups let either site deploy independently. `GITHUB_TOKEN` is provided by Actions for GHCR publishing/pulling; no personal registry token is required.

### One-time VM bootstrap

Copy `deploy/` and the dedicated public key to a private temporary directory on the VM. Run `sh deploy/bootstrap-website.sh /path/to/deployment-key.pub` as root. This installs both site Compose files, the controller and forced SSH command; it does not change memory services. The website keeps its existing `/opt/cully-web/state.json`; docs use `/opt/cully-web/docs-state.json`.

After installing the new controller, set `CULLY_WEB_DEPLOY_ENABLED=true` and `CULLY_SITE_DEPLOY_V2_ENABLED=true`, leaving `CULLY_DOCS_ROUTE_V2_ENABLED` unset. Dispatch `docs-deploy.yml`: it starts and health-checks `cully-docs` while the public docs still use the old combined container. Then change the [Caddy fragment](https://github.com/mcp-runtime/cully/blob/main/deploy/Caddyfile.website.fragment) in `/opt/workspace/Caddyfile` so `docs.cully.net` points to `cully-docs:8080`. Preserve existing routes. Validate with `docker exec workspace-caddy caddy validate --config /etc/caddy/Caddyfile`, then reload with `docker exec workspace-caddy caddy reload --config /etc/caddy/Caddyfile`. Set `CULLY_DOCS_ROUTE_V2_ENABLED=true` and dispatch the docs workflow again to verify the public route. Finally dispatch `website-deploy.yml` to replace the combined website container with its website-only image.

Future changes to either site deploy only that site from `main`. The new workflow actions are rejected by the old controller, so a missed bootstrap cannot replace the combined image with a single-site image.

Point the apex and `docs` DNS records at the public web host, allow HTTPS certificate issuance, then verify both domains and a deep docs link. On the currently planned VM (`103.181.176.61`), `workspace.mcpruntime.org` resolved to that address on 2026-10-06. Recheck the address before changing DNS.

Keep only `103.181.176.61` in the apex A records; remove the parking addresses `3.33.130.190` and `15.197.148.33`. Verify A/AAAA records and reachability on ports 80/443 before certificate issuance. The VM identifies itself as `devbox-2`; the local `devbox2` and `cully-vm` aliases use its configured SSH key.

The website domains do not change the MCP resource URL. The planned hosted MCP endpoint is `https://mcp.mcpruntime.org/cully/mcp`; its service cutover has separate requirements in the [VM migration guide](migration.md).

## Release installer copy

The website defaults to the source build while current GitHub releases have no CLI archives. After the next release publishes the `cully_<os>_<arch>.tar.gz` files and `checksums.txt`, update the website install panel and docs to recommend the one-line installer, then rebuild and publish the static files.
