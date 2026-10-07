# Changelog

## Unreleased

## 0.7.0

- Require a changelog update in every PR through repository instructions, a PR checklist and a dedicated CI check.
- Add a compact full-width Codex pane with two prioritized comments, suggestion/tip counts and an interactive advisor opened by mouse or Ctrl+] / F6.
- Preview and accept advice with mouse/arrow selection and Enter; apply supported shared instruction/skill changes locally or hand task actions to the current coding input without discarding its draft.
- Share owner-scoped memory recall and targeted research across Claude, Codex and Cursor advisor workers, with isolated native CLI adapters and clear startup/source status.
- Add an evidence-based red Messy phase, terminal-aware command guidance, native-background styling and adaptive group spacing while retaining full session instruments.
- Release Codex sessions and process descendants when the wrapper exits or loses its terminal.

## 0.6.1

- Bind Codex pane signals to their SessionStart session when hooks run through a persistent app server, without saving prompts or tool output.
- Document the distinct Claude, Codex and Cursor surfaces, live signal checks, and local CLI setup with a remote MCP/data/Mem0 stack.

## 0.6.0

- Replace `cully agent setup` with one `cully setup` command for local services, detected agent integrations and the advisor; use `--mcp-url` for existing servers.
- Log startup stages and return errors when the advisor or another component fails; keep `--prepare` configuration-only.
- Install project hooks in the caller's directory, and have the installer direct local users to the complete setup command.

## Earlier changes

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
