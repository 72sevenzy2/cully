"""FastMCP application and tool-layer input validation."""

from __future__ import annotations

import re
from uuid import UUID, uuid4

from mcp.server.auth.settings import AuthSettings
from mcp.server.fastmcp import FastMCP

from buddy_mcp.auth import BuddyTokenVerifier, current_owner
from buddy_mcp.repository import BuddyRepository
from buddy_mcp.settings import BuddySettings
from buddy_mcp.time_utils import now_ist, parse_timestamp
from buddy_mcp.validation import (
    MAX_TAGS,
    checked_embedding,
    checked_text,
    normalize_project_url,
)


ENTRY_TYPES = {"work", "issue", "learning", "decision"}
SECTIONS = {"personal", "company"}
PERSONAL_CATEGORIES = {"career", "fitness", "relationship", "finance", "food", "water", "reading", "mood", "check-in", "other"}
ASSISTANT_PATTERN = re.compile(r"(?i)(codex|claude|cursor|chatgpt|other)(?:[- ][a-z0-9_.-]{1,40})?")


def create_server(
    settings: BuddySettings | None = None,
    repository: BuddyRepository | None = None,
) -> tuple[FastMCP, BuddyRepository]:
    """Build the MCP application and bind its persistence dependency."""
    settings = settings or BuddySettings.from_env()
    repository = repository or BuddyRepository(settings.data_api_url, settings.data_api_token)
    app = FastMCP(
        "Buddy",
        instructions=(
            "Shared Personal and Company memory. Every record is sectioned as personal or company. "
            "Use categories for personal notes and a GitHub project URL when company work belongs "
            "to a repository. Search within the requested section before answering memory questions."
        ),
        token_verifier=BuddyTokenVerifier(settings.issuer, settings.resource),
        auth=AuthSettings(
            issuer_url=settings.issuer,
            resource_server_url=settings.resource,
            required_scopes=["tools:read"],
            validate_token_resource=True,
        ),
        host=settings.host,
        port=settings.port,
        streamable_http_path="/mcp",
    )

    @app.tool()
    def buddy_log(
        summary: str,
        assistant: str,
        section: str,
        project_url: str | None = None,
        category: str | None = None,
        approach: str | None = None,
        entry_type: str = "work",
        outcome: str | None = None,
        issue: str | None = None,
        learning: str | None = None,
        next_steps: str | None = None,
        tags: list[str] | None = None,
        embedding: list[float] | None = None,
        occurred_at: str | None = None,
    ) -> dict:
        """Save a Personal or Company memory entry. Choose a section; project_url is optional for personal notes."""
        if entry_type not in ENTRY_TYPES:
            raise ValueError("entry_type must be work, issue, learning, or decision")
        if section not in SECTIONS:
            raise ValueError("section must be personal or company")
        if category and category not in PERSONAL_CATEGORIES:
            raise ValueError("category must be one of: " + ", ".join(sorted(PERSONAL_CATEGORIES)))
        if not ASSISTANT_PATTERN.fullmatch(assistant.strip()):
            raise ValueError("assistant must identify Codex, Claude, Cursor, ChatGPT, or other")
        tag_list = [str(tag).strip().lower()[:64] for tag in (tags or []) if str(tag).strip()]
        if len(tag_list) > MAX_TAGS:
            raise ValueError(f"at most {MAX_TAGS} tags are allowed")
        when = parse_timestamp(occurred_at, field="occurred_at") if occurred_at else now_ist()
        values = {
            "id": uuid4(),
            "project_url": normalize_project_url(project_url) if project_url else None,
            "section": section,
            "category": category,
            "entry_type": entry_type,
            "summary": checked_text("summary", summary, required=True),
            "approach": checked_text("approach", approach),
            "outcome": checked_text("outcome", outcome),
            "issue": checked_text("issue", issue),
            "learning": checked_text("learning", learning),
            "next_steps": checked_text("next_steps", next_steps),
            "assistant": assistant.strip().lower(),
            "tags": tag_list,
            "embedding": checked_embedding(embedding),
            "occurred_at": when,
        }
        return repository.create_entry(current_owner(), values)

    @app.tool()
    def buddy_search(
        query: str = "",
        project_url: str | None = None,
        entry_type: str | None = None,
        section: str | None = None,
        category: str | None = None,
        since: str | None = None,
        limit: int = 20,
        query_embedding: list[float] | None = None,
    ) -> list[dict]:
        """Search by words, section, or category and optionally by client-generated vector embeddings."""
        query = checked_text("query", query) or ""
        embedding = checked_embedding(query_embedding)
        if not query and not embedding:
            raise ValueError("provide query text or a query_embedding")
        if entry_type and entry_type not in ENTRY_TYPES:
            raise ValueError("invalid entry_type")
        if section and section not in SECTIONS:
            raise ValueError("section must be personal or company")
        if category and category not in PERSONAL_CATEGORIES:
            raise ValueError("invalid category")
        since_at = parse_timestamp(since, field="since") if since else None
        return repository.search_entries(
            current_owner(),
            query=query,
            embedding=embedding,
            project=normalize_project_url(project_url) if project_url else None,
            entry_type=entry_type,
            section=section,
            category=category,
            since=since_at,
            limit=max(1, min(int(limit), 50)),
        )

    @app.tool()
    def buddy_recent(
        project_url: str | None = None,
        entry_type: str | None = None,
        section: str | None = None,
        category: str | None = None,
        limit: int = 20,
    ) -> list[dict]:
        """Show recent entries, optionally scoped to a Personal or Company section and category."""
        if entry_type and entry_type not in ENTRY_TYPES:
            raise ValueError("invalid entry_type")
        if section and section not in SECTIONS:
            raise ValueError("section must be personal or company")
        if category and category not in PERSONAL_CATEGORIES:
            raise ValueError("invalid category")
        return repository.recent_entries(
            current_owner(),
            project=normalize_project_url(project_url) if project_url else None,
            entry_type=entry_type,
            section=section,
            category=category,
            limit=max(1, min(int(limit), 50)),
        )

    @app.tool()
    def buddy_get(entry_id: str) -> dict | None:
        """Fetch one entry owned by the authenticated user."""
        return repository.get_entry(current_owner(), parse_entry_id(entry_id))

    @app.tool()
    def buddy_update(
        entry_id: str,
        summary: str | None = None,
        approach: str | None = None,
        outcome: str | None = None,
        issue: str | None = None,
        learning: str | None = None,
        next_steps: str | None = None,
        section: str | None = None,
        category: str | None = None,
        tags: list[str] | None = None,
        embedding: list[float] | None = None,
    ) -> dict | None:
        """Update selected fields on one entry owned by the authenticated user."""
        fields = {}
        for name, value in {
            "summary": summary,
            "approach": approach,
            "outcome": outcome,
            "issue": issue,
            "learning": learning,
            "next_steps": next_steps,
            "section": section,
            "category": category,
        }.items():
            if value is not None:
                if name == "section" and value not in SECTIONS:
                    raise ValueError("section must be personal or company")
                if name == "category" and value not in PERSONAL_CATEGORIES:
                    raise ValueError("invalid category")
                fields[name] = checked_text(name, value, required=name in {"summary", "approach"})
        if tags is not None:
            if len(tags) > MAX_TAGS:
                raise ValueError(f"at most {MAX_TAGS} tags are allowed")
            fields["tags"] = [str(tag).strip().lower()[:64] for tag in tags]
        if embedding is not None:
            fields["embedding"] = checked_embedding(embedding)
        if not fields:
            raise ValueError("provide at least one field to update")
        fields["updated_at"] = now_ist()
        return repository.update_entry(current_owner(), parse_entry_id(entry_id), fields)

    @app.tool()
    def buddy_delete(entry_id: str) -> dict[str, str | bool]:
        """Delete one exact entry owned by the authenticated user."""
        identifier = parse_entry_id(entry_id)
        deleted = repository.delete_entry(current_owner(), identifier)
        return {"deleted": deleted, "id": str(identifier)}

    @app.tool()
    def buddy_projects(limit: int = 50, section: str | None = None) -> list[dict]:
        """List the authenticated user's projects with recent Buddy activity."""
        if section and section not in SECTIONS:
            raise ValueError("section must be personal or company")
        return repository.project_summaries(current_owner(), max(1, min(int(limit), 100)), section)

    return app, repository


def parse_entry_id(value: str) -> UUID:
    try:
        return UUID(value)
    except ValueError as exc:
        raise ValueError("entry_id must be a UUID") from exc


def main() -> None:
    settings = BuddySettings.from_env()
    app, repository = create_server(settings)
    repository.initialize()
    try:
        app.run(transport="streamable-http")
    finally:
        repository.close()


if __name__ == "__main__":
    main()
