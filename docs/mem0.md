# Self-hosted Mem0

Cully targets the self-hosted Mem0 REST server, not Mem0's hosted API. Its Go adapter calls `GET /memories`, `POST /memories`, `POST /search` and `DELETE /memories/{id}` with `X-API-Key` authentication. Endpoint details follow [the upstream server](https://github.com/mem0ai/mem0/blob/c93420c49a6b14c3d446bdb156d96811908fd90a/server/main.py).

The [self-hosted full stack](/hosting#one-command-full-stack) and gated [personal release stack](/personal-deployment) each start a pinned Mem0 REST server with a persistent vector database and history volume. Keep the REST endpoint private to the service network. A separately operated Mem0 instance must expose the REST contract below.

Set `CULLY_MEM0_URL` and `CULLY_MEM0_API_KEY` on `cully-data`. Both are required to enable indexing. These values are never needed in the public MCP workload or coding-agent configuration. The base URL must omit a trailing endpoint path such as `/search`; a reverse-proxy base prefix is supported.

[`deploy/self-hosted/compose.yaml`](https://github.com/mcp-runtime/cully/blob/main/deploy/self-hosted/compose.yaml) builds the pinned upstream Mem0 server, provisions its own persistent PostgreSQL database and history volume, and keeps its REST endpoint private. `./setup.sh` brings it up with Cully's other services. Mem0 maintains its own Python runtime inside its container; the Cully CLI, MCP server and data API have no Python runtime dependency. The setup helper uses Python 3 on the Docker host to create private credentials.

The full-stack image uses FastEmbed with the open source `BAAI/bge-small-en-v1.5` model and 384-dimensional vectors. The model is cached in the image and runs on the host CPU, so indexing does not need an external embedding API. `infer=false` skips Mem0 fact extraction. Operators can configure a different embedder in their Mem0 build; a separate Mem0 instance can still be used through `CULLY_MEM0_URL` and `CULLY_MEM0_API_KEY`.

Each source entry gets an owner-scoped Mem0 projection. The Mem0 user ID is a SHA-256 namespace derived from the Cully owner (fixed single-user owner or verified OAuth subject), and `run_id` is the Cully entry UUID. The projection carries section/project/category and source-update provenance. This is a service-mediated owner boundary; agents never receive the Mem0 service key.

V1 uses `infer=false`: compact, already authored summaries are embedded directly. Automatic fact extraction can be added later with explicit model configuration and a stronger source-to-fact deletion contract. Do not assume the current projection worker consolidates or extracts facts.

After configuring Mem0 for an existing Cully database, queue source records:

```sh
docker compose run --rm data-api reindex
```

Logging remains available during Mem0 outages. Projections retry with bounded exponential backoff. Semantic recall may temporarily miss recently created or updated records until indexing completes; `cully_search` and `cully_recent` continue to read PostgreSQL. Deleted source entries are filtered immediately from recall even before Mem0 cleanup finishes.

The worker replaces one source projection at a time using a PostgreSQL row lock and bounded HTTP deadlines. It does not provide exactly-once delivery to Mem0; retries reconcile existing projections. Source UUIDs are deduplicated during recall. Future optimization can use leases and multiple bounded workers when measurements justify that complexity.
