"""Runtime configuration and shared time-zone policy."""

from __future__ import annotations

import os
from dataclasses import dataclass
from datetime import timedelta, timezone

IST = timezone(timedelta(hours=5, minutes=30), name="IST")
DEFAULT_ISSUER = "https://auth.mcpruntime.org/mcp-auth"
DEFAULT_RESOURCE = "https://mcp.mcpruntime.org/buddy/mcp"


def _first_env(*names: str, default: str) -> str:
    for name in names:
        value = os.getenv(name, "").strip()
        if value:
            return value.rstrip("/")
    return default


@dataclass(frozen=True)
class BuddySettings:
    data_api_url: str
    data_api_token: str
    issuer: str = DEFAULT_ISSUER
    resource: str = DEFAULT_RESOURCE
    host: str = "0.0.0.0"
    port: int = 8080

    @classmethod
    def from_env(cls) -> BuddySettings:
        data_api_url = os.environ.get("BUDDY_DATA_API_URL", "").strip()
        data_api_token = os.environ.get("BUDDY_DATA_API_TOKEN", "").strip()
        if not data_api_url or not data_api_token:
            raise RuntimeError("BUDDY_DATA_API_URL and BUDDY_DATA_API_TOKEN are required")
        return cls(
            data_api_url=data_api_url,
            data_api_token=data_api_token,
            issuer=_first_env("MCP_AUTH_ISSUER", default=DEFAULT_ISSUER),
            # MCP Runtime derives the OAuth resource from the server's public
            # route and injects it as MCP_AUTH_RESOURCE; BUDDY_MCP_URL is only
            # for running outside the platform.
            resource=_first_env("MCP_AUTH_RESOURCE", "BUDDY_MCP_URL", default=DEFAULT_RESOURCE),
            host=os.getenv("BUDDY_HOST", "0.0.0.0"),
            port=int(os.getenv("BUDDY_PORT", "8080")),
        )
