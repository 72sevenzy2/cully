# Changelog

## Unreleased

- Combine Buddy and Agent Flightdeck as Cully, preserving both Git histories.
- Rename the local command, agent integrations, MCP tools and configuration to Cully.
- Port the MCP and private memory data services from Python to Go.
- Add self-hosted Mem0 semantic recall with transactional indexing retries, source hydration and deletion propagation.
- Retain personal/company memory, owner isolation, PostgreSQL full-text search, optional pgvector search and IST timestamps.
- Add explicit data-preserving migrations, VM cutover documentation, Go CI, database E2E and server container builds.
- Keep deployment manual while preparing the future cutover on the existing Buddy VM.

Historical Flightdeck releases are documented in [the imported changelog](docs/flightdeck/CHANGELOG.md).
