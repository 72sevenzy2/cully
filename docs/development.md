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

Separate workflows provide independent checks and badge targets:

| Workflow | Check |
| --- | --- |
| Lint | Go formatting and vet, shell syntax, Compose configuration |
| Unit Test | Go and deployment-controller unit tests |
| Integration Test | PostgreSQL-backed owner isolation and OAuth MCP tests |
| E2E Test | Disposable MCP, data API and PostgreSQL containers; live tool write and read |
| Trivy | MCP, data, website and docs container vulnerability reports |
| CodeQL | Go, Python and JavaScript/TypeScript analysis |

These run on pushes and pull requests. The website and docs retain separate build/deploy workflows for their own paths. The actual personal VM instance is verified during release deployment after a tag is published and the cutover gate is enabled.

Build server images with `docker build -t cully-mcp .` and `docker build -f Dockerfile.data-api -t cully-data .`. GoReleaser builds independent CLI, MCP and data archives from a Cully release tag. After release, the [personal deployment workflow](personal-deployment.md) publishes data and Mem0 images and deploys the isolated personal stack only when its cutover gate is enabled.

Keep domain validation in `internal/memory`, SQL in `internal/store/postgres`, protocol handlers in `internal/transport`, and initialization in `internal/app`. Follow the [architecture](architecture.md) and [roadmap](roadmap.md) when extending session capture or synchronization.
