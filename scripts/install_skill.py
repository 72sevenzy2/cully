#!/usr/bin/env python3
"""Install Buddy's shared skill for Codex, Claude Code, and Cursor."""

from __future__ import annotations

import argparse
import shutil
from pathlib import Path


def install(source: Path, home: Path) -> list[Path]:
    destinations = [
        home / ".codex" / "skills" / "buddy",
        home / ".claude" / "skills" / "buddy",
        home / ".cursor" / "skills" / "buddy",
    ]
    for destination in destinations:
        destination.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source / "SKILL.md", destination / "SKILL.md")
    codex_agents = destinations[0] / "agents"
    codex_agents.mkdir(exist_ok=True)
    shutil.copy2(source / "agents" / "openai.yaml", codex_agents / "openai.yaml")
    return destinations


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--home", type=Path, default=Path.home(), help="agent configuration home")
    args = parser.parse_args()
    source = Path(__file__).resolve().parent.parent
    for path in install(source, args.home):
        print(path)


if __name__ == "__main__":
    main()
