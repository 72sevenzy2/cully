"""Private, service-authenticated HTTP boundary to Buddy's PostgreSQL store."""

from __future__ import annotations

import hmac
import os
from contextlib import asynccontextmanager
from datetime import datetime
from typing import Any
from uuid import UUID

from fastapi import Depends, FastAPI, Header, HTTPException
from pydantic import BaseModel, ConfigDict, Field

from buddy_mcp.postgres_repository import BuddyRepository


class StrictModel(BaseModel):
    model_config = ConfigDict(extra="forbid")


class CreateRequest(StrictModel):
    owner_subject: str = Field(min_length=1, max_length=512)
    values: dict[str, Any]


class SearchRequest(StrictModel):
    owner_subject: str = Field(min_length=1, max_length=512)
    query: str = ""
    embedding: str | None = None
    project: str | None = None
    entry_type: str | None = None
    section: str | None = Field(default=None, pattern="^(personal|company)$")
    category: str | None = Field(default=None, max_length=64)
    since: datetime | None = None
    limit: int = Field(ge=1, le=50)


class RecentRequest(StrictModel):
    owner_subject: str = Field(min_length=1, max_length=512)
    project: str | None = None
    entry_type: str | None = None
    section: str | None = Field(default=None, pattern="^(personal|company)$")
    category: str | None = Field(default=None, max_length=64)
    limit: int = Field(ge=1, le=50)


class GetRequest(StrictModel):
    owner_subject: str = Field(min_length=1, max_length=512)
    identifier: UUID


class UpdateRequest(StrictModel):
    owner_subject: str = Field(min_length=1, max_length=512)
    fields: dict[str, Any]


class OwnerRequest(StrictModel):
    owner_subject: str = Field(min_length=1, max_length=512)


class ProjectsRequest(StrictModel):
    owner_subject: str = Field(min_length=1, max_length=512)
    limit: int = Field(ge=1, le=100)
    section: str | None = Field(default=None, pattern="^(personal|company)$")


def create_app(repository: BuddyRepository | None = None, service_token: str | None = None) -> FastAPI:
    token = service_token or os.environ.get("BUDDY_DATA_API_TOKEN", "")
    database_url = os.environ.get("BUDDY_DATABASE_URL", "")
    if repository is None and not database_url:
        raise RuntimeError("BUDDY_DATABASE_URL is required by the data API")
    if not token:
        raise RuntimeError("BUDDY_DATA_API_TOKEN is required by the data API")

    store = repository or BuddyRepository(database_url)

    @asynccontextmanager
    async def lifespan(_: FastAPI):
        store.initialize()
        try:
            yield
        finally:
            store.close()

    app = FastAPI(title="Buddy Data API", docs_url=None, redoc_url=None, openapi_url=None, lifespan=lifespan)

    async def require_service(authorization: str | None = Header(default=None)) -> None:
        scheme, _, supplied = (authorization or "").partition(" ")
        if scheme.lower() != "bearer" or not supplied or not hmac.compare_digest(supplied, token):
            raise HTTPException(status_code=401, detail="unauthorized", headers={"WWW-Authenticate": "Bearer"})

    @app.get("/healthz")
    async def healthz() -> dict[str, str]:
        return {"status": "ok"}

    @app.post("/v1/entries", dependencies=[Depends(require_service)])
    async def create_entry(request: CreateRequest) -> dict[str, Any]:
        values = dict(request.values)
        values["id"] = UUID(str(values["id"]))
        values["occurred_at"] = datetime.fromisoformat(str(values["occurred_at"]))
        return store.create_entry(request.owner_subject, values)

    @app.post("/v1/entries/search", dependencies=[Depends(require_service)])
    async def search_entries(request: SearchRequest) -> list[dict[str, Any]]:
        return store.search_entries(
            request.owner_subject, query=request.query, embedding=request.embedding,
            project=request.project, entry_type=request.entry_type, since=request.since,
            section=request.section, category=request.category, limit=request.limit,
        )

    @app.post("/v1/entries/recent", dependencies=[Depends(require_service)])
    async def recent_entries(request: RecentRequest) -> list[dict[str, Any]]:
        return store.recent_entries(
            request.owner_subject, project=request.project,
            entry_type=request.entry_type, section=request.section,
            category=request.category, limit=request.limit,
        )

    @app.post("/v1/entries/get", dependencies=[Depends(require_service)])
    async def get_entry(request: GetRequest) -> dict[str, Any] | None:
        return store.get_entry(request.owner_subject, request.identifier)

    @app.patch("/v1/entries/{identifier}", dependencies=[Depends(require_service)])
    async def update_entry(identifier: UUID, request: UpdateRequest) -> dict[str, Any] | None:
        fields = dict(request.fields)
        if "updated_at" in fields:
            fields["updated_at"] = datetime.fromisoformat(str(fields["updated_at"]))
        return store.update_entry(request.owner_subject, identifier, fields)

    @app.delete("/v1/entries/{identifier}", dependencies=[Depends(require_service)])
    async def delete_entry(identifier: UUID, request: OwnerRequest) -> dict[str, bool]:
        return {"deleted": store.delete_entry(request.owner_subject, identifier)}

    @app.post("/v1/projects", dependencies=[Depends(require_service)])
    async def projects(request: ProjectsRequest) -> list[dict[str, Any]]:
        return store.project_summaries(request.owner_subject, request.limit, request.section)

    return app


app = create_app()
