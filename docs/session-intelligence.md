---
title: Session intelligence
description: How Cully turns tool events into a private session journal, loop detection, session health, timelines, handoffs and rescue.
---

# Session intelligence

Session intelligence is how Cully understands what a coding agent is doing. The [Cully terminal](/terminal) and agent hooks record small events. Everything on this page is derived from those events, plus live Git state read at the moment you ask.

```text
Terminal activity
       +
Agent hooks          ──►  Session journal  ──►  Loops · health · replay · handoff · rescue
       +
Live Git state
```

This works for an agent only while it runs inside the Cully terminal. Run `cully run claude`, `cully run codex`, `cully run cursor` or `cully run AGENT` first.

## The session journal

One journal file per session, in your Cully directory, readable only by you. Each line is one event:

| Field | Example |
| --- | --- |
| Time and agent | `10:34`, Claude Code |
| Kind | edit, check, search, other tool, Cully memory call, note, start, end |
| Result | passed or failed |
| File path and operation | `internal/auth/resource.go`, edited |
| Command | `go test` (program and subcommand only) |
| Command hash | a one-way hash, so the same failing command is recognized without keeping it |

It never holds prompts, file contents, command arguments or tool output. It is bounded in size, pruned after 30 days, and path and command recording can be turned off. See [privacy](/privacy) for the full guarantee.

## Loop detection

Cully reports when an agent appears to repeat itself. It states what repeated and never claims to know why.

| Rule | Fires when |
| --- | --- |
| Repeated failure | The same command failed 3 times within 20 events or 15 minutes. A passing run of that command resets it. |
| Edit and failed-check cycles | An edit was followed by a failed check 3 times with no passing check in between. Any passing check resets it. |

When a loop is detected the terminal shows `⚠ Loop 3x` in the health bar, adds a caution to the advisor, and writes one note to the journal for the timeline. The wording stays factual: "Cully noticed something: the same command failed 3 times after edits. Inspect the first failure before editing again."

Cursor's shell hook reports no exit code, so a failed Cursor command is not marked failed and these rules fire less often for Cursor.

## A workflow hint from your own history

If at least three earlier sessions in a project had edits, and a check ran after the last edit in at least 80% of them, Cully warns when the current session has edits with no check since. It reads at most 20 earlier journals and says nothing otherwise.

## `cully status`

```text
CULLY

  Agent          Codex
  Branch         main
  Session        27m (running)
  Context        ███████░░░ 71%
  Tests          34 ✓  2 ✕
  Verification   failed
  Loops          1 ⚠
  Uncommitted    4 files
  Memory         2 recalls · 1 save
  Risk           HIGH · the agent may be looping
```

Only measured values appear. Context shows when the agent reports a window size. Uncommitted comes from Git at run time. **Risk** is a plain rule, not a score, and it names its reason:

| Level | When |
| --- | --- |
| HIGH | A loop is active, the last check failed, or context is 90% used or more |
| MEDIUM | Edits have not been checked, context is 75% used or more, or 15 or more files are uncommitted |
| LOW | Nothing is flagged |

There is no Task or Progress row. Cully does not record prompts, so it does not know the task, and it has no honest measure of progress.

## `cully timeline`

A grouped history of a session, not terminal output.

```text
00:09  Edited ×1
00:09  Check failed ×3 (same command)

Summary: 12m, 1 edit, checks 0 passed / 3 failed, 1 loop, 0 memory saves
```

`cully timeline` shows the newest session of the current project. `--all` lists recent sessions, `--session ID` picks one, and `--cwd DIR` chooses a project.

## `cully handoff`

A structured handoff so a different agent can continue without rebuilding context. It includes only sections that have data.

```sh
cully handoff --print            # print it
cully handoff codex              # start Codex in the Cully terminal with the handoff as its first prompt
```

It lists the branch, a short diff stat and the changed file names read live from Git, verification state, open problems such as detected loops, failing checks as counts only, and what the next agent should do: call `cully_context` to recover the task and earlier decisions, verify state with the project's checks, and save progress with `cully_log`. It does not guess the task, because Cully does not record prompts.

`cully handoff AGENT` needs an interactive terminal. Otherwise it prints.

## `cully rescue`

For a stuck session. It prints evidence from recorded data and rule-based recovery steps.

```text
Evidence
  - Newest session: claude, under 1m long
  - 1 edits; 0 checks passed, 3 failed
  - Verification: failed
  - The same check failed 3 times, with no edits between the failures
  - Git branch: main
  - 1 uncommitted file(s)

Recovery steps
  1. Stop editing. Run only the failing check and read the first failure before changing anything.
  2. Run the focused check for the code you just edited.
```

Without `--no-ai`, it also asks a headless advisor run (`--agent claude|codex|cursor`, default the first one installed) for a probable cause and next steps, using the evidence and file names but never file contents, with a 90 second limit. If the advisor is unavailable it says so and still prints the evidence. Cully never prints a confidence percentage.

## Why it is built this way

- **Measured, not guessed.** A signal Cully does not have is shown as missing.
- **Derived from one record.** Every command reads the same journal, so they agree with each other.
- **Private by design.** The journal is coarse on purpose. See [privacy](/privacy).
