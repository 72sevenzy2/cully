---
title: Cully documentation
description: Install Cully, start memory on your laptop, and use it with your coding agent.
---

# Cully docs

Cully helps Claude Code, Codex and Cursor keep useful context between sessions and improve the session you are in. Its memory stack runs on your laptop, with Mem0 to find related notes even when you ask in different words. Installed session hooks ask your connected agent to find relevant prior work and save concise summaries of substantive tasks. Cully does not upload raw transcripts.

## How the parts fit together

| Part | What it does |
| --- | --- |
| Local advisor and agent integrations | Starts with the installer, shows available session signals, and offers workflow suggestions you can preview before applying. It can keep guiding you if the memory server is temporarily unavailable. |
| Cully memory | `cully setup codex` starts Cully MCP, the private data API, PostgreSQL and self-hosted Mem0. Your connected agent uses the MCP tools to save and retrieve owned notes across sessions and agents. |

For yourself, the default stack runs on your laptop with one stable owner and no sign-in. For a company team, an ops team can run the services on its platform and add [OAuth](/oauth) so each person has private records. See [self-hosting](/hosting) and [team deployment](/team-deployment) for the two paths.

## Set up Cully

For Codex:

```sh
curl -fsSL https://cully.net/install.sh | sh -s -- --agent codex
```

With Docker Compose and Python 3 installed, start memory and connect the same agent:

```sh
cully setup codex
```

Use `claude` or `cursor` in place of `codex` in both commands if needed. Agent setup starts the advisor daemon automatically. Restart your agent and work on a task; Cully will prompt the agent to look up useful notes and save a short result for later. In Codex, review and trust Cully's hooks in `/hooks` when asked. You do not need to clone the repository or choose database passwords. Follow the [quickstart](/quickstart) for the full sequence.

## Guides

| What you want to do | Guide |
| --- | --- |
| Install Cully and continue work across sessions | [Quickstart](/quickstart) |
| Save and find notes | [Memory](/memory) |
| Use session guidance | [Local advisor](/advisor) |
| Ship with a clear session trail and keep context focused | [Session optimization](/session-optimization) |
| Run the services on your laptop | [Self-hosting](/hosting) |
| Connect to an existing server | [Connect an agent](/agents) |
| Let multiple people sign in | [OAuth setup](/oauth) |

Manual service settings and system details are in the [configuration reference](/configuration) and [architecture](/architecture).

For project work, see [development](/development), [website and docs hosting](/website), the [roadmap](/roadmap) and the [changelog](https://github.com/mcp-runtime/cully/blob/main/CHANGELOG.md).
