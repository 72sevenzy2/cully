# Roadmap

Cully provides local session guidance and owner-scoped shared memory through MCP. PostgreSQL stores source records, and Mem0 indexes summaries for semantic recall. See [architecture](architecture.md) for the current design.

## Planned work

1. Verify the [team deployment guide](team-deployment.md) with an agent sign-in, tool write and read, full-text search, and Mem0 recall.
2. Measure advisor latency and project isolation before changing its job scheduling. Coalesce repeated work and use bounded concurrency where measurements justify it.
3. Measure PostgreSQL search plans and Mem0 recall latency. Keep source-record hydration and owner checks in Cully as the system scales.
4. Add opt-in capture of selected work summaries after defining consent, project identity, idempotency, ownership and deletion behavior.
5. Consider richer Mem0 extraction only with a clear source-to-fact attribution and edit/delete contract. Current projections embed authored summaries with `infer=false`.

Changes to storage and authorization need tests for owner isolation, recovery and failure behavior. The local advisor must remain usable when the hosted service is unavailable.
