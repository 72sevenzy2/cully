"""OAuth-protected Buddy MCP server with PostgreSQL full-text and vector search."""

from __future__ import annotations

import logging
import math
import os
import re
from datetime import datetime, timezone
from typing import Any
from urllib.parse import urlparse
from uuid import UUID, uuid4

import jwt
from mcp.server.auth.provider import AccessToken, TokenVerifier
from mcp.server.auth.settings import AuthSettings
from mcp.server.fastmcp import FastMCP
from psycopg.rows import dict_row
from psycopg_pool import ConnectionPool

logging.basicConfig(level=os.getenv("BUDDY_LOG_LEVEL", "INFO"))
log = logging.getLogger("buddy")

ISSUER = os.getenv("MCP_AUTH_ISSUER", "https://auth.mcpruntime.org/mcp-auth").rstrip("/")
RESOURCE = os.getenv("BUDDY_MCP_URL", "https://workspace.mcpruntime.org/buddy/mcp").rstrip("/")
DATABASE_URL = os.environ.get("BUDDY_DATABASE_URL", "")
if not DATABASE_URL:
    raise RuntimeError("BUDDY_DATABASE_URL is required")

MAX_TEXT = 8000
MAX_TAGS = 20
EMBEDDING_DIM = 1536
SECRET_PATTERNS = [
    re.compile(r"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----"),
    re.compile(r"\b(?:gh[pousr]_[A-Za-z0-9_]{20,}|github_pat_[A-Za-z0-9_]{20,})\b"),
    re.compile(r"(?i)\b(?:password|secret|api[_-]?key|access[_-]?token)\s*[:=]\s*[^\s,;]{8,}"),
]


def normalize_project_url(value: str) -> str:
    value = value.strip()
    match = re.fullmatch(r"git@github\.com:([^/]+)/([^/]+?)(?:\.git)?", value, re.I)
    if match:
        owner, repo = match.groups()
    else:
        parsed = urlparse(value)
        if parsed.scheme not in {"https", "http"} or parsed.hostname != "github.com":
            raise ValueError("project_url must be a GitHub repository URL")
        pieces = parsed.path.strip("/").split("/")
        if len(pieces) != 2:
            raise ValueError("project_url must point to OWNER/REPO")
        owner, repo = pieces
        repo = repo.removesuffix(".git")
    if not re.fullmatch(r"[A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?", owner):
        raise ValueError("invalid GitHub owner in project_url")
    if not re.fullmatch(r"[A-Za-z0-9_.-]{1,100}", repo):
        raise ValueError("invalid GitHub repository in project_url")
    return f"https://github.com/{owner.lower()}/{repo.lower()}"


def checked_text(field: str, value: str | None, *, required: bool = False) -> str | None:
    if value is None:
        if required:
            raise ValueError(f"{field} is required")
        return None
    value = value.strip()
    if required and not value:
        raise ValueError(f"{field} cannot be empty")
    if len(value) > MAX_TEXT:
        raise ValueError(f"{field} is longer than {MAX_TEXT} characters")
    for pattern in SECRET_PATTERNS:
        if pattern.search(value):
            raise ValueError(f"{field} appears to contain a credential; remove it before saving")
    return value or None


def checked_embedding(values: list[float] | None) -> str | None:
    if values is None:
        return None
    if len(values) != EMBEDDING_DIM:
        raise ValueError(f"embedding must contain exactly {EMBEDDING_DIM} values")
    vector = [float(value) for value in values]
    if not all(math.isfinite(value) for value in vector):
        raise ValueError("embedding values must be finite numbers")
    return "[" + ",".join(format(value, ".9g") for value in vector) + "]"


