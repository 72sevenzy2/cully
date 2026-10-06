# Website and documentation hosting

Cully's public website is intended for **https://cully.net** and its docs for **https://docs.cully.net**. Both sites are static. This repository produces the files; merging a PR does not change DNS or deploy them.

## Build

Use Node.js 20 or newer:

```sh
npm ci
npm run site:build
```

The build produces `dist/site` for the product website and `dist/docs` for the docs. To preview the docs locally, run `npm run site:preview`. You can serve `dist/site` with any static file server. Do not serve the repository root: it includes source files and deployment notes.

The docs source is Markdown in `docs/`. `docs/.vitepress/config.mts` defines the public navigation. Every Markdown guide under `docs/` is built, including architecture, migration, and website hosting. Product and documentation links assume the two production domains. Local previews can follow the corresponding links in the sidebar.

## Publish on the existing VM

Copy each built directory to its own read-only web root, such as `/srv/cully/site` and `/srv/cully/docs`. Use separate Caddy routes for `cully.net` and `docs.cully.net`, as shown in the [Caddy fragment](https://github.com/mcp-runtime/cully/blob/main/deploy/Caddyfile.website.fragment). The docs route needs a fallback from clean URLs such as `/quickstart` to `quickstart.html`. Preserve the existing workspace, Keycloak, and Cully service routes when adding these sites.

Point the apex and `docs` DNS records at the public web host, allow HTTPS certificate issuance, then verify both domains and a deep docs link. On the currently planned VM (`103.181.176.61`), `workspace.mcpruntime.org` resolved to that address on 2026-10-06. Recheck the address before changing DNS.

The website domains do not change the MCP resource URL. The planned hosted MCP endpoint is `https://mcp.mcpruntime.org/cully/mcp`; its service cutover has separate requirements in the [VM migration guide](migration.md).

## Release installer copy

The website defaults to the source build while current GitHub releases have no CLI archives. After the next release publishes the `cully_<os>_<arch>.tar.gz` files and `checksums.txt`, update the website install panel and docs to recommend the one-line installer, then rebuild and publish the static files.
