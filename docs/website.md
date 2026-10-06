# Website and documentation hosting

Cully's website is **https://cully.net** and its docs are **https://docs.cully.net**. They have separate CI, images, containers, deployment workflows and rollback histories. Changes under `site/` deploy the website; changes under `docs/` deploy the docs.

## Build

Use Node.js 20 or newer:

```sh
npm ci
npm run site:build
```

The combined build produces `dist/site` and `dist/docs`. Run `npm run website:build` or `npm run docs:build` to build one site. To preview the docs locally, run `npm run site:preview`. You can serve `dist/site` with any static file server. Do not serve the repository root: it includes source files and deployment notes.

The docs source is Markdown in `docs/`. `docs/.vitepress/config.mts` defines the public navigation. Every Markdown guide under `docs/` is built, including architecture and website hosting. Product and documentation links assume the two production domains. Local previews can follow the corresponding links in the sidebar.

## Publish on the existing VM

`Dockerfile.website` packages only the website as `ghcr.io/mcp-runtime/cully-web:<commit>`. `Dockerfile.docs` packages only the docs as `ghcr.io/mcp-runtime/cully-docs:<commit>`. Their Compose files run `cully-web` and `cully-docs` separately on the existing `workspace_workspace` network. The VM's Docker Caddy routes `cully.net` to `cully-web:8080` and `docs.cully.net` to `cully-docs:8080`. Nginx resolves clean docs URLs such as `/quickstart` to generated HTML. No public container ports are required.

`website-deploy.yml` runs for website inputs; `docs-deploy.yml` runs for docs and package inputs. Each workflow builds and smoke-tests its own image, publishes it to GHCR, deploys only its own Compose project, and verifies its own public revision marker. Pull requests run the build checks without publishing. The restricted SSH controller keeps independent state and restores the previous image if startup or public verification fails. Its temporary registry token is not retained on the VM.

### GitHub configuration

| Setting | Type | Purpose |
| --- | --- | --- |
| `CULLY_WEB_SSH_KEY` | Secret | Dedicated website deployment private key |
| `CULLY_WEB_KNOWN_HOSTS` | Secret | Pinned VM SSH host key |
| `CULLY_WEB_HOST` | Variable | VM IP, currently `103.181.176.61` |
| `CULLY_WEB_SSH_USER` | Variable | `root`, restricted to the deployment controller |
| `CULLY_WEB_SSH_PORT` | Variable | `22` |
| `CULLY_WEB_DEPLOY_ENABLED` | Variable | Enables the production deploy jobs |
| `CULLY_SITE_DEPLOY_V2_ENABLED` | Variable | Enables the separate website and docs controller |
| `CULLY_DOCS_ROUTE_V2_ENABLED` | Variable | Enables public docs route verification |

The `website-production` environment records both site deployments, while separate concurrency groups let either site deploy independently. `GITHUB_TOKEN` is provided by Actions for GHCR publishing/pulling; no personal registry token is required.

### VM controller

The current VM has separate `cully-web` and `cully-docs` Compose projects. Its Caddy configuration routes `cully.net` to `cully-web:8080` and `docs.cully.net` to `cully-docs:8080`. The controller is installed by `deploy/bootstrap-website.sh` and accepts only the website and docs deployment actions through its forced SSH command. The website state is in `/opt/cully-web/state.json`; docs state is in `/opt/cully-web/docs-state.json`.

To rebuild the controller on a replacement VM, copy `deploy/` and the dedicated deployment public key there, then run `sh deploy/bootstrap-website.sh /path/to/deployment-key.pub` as root. Install the two Caddy routes from `deploy/Caddyfile.website.fragment`, validate the Caddyfile and reload Caddy. Set the three deployment variables to `true` after the routes and controller are ready. Changes to either site then deploy that site from `main`.

The website domains do not change an operator's MCP resource URL. See the [team deployment guide](team-deployment.md) for the MCP and authorization server relationship.

## Installation copy

The website serves the repository's installer at `https://cully.net/install.sh`. The website build copies `install.sh` into its static image, and an installer change triggers only the website deployment. Keep the displayed `curl -fsSL https://cully.net/install.sh | sh` command aligned with [installation](installation.md).
