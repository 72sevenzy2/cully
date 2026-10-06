---
title: Mem0 and semantic recall
description: Understand how Cully indexes notes with self-hosted Mem0.
---

# Mem0 and semantic recall

Mem0 is part of Cully's standard memory stack. `cully setup codex` starts Mem0 alongside Cully, PostgreSQL and the private data API. You do not need a Mem0 account or embedding API key.

PostgreSQL stores the source note. Cully sends an owner-scoped projection of its authored summary to Mem0, which indexes it by meaning. When an agent uses `cully_recall`, Mem0 finds candidates and Cully checks the live source records before returning them. `cully_search` and `cully_recent` read PostgreSQL directly.

## What to expect

A new or changed note can take a little time to appear in semantic recall because indexing runs in the background. Text search remains available during a Mem0 outage, and the worker retries indexing. Deleted notes are filtered from recall immediately, even if projection cleanup is still pending.

The standard stack uses self-hosted Mem0 with a separate pgvector/PostgreSQL database and persistent history volume. Its FastEmbed model runs locally on the host CPU. Cully sends authored summaries with `infer=false`; it does not ask Mem0 to extract new facts from transcripts.

## Manual deployment

The [Compose file](https://github.com/mcp-runtime/cully/blob/main/deploy/self-hosted/compose.yaml) shows the Mem0 service, database, volumes and private network. If you deploy Cully's data API yourself, point `CULLY_MEM0_URL` at the Mem0 REST base URL and give the data API `CULLY_MEM0_API_KEY`. Mem0's `ADMIN_API_KEY` must match. Do not put these values in an agent's configuration or expose Mem0 REST publicly.

If you add Mem0 to a database with existing notes, queue them for indexing with the data API's `reindex` command. See the [configuration reference](/configuration) and [architecture](/architecture) for service details.
