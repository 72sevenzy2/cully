# Cully combination and optimization roadmap

Cully combines Flightdeck's agent-session controls and Buddy's durable memory
as one Go project. It helps the user orchestrate coding agents and projects,
understand their working habits and suggest improvements. Personal context is
remembered when the user asks.

## Implemented

| Capability | Current design |
| --- | --- |
| Agent setup and session status | Local CLI with supported Claude/Codex/Cursor surfaces |
| Suggestions and applying improvements | Local advisor, numbered suggestions and reviewed apply plans |
| Personal and project records | Shared typed memory service with authenticated owner scoping |
| Search | PostgreSQL full-text, optional pgvector and reciprocal-rank hybrid fusion |
| Semantic recall | Self-hosted Mem0 with source-record hydration and stale/deleted filtering |
| Updates and deletion | Source transaction plus durable Mem0 projection jobs and retry |
| Hosted/self-hosted agent connection | One `cully mcp add` invocation per agent and configurable URL |
| Service runtime | Go MCP and private Go data API, bounded database pool and HTTP deadlines |
| VM migration | Explicit transactional migration preserving existing record IDs and owner subjects |

The redundant transcript scanner and local JSONL memory feature were removed.
Status combines session and integration views. Internal hooks/workers are hidden
from the public CLI. Durable memory has one source of truth: PostgreSQL, indexed
asynchronously by self-hosted Mem0. No old product aliases are maintained.

## Deployment

Complete the [existing VM migration](migration.md), provision the separately
hosted Mem0 runtime, register the Cully OAuth resource and verify client sign-in.
The purchased website/docs domains do not change the MCP audience. Publish the
website at cully.net and documentation at docs.cully.net after configuring their
routes. [Hosting](hosting.md) separates client setup from operator configuration.

## Further optimization

1. Split local session controls into agent adapters, session analysis, suggestions
   and installation packages, retaining one shared memory domain.
2. Make advisor jobs immutable, coalesce repeated jobs by session, and use bounded
   concurrent workers after measuring throughput and cross-project isolation.
3. Measure Go service latency, PostgreSQL plans and allocation rates. Preserve
   index-friendly nearest-neighbor retrieval and calibrated rank fusion.
4. Add explicit, authenticated capture of selected work summaries if needed.
   Define consent, project identity, idempotency, ownership and deletion semantics
   before adding any background upload. Do not reinstate blanket transcript scans.
5. Consider Mem0 fact extraction only after defining source-to-fact attribution,
   edit/delete propagation and model configuration. Current projection is infer=false.

Each optimization needs evidence of improved behavior plus isolation, migration
and outage tests. Local advisor availability must not depend on remote recall.
Existing personal/company sections remain owner-scoped; team-wide memory requires
an explicit membership and authorization design.
