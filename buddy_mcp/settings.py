"""Runtime configuration and shared time-zone policy."""

from __future__ import annotations

import os
from dataclasses import dataclass
from datetime import timedelta, timezone

IST = timezone(timedelta(hours=5, minutes=30), name="IST")
DEFAULT_ISSUER = "https://auth.mcpruntime.org/mcp-auth"
DEFAULT_RESOURCE = "https://mcp.mcpruntime.org/buddy/mcp"


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
            issuer=os.getenv("MCP_AUTH_ISSUER", DEFAULT_ISSUER).rstrip("/"),
            resource=os.getenv("BUDDY_MCP_URL", DEFAULT_RESOURCE).rstrip("/"),
            host=os.getenv("BUDDY_HOST", "0.0.0.0"),
            port=int(os.getenv("BUDDY_PORT", "8080")),
        )
