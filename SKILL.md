---
name: buddy
description: Use when the user wants to log, search, recall, or analyze personal life data in Buddy, including career, fitness, relationships, finance, food, water, reading, mood, spending, habits, check-ins, reviews, and life optimization.
---

# Buddy

Buddy is the user's local life memory and analytics system. Use `/Users/proshan/buddy/buddy.py` and store data only in `/Users/proshan/buddy/memory/buddy_memory.db`.

## Core Behavior

- The user usually feeds raw data. Save it with minimal friction.
- Ask for data only during check-ins or when one missing detail materially improves retrieval or analytics.
- Ask at most one follow-up for concrete values such as amount, time, duration, cost, quantity, pages, reps, distance, or sleep.
- Do not turn casual logging into an interview.
- For analytics, summarize patterns and practical next actions from saved memory.

## Commands

Save a raw event:

```bash
python3 /Users/proshan/buddy/buddy.py log <category> "<text>"
```

Save food with nutrition estimates:

```bash
python3 /Users/proshan/buddy/buddy.py food --at 2026-07-20T06:00:00+05:30 --calories 660 --confidence medium --assumptions "serving sizes estimated" "<food text>"
```

Run a check-in:

```bash
python3 /Users/proshan/buddy/buddy.py ask --mode quick
python3 /Users/proshan/buddy/buddy.py ask --mode daily
python3 /Users/proshan/buddy/buddy.py ask --mode weekly
```

Search and analyze memory:

```bash
python3 /Users/proshan/buddy/buddy.py search "<query>"
python3 /Users/proshan/buddy/buddy.py summary
python3 /Users/proshan/buddy/buddy.py analytics
```

Explore with open-source analytics wrappers:

```bash
/Users/proshan/buddy/scripts/buddy-datasette.sh
/Users/proshan/buddy/scripts/buddy-dogsheep-index.sh
```

## Categories

- `career`: work, learning, projects, reading, reputation, opportunities.
- `fitness`: food, training, sleep, water, body metrics, recovery, energy.
- `relationship`: partner, family, friends, network, conflict, support.
- `finance`: spending, income, investing, subscriptions, debt, waste.
- `food`, `water`, `reading`, `mood`, `other`: quick captures.

## Analytics

When asked for optimization, review the SQLite memory through `analytics`, `summary`, and targeted `search` calls. Look for:

- Food analytics should use the structured `food` command when calories, macros, confidence, assumptions, or event timestamps are available.
- Repeated drains on time, money, health, attention, or relationships.
- Missing signals Buddy should ask for in the next check-in.
- Career leverage, wasted spend, workout consistency, sleep/energy patterns, reading consistency, and neglected relationships.
- Concrete next actions, not generic advice.
