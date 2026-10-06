# Website and documentation

Cully's website is **https://cully.net**. Documentation will be published at
**https://docs.cully.net**. The Markdown guides in this repository are the source
for the documentation; hosting is planned and has not been deployed by this merge.

## DNS on the existing Buddy VM

The existing Buddy VM is `103.181.176.61`. On 2026-10-06,
`workspace.mcpruntime.org` resolved to that address; the deployment reference also
identifies it as the Buddy VM. SSH inspection was unavailable with the current key.

| Type | Name | Value |
| --- | --- | --- |
| A | `@` | `103.181.176.61` |
| A | `docs` | `103.181.176.61` |
| CNAME | `www` (optional) | `cully.net` |

Before publishing, configure separate Caddy routes for `cully.net` and
`docs.cully.net`, serving the website and generated docs respectively. Preserve
the existing workspace and Keycloak routes. Allow HTTPS certificate issuance and
verify both domains after DNS propagation. Remove conflicting A/AAAA records if
they target another host.

These domains serve the website and docs. The MCP resource remains
`https://mcp.mcpruntime.org/cully/mcp`; changing a website domain does not change
OAuth audiences or deploy the Cully services. Follow the [VM migration
guide](migration.md) for the future service cutover.
