# Changelog

## Unreleased

- Accept GitHub source archives with global PAX headers so the published CLI can download and start its matching self-hosted stack.
- Clarify that `cully setup` prepares configuration and starts the full local stack; `--prepare` stops before starting services for team configuration.
- Add `cully uninstall` for local Docker and agent cleanup, with explicit `--purge-data` for memory deletion.
- Combine local agent guidance and shared memory as Cully.
- Rename the local command, agent integrations, MCP tools and configuration to Cully.
- Port the MCP and private memory data services from Python to Go.
- Add self-hosted Mem0 semantic recall with transactional indexing retries, source hydration and deletion propagation.
- Store authoritative records in plain PostgreSQL with owner isolation, full-text search and IST timestamps; use Mem0 for semantic recall.
- Add an explicit fresh-database schema, VM cutover documentation, separate CI checks, database E2E and server container builds.
- Add separate website/docs delivery and a gated personal-service release workflow.
