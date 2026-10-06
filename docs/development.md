# Development

Use Go 1.25 or newer. All Cully production entry points build from the root module; upstream Mem0 is deployed separately.

```sh
gofmt -w cmd internal migrations skills
go vet ./...
go build ./...
go test ./... -race -count=1
```

Unit and transport tests run without a database. PostgreSQL tests require `CULLY_TEST_DATABASE_URL` pointing to a disposable pgvector database. They create and remove generated schemas; use a separate test database, never production.

```sh
CULLY_TEST_DATABASE_URL=postgresql://test:test@localhost:5432/cully_test \
  go test ./internal/store/postgres -race -count=1 -v
```

Database coverage includes every memory operation, owner isolation, text/vector/hybrid search, personal records, source-table/index migration, Mem0 retries/deletion and authenticated MCP-to-data-API-to-PostgreSQL E2E.

CI provisions `pgvector/pgvector:pg17`, runs the Go checks and builds both server containers. Mem0 tests use its documented REST contract; the actual self-hosted VM instance is verified during deployment.

Build server images with `docker build -t cully-mcp .` and `docker build -f Dockerfile.data-api -t cully-data .`. GoReleaser builds independent CLI, MCP and data archives from a Cully release tag. After release, the [personal deployment workflow](personal-deployment.md) publishes the data image and deploys your personal Cully service only when its cutover gate is enabled.

Keep domain validation in `internal/memory`, SQL in `internal/store/postgres`, protocol handlers in `internal/transport`, and initialization in `internal/app`. Follow the [architecture](architecture.md) and [roadmap](roadmap.md) when extending session capture or synchronization.