SCHEMA = """
CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE IF NOT EXISTS buddy_entries (
  id uuid PRIMARY KEY,
  project_url text NOT NULL,
  theme text NOT NULL DEFAULT 'mcp',
  entry_type text NOT NULL CHECK (entry_type IN ('work','issue','learning','decision')),
  summary text NOT NULL,
  approach text,
  outcome text,
  issue text,
  learning text,
  next_steps text,
  assistant text NOT NULL,
  tags text[] NOT NULL DEFAULT '{}',
  occurred_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  search_vector tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(summary,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(approach,'')), 'B') ||
    setweight(to_tsvector('english', coalesce(outcome,'')), 'B') ||
    setweight(to_tsvector('english', coalesce(issue,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(learning,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(next_steps,'')), 'C')
  ) STORED,
  embedding vector(1536)
);
ALTER TABLE buddy_entries ADD COLUMN IF NOT EXISTS embedding vector(1536);
CREATE INDEX IF NOT EXISTS buddy_entries_search_idx ON buddy_entries USING gin(search_vector);
CREATE INDEX IF NOT EXISTS buddy_entries_project_time_idx ON buddy_entries(project_url, occurred_at DESC);
CREATE INDEX IF NOT EXISTS buddy_entries_type_idx ON buddy_entries(entry_type, occurred_at DESC);
CREATE INDEX IF NOT EXISTS buddy_entries_embedding_idx ON buddy_entries USING hnsw (embedding vector_cosine_ops) WHERE embedding IS NOT NULL;
"""

pool = ConnectionPool(DATABASE_URL, min_size=1, max_size=8, kwargs={"row_factory": dict_row}, open=False)


def initialize() -> None:
    pool.open(wait=True)
    with pool.connection() as conn:
        conn.execute(SCHEMA)


class BuddyTokenVerifier(TokenVerifier):
    def __init__(self) -> None:
        self.jwks = jwt.PyJWKClient(f"{ISSUER}/.well-known/jwks.json", cache_jwk_set=True)

    async def verify_token(self, token: str) -> AccessToken | None:
        try:
            key = self.jwks.get_signing_key_from_jwt(token).key
            claims = jwt.decode(
                token,
                key,
                algorithms=["RS256"],
                issuer=ISSUER,
                audience=RESOURCE,
                options={"require": ["exp", "iss", "aud", "sub"]},
            )
            scopes = claims.get("scope", "")
            if isinstance(scopes, str):
                scopes = scopes.split()
            if "tools:read" not in scopes:
                return None
            return AccessToken(
                token=token,
                client_id=str(claims.get("client_id", claims.get("azp", "oauth-client"))),
                scopes=list(scopes),
                expires_at=int(claims["exp"]),
                resource=RESOURCE,
                subject=str(claims["sub"]),
                claims={"iss": ISSUER},
            )
        except (jwt.PyJWTError, ValueError, KeyError, TypeError):
            return None


mcp = FastMCP(
    "Buddy",
    instructions=(
        "Shared work memory for GitHub projects. Record concise work, approach, "
        "outcomes, issues, lessons, and next steps. Search before answering memory questions."
    ),
    token_verifier=BuddyTokenVerifier(),
    auth=AuthSettings(
        issuer_url=ISSUER,
        resource_server_url=RESOURCE,
        required_scopes=["tools:read"],
        validate_token_resource=True,
    ),
    host="0.0.0.0",
    port=8080,
    streamable_http_path="/mcp",
)


def as_record(row: dict[str, Any]) -> dict[str, Any]:
    row.pop("search_rank", None)
    row.pop("embedding", None)
    if row.get("id") is not None:
        row["id"] = str(row["id"])
    row["tags"] = list(row.get("tags") or [])
    for key in ("occurred_at", "created_at", "updated_at"):
        if row.get(key):
            row[key] = row[key].isoformat()
    return row


