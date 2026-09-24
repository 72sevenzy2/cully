"""Authenticated HTTP client for Buddy's VM-local PostgreSQL data API."""

from __future__ import annotations

from datetime import datetime
from typing import Any
from uuid import UUID

import httpx


class BuddyRepository:
    """MCP-side repository; SQL credentials stay on the Buddy VM."""

    def __init__(self, base_url: str, token: str, timeout: float = 15.0) -> None:
        if not base_url or not token:
            raise RuntimeError("BUDDY_DATA_API_URL and BUDDY_DATA_API_TOKEN are required")
        self._client = httpx.Client(
            base_url=base_url.rstrip("/") + "/",
            headers={"Authorization": f"Bearer {token}"},
            timeout=timeout,
        )

    def initialize(self) -> None:
        """The database schema is initialized by the data API, never by MCP."""

    def close(self) -> None:
        self._client.close()

    def _request(self, method: str, path: str, **kwargs: Any) -> Any:
        path = path.lstrip("/")
        if "json" in kwargs:
            kwargs["json"] = _json_safe(kwargs["json"])
        response = self._client.request(method, path, **kwargs)
        if response.status_code >= 400:
            # Do not include response bodies, SQL errors, or credentials in MCP errors.
            raise RuntimeError(f"Buddy data API request failed ({response.status_code})")
        return response.json()

    def create_entry(self, owner_subject: str, values: dict[str, Any]) -> dict[str, Any]:
        return self._request("POST", "v1/entries", json={"owner_subject": owner_subject, "values": values})

    def search_entries(self, owner_subject: str, **params: Any) -> list[dict[str, Any]]:
        return self._request("POST", "v1/entries/search", json={"owner_subject": owner_subject, **params})

    def recent_entries(self, owner_subject: str, *, project: str | None, entry_type: str | None, limit: int) -> list[dict[str, Any]]:
        return self._request("POST", "v1/entries/recent", json={"owner_subject": owner_subject, "project": project, "entry_type": entry_type, "limit": limit})

    def get_entry(self, owner_subject: str, identifier: UUID) -> dict[str, Any] | None:
        return self._request("POST", "v1/entries/get", json={"owner_subject": owner_subject, "identifier": str(identifier)})

    def update_entry(self, owner_subject: str, identifier: UUID, fields: dict[str, Any]) -> dict[str, Any] | None:
        return self._request("PATCH", f"v1/entries/{identifier}", json={"owner_subject": owner_subject, "fields": fields})

    def delete_entry(self, owner_subject: str, identifier: UUID) -> bool:
        result = self._request("DELETE", f"v1/entries/{identifier}", json={"owner_subject": owner_subject})
        return bool(result["deleted"])

    def project_summaries(self, owner_subject: str, limit: int) -> list[dict[str, Any]]:
        return self._request("POST", "v1/projects", json={"owner_subject": owner_subject, "limit": limit})


def _json_safe(value: Any) -> Any:
    if isinstance(value, (datetime, UUID)):
        return value.isoformat() if isinstance(value, datetime) else str(value)
    if isinstance(value, dict):
        return {key: _json_safe(item) for key, item in value.items()}
    if isinstance(value, (list, tuple)):
        return [_json_safe(item) for item in value]
    return value
