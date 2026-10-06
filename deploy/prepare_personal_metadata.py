#!/usr/bin/env python3
"""Inject the personal data API service token into ephemeral Runtime metadata."""

import json
import os
from pathlib import Path
import sys


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit("usage: prepare_personal_metadata.py SERVERS_YAML")
    token = os.environ.get("CULLY_DATA_API_TOKEN", "")
    if not token or "\n" in token or "\r" in token:
        raise SystemExit("CULLY_DATA_API_TOKEN must be a nonempty single-line secret")
    path = Path(sys.argv[1])
    source = path.read_text()
    marker = "      namespace: mcp-servers\n"
    if source.count(marker) != 1 or "CULLY_DATA_API_TOKEN" in source:
        raise SystemExit("unexpected personal MCP metadata shape")
    item = "        - name: CULLY_DATA_API_TOKEN\n          value: " + json.dumps(token) + "\n"
    path.write_text(source.replace(marker, item + marker))


if __name__ == "__main__":
    main()