@mcp.tool()
def buddy_log(
    project_url: str,
    summary: str,
    approach: str,
    assistant: str,
    entry_type: str = "work",
    outcome: str | None = None,
    issue: str | None = None,
    learning: str | None = None,
    next_steps: str | None = None,
    tags: list[str] | None = None,
    embedding: list[float] | None = None,
    occurred_at: str | None = None,
) -> dict[str, Any]:
    """Save a work update, issue, learning, or decision to shared Buddy memory."""
    if entry_type not in {"work", "issue", "learning", "decision"}:
        raise ValueError("entry_type must be work, issue, learning, or decision")
    if not re.fullmatch(r"(?i)(codex|claude|cursor|chatgpt|other)(?:[- ][a-z0-9_.-]{1,40})?", assistant.strip()):
        raise ValueError("assistant must identify Codex, Claude, Cursor, ChatGPT, or other")
    tag_list = [str(tag).strip().lower()[:64] for tag in (tags or []) if str(tag).strip()]
    if len(tag_list) > MAX_TAGS:
        raise ValueError(f"at most {MAX_TAGS} tags are allowed")
    when = datetime.fromisoformat(occurred_at) if occurred_at else datetime.now(timezone.utc)
    if when.tzinfo is None:
        raise ValueError("occurred_at must include a timezone")
    values = {
        "id": uuid4(),
        "project_url": normalize_project_url(project_url),
        "entry_type": entry_type,
        "summary": checked_text("summary", summary, required=True),
        "approach": checked_text("approach", approach, required=True),
        "outcome": checked_text("outcome", outcome),
        "issue": checked_text("issue", issue),
        "learning": checked_text("learning", learning),
        "next_steps": checked_text("next_steps", next_steps),
        "assistant": assistant.strip().lower(),
        "tags": tag_list,
        "embedding": checked_embedding(embedding),
        "occurred_at": when,
    }
    with pool.connection() as conn:
        row = conn.execute(
            """INSERT INTO buddy_entries
               (id,project_url,entry_type,summary,approach,outcome,issue,learning,next_steps,assistant,tags,embedding,occurred_at)
               VALUES (%(id)s,%(project_url)s,%(entry_type)s,%(summary)s,%(approach)s,%(outcome)s,%(issue)s,%(learning)s,%(next_steps)s,%(assistant)s,%(tags)s,%(embedding)s::vector,%(occurred_at)s)
               RETURNING id,project_url,theme,entry_type,summary,approach,outcome,issue,learning,next_steps,assistant,tags,occurred_at,created_at,updated_at""",
            values,
        ).fetchone()
    return as_record(row)


@mcp.tool()
def buddy_search(
    query: str = "",
    project_url: str | None = None,
    entry_type: str | None = None,
    since: str | None = None,
    limit: int = 20,
    query_embedding: list[float] | None = None,
) -> list[dict[str, Any]]:
    """Search by words and optionally by client-generated vector embeddings."""
    query = checked_text("query", query) or ""
    embedding = checked_embedding(query_embedding)
    if not query and not embedding:
        raise ValueError("provide query text or a query_embedding")
    if entry_type and entry_type not in {"work", "issue", "learning", "decision"}:
        raise ValueError("invalid entry_type")
    limit = max(1, min(int(limit), 50))
    project = normalize_project_url(project_url) if project_url else None
    since_at = datetime.fromisoformat(since) if since else None
    if since_at and since_at.tzinfo is None:
        raise ValueError("since must include a timezone")
    with pool.connection() as conn:
        rows = conn.execute(
            """SELECT id,project_url,theme,entry_type,summary,approach,outcome,issue,learning,next_steps,assistant,tags,occurred_at,created_at,updated_at,
                      (CASE WHEN %(embedding)s::vector IS NOT NULL AND embedding IS NOT NULL
                            THEN 1 - (embedding <=> %(embedding)s::vector) ELSE 0 END
                       + CASE WHEN %(query)s <> ''
                              THEN ts_rank_cd(search_vector, websearch_to_tsquery('english', %(query)s)) ELSE 0 END) AS search_rank
               FROM buddy_entries
               WHERE ((%(query)s <> '' AND search_vector @@ websearch_to_tsquery('english', %(query)s))
                   OR (%(embedding)s::vector IS NOT NULL AND embedding IS NOT NULL))
                 AND (%(project)s::text IS NULL OR project_url=%(project)s::text)
                 AND (%(entry_type)s::text IS NULL OR entry_type=%(entry_type)s::text)
                 AND (%(since)s::timestamptz IS NULL OR occurred_at >= %(since)s::timestamptz)
               ORDER BY search_rank DESC, occurred_at DESC LIMIT %(limit)s""",
            {"query": query, "embedding": embedding, "project": project, "entry_type": entry_type, "since": since_at, "limit": limit},
        ).fetchall()
    return [as_record(row) for row in rows]


