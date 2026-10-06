# Personal VM cutover

Cully starts as a fresh deployment. The release workflow creates new Cully and Mem0 database volumes; it does not attach, convert or import another product's database. Keep `CULLY_PERSONAL_DEPLOY_ENABLED` unset until the new routes, secrets and OAuth configuration are ready.

## Prepare the VM

Inventory the current Caddy routes and MCP Runtime workload before changing them. Back up any existing service you may need to restore. The [personal deployment guide](personal-deployment.md) lists the new environment file, restricted SSH controller, image registry access, OAuth settings and GitHub variables. Use distinct credentials for the Cully source database and Mem0 database.

The Cully data API uses plain PostgreSQL for authoritative source records. Mem0 has its own pgvector database for semantic recall. The `cully-data migrate` command creates Cully's schema transactionally on a fresh database. It does not upgrade a prior database schema. No automatic record import is provided.

## Deploy and verify

Bootstrap the [personal stack](personal-deployment.md), then enable its deployment gate and dispatch its workflow with a release tag. The controller starts both databases, Mem0 and the data API, runs the Cully schema command, and checks service health before the MCP Runtime update.

Verify a Cully MCP tool write and read for the expected owner, a full-text search, and a Mem0 recall after its projection worker has processed the record. In OAuth mode, verify read and write scopes and the public protected-resource metadata. Keep the Caddy data route private to MCP Runtime's egress addresses.

## Switch clients and recover

Connect agents to the Cully MCP URL using the [agent guide](agents.md). Remove previous managed agent integrations with their own uninstallers if they are still installed. Keep user-owned agent settings intact.

Retire an earlier service only after Cully passes its checks. If the new deployment fails, restore its previous service images or route clients back to the earlier service. Cully database volumes remain separate; image rollback does not reverse schema changes. Any transfer of historical records requires a separate, reviewed export and import process.
