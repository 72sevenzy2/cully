# Future Buddy VM to Cully cutover

Repository consolidation does not deploy Cully. The current VM and OAuth clients continue using their running configuration until this runbook is executed. The [personal deployment release workflow](personal-deployment.md) stays gated until this cutover is complete.

This runbook covers preserving records from the existing VM database. The new personal release Compose project starts with separate named volumes; enabling its gate does not attach or import the old database automatically. Complete an export/restore and verify owner records before routing clients to that new stack.

## Inventory and backup

Record the existing VM Compose project name, PostgreSQL database/user, actual Docker volume name, Caddy routing and current MCP Runtime manifest. Record image digests and configuration for rollback. Back up PostgreSQL and verify restoration to a separate disposable database. Preserve existing owner subjects; changing the identity provider's subject mapping would disconnect users from their records.

Copy `.env.example` to a private `.env`. Set `CULLY_DB_VOLUME` to the existing Docker volume, and retain the existing database name, role and password. The historical database/role may still be named `buddy`; renaming those storage assets is unnecessary to give Cully a clean product interface. Set a correctly escaped `CULLY_DATABASE_URL` and a service token. Compose uses an external volume so it cannot silently create an empty replacement database.

Do not run two PostgreSQL containers against the same data volume. Stop the old database container before starting the Cully Compose database service, or keep the existing database container and connect the new data API to its actual network/address. Do not change PostgreSQL major versions during this cutover.

## Migrate and verify

Pause old ingestion and stop the old data API. Build the Go services, start the intended PostgreSQL service and run the explicit migration:

```sh
docker compose build
docker compose up -d db
docker compose run --rm migrate
docker compose up -d data-api
docker compose exec data-api /cully-data health
```

The transactional migration renames `buddy_entries` to `cully_entries` when present, preserving IDs, owner subjects, fields and timestamps. It creates the Cully schema-version table and durable Mem0 outbox. If both source and destination tables exist, it refuses to choose one. Rows without an owner also prevent migration; reconcile them explicitly instead of assigning arbitrary ownership.

Compare record counts and representative per-owner record IDs/contents to the backup. Verify personal/company retrieval, timestamp offsets, filters, edits and deletes. The serving process does not run migrations on startup.

Configure the self-hosted Mem0 service as described in [mem0.md](mem0.md), then run `docker compose run --rm data-api reindex` to populate projections. Full-text retrieval remains available while this queue drains.

## Public service and clients

Prepare the new `cully-data-api-token` Runtime secret and `CULLY_DATA_API_URL=https://workspace.mcpruntime.org/cully-data`. Adapt `deploy/Caddyfile.fragment` to the existing site block; review the Runtime egress-IP allowlist rather than blindly assuming it is unchanged.

Register `https://mcp.mcpruntime.org/cully/mcp` with the OAuth issuer and configure grants for `tools:read` and `tools:write`. Deploy `cully-mcp` using `.mcp/servers.yaml`; gateway is disabled because Cully verifies and retains the bearer itself. Validate that the protected-resource metadata advertises the exact public resource even through proxy Host rewrites. Client setup is in the [client reference](https://github.com/mcp-runtime/cully/blob/main/clients/README.md).

Reconnect supported agents under the new Cully name. Install the Cully CLI and shared skill; explicitly remove the old product's managed hooks/prompts/status/config blocks using its old uninstaller before installing the new names. Do not delete user-owned settings or unrelated MCP connections. No aliases are retained.

Only retire old running services/routes after the new service passes its smoke checks. A rollback restores the old image/config and the verified pre-cutover database backup; it is a deployment recovery procedure, not a compatibility layer.