@mcp.tool()
def buddy_recent(
    project_url: str | None = None,
    entry_type: str | None = None,
    limit: int = 20,
) -> list[dict[str, Any]]:
    """Show the latest work-memory entries, optionally scoped to a project."""
    if entry_type and entry_type not in {"work", "issue", "learning", "decision"}:
        raise ValueError("invalid entry_type")
    limit = max(1, min(int(limit), 50))
    project = normalize_project_url(project_url) if project_url else None
    with pool.connection() as conn:
        rows = conn.execute(
            """SELECT * FROM buddy_entries
               WHERE (%(project)s::text IS NULL OR project_url=%(project)s::text)
                 AND (%(entry_type)s::text IS NULL OR entry_type=%(entry_type)s::text)
               ORDER BY occurred_at DESC LIMIT %(limit)s""",
            {"project": project, "entry_type": entry_type, "limit": limit},
        ).fetchall()
    return [as_record(row) for row in rows]


@mcp.tool()
def buddy_get(entry_id: str) -> dict[str, Any] | None:
    """Fetch one Buddy entry by its ID."""
    try:
        identifier = UUID(entry_id)
    except ValueError as exc:
        raise ValueError("entry_id must be a UUID") from exc
    with pool.connection() as conn:
        row = conn.execute("SELECT * FROM buddy_entries WHERE id=%s", (identifier,)).fetchone()
    return as_record(row) if row else None


@mcp.tool()
def buddy_update(
    entry_id: str,
    summary: str | None = None,
    approach: str | None = None,
    outcome: str | None = None,
    issue: str | None = None,
    learning: str | None = None,
    next_steps: str | None = None,
    tags: list[str] | None = None,
    embedding: list[float] | None = None,
) -> dict[str, Any] | None:
    """Update selected fields of an existing Buddy entry."""
    try:
        identifier = UUID(entry_id)
    except ValueError as exc:
        raise ValueError("entry_id must be a UUID") from exc
    fields: dict[str, Any] = {}
    for name, value in {"summary": summary, "approach": approach, "outcome": outcome,
                        "issue": issue, "learning": learning, "next_steps": next_steps}.items():
        if value is not None:
            fields[name] = checked_text(name, value, required=name in {"summary", "approach"})
    if tags is not None:
        if len(tags) > MAX_TAGS:
            raise ValueError(f"at most {MAX_TAGS} tags are allowed")
        fields["tags"] = [str(tag).strip().lower()[:64] for tag in tags]
    if embedding is not None:
        fields["embedding"] = checked_embedding(embedding)
    if not fields:
        raise ValueError("provide at least one field to update")
    fields["updated_at"] = datetime.now(timezone.utc)
    assignments = ",".join(
        f"{name}=%({name})s::vector" if name == "embedding" else f"{name}=%({name})s"
        for name in fields
    )
    fields["id"] = identifier
    with pool.connection() as conn:
        row = conn.execute(
            f"UPDATE buddy_entries SET {assignments} WHERE id=%(id)s RETURNING *", fields
        ).fetchone()
    return as_record(row) if row else None


@mcp.tool()
def buddy_delete(entry_id: str) -> dict[str, Any]:
    """Delete one exact Buddy entry by UUID. This cannot bulk-delete records."""
    try:
        identifier = UUID(entry_id)
    except ValueError as exc:
        raise ValueError("entry_id must be a UUID") from exc
    with pool.connection() as conn:
        row = conn.execute("DELETE FROM buddy_entries WHERE id=%s RETURNING id", (identifier,)).fetchone()
    return {"deleted": row is not None, "id": str(identifier)}


@mcp.tool()
def buddy_projects(limit: int = 50) -> list[dict[str, Any]]:
    """List project repositories with recent Buddy activity."""
    limit = max(1, min(int(limit), 100))
    with pool.connection() as conn:
        rows = conn.execute(
            """SELECT project_url, count(*) AS entry_count, max(occurred_at) AS last_activity
               FROM buddy_entries GROUP BY project_url ORDER BY last_activity DESC LIMIT %s""",
            (limit,),
        ).fetchall()
    for row in rows:
        row["last_activity"] = row["last_activity"].isoformat()
    return rows


if __name__ == "__main__":
    initialize()
    mcp.run(transport="streamable-http")
