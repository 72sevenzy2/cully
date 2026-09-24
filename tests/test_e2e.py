"""End-to-end test: MCP client -> Buddy MCP server -> data API -> PostgreSQL.

Every component is real except the OAuth identity provider, which is a local
issuer that publishes a JWKS and signs RS256 access tokens. Requires a
PostgreSQL database with pgvector; set BUDDY_E2E_DATABASE_URL to run it, e.g.

    docker run -d --rm -p 5432:5432 -e POSTGRES_DB=buddy -e POSTGRES_USER=buddy \\
      -e POSTGRES_PASSWORD=buddy pgvector/pgvector:pg17
    BUDDY_E2E_DATABASE_URL=postgresql://buddy:buddy@127.0.0.1:5432/buddy pytest tests/test_e2e.py
"""

from __future__ import annotations

import asyncio
import json
import os
import socket
import threading
import time
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from uuid import uuid4

DATABASE_URL = os.environ.get("BUDDY_E2E_DATABASE_URL", "")
SERVICE_TOKEN = "e2e-service-token"
MCP_PATH = "/buddy/mcp"

# data_api builds its module-level app at import time from the environment.
if DATABASE_URL:
    os.environ["BUDDY_DATABASE_URL"] = DATABASE_URL
    os.environ["BUDDY_DATA_API_TOKEN"] = SERVICE_TOKEN
os.environ["MCP_PATH"] = MCP_PATH


def free_port() -> int:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


class LocalIssuer:
    """Minimal OAuth issuer: serves a JWKS and mints RS256 access tokens."""

    def __init__(self) -> None:
        import jwt
        from cryptography.hazmat.primitives.asymmetric import rsa

        self._jwt = jwt
        self._key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
        jwk = json.loads(jwt.algorithms.RSAAlgorithm.to_jwk(self._key.public_key()))
        jwk.update({"kid": "e2e", "alg": "RS256", "use": "sig"})
        body = json.dumps({"keys": [jwk]}).encode()

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                if self.path.endswith("/.well-known/jwks.json"):
                    self.send_response(200)
                    self.send_header("content-type", "application/json")
                    self.end_headers()
                    self.wfile.write(body)
                else:
                    self.send_error(404)

            def log_message(self, *_args):
                pass

        self._server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.url = f"http://127.0.0.1:{self._server.server_address[1]}/mcp-auth"
        threading.Thread(target=self._server.serve_forever, daemon=True).start()

    def token(self, subject: str, audience: str, scope: str = "tools:read", expires_in: int = 300) -> str:
        now = int(time.time())
        claims = {"iss": self.url, "sub": subject, "aud": audience, "scope": scope, "iat": now, "exp": now + expires_in}
        return self._jwt.encode(claims, self._key, algorithm="RS256", headers={"kid": "e2e"})

    def close(self) -> None:
        self._server.shutdown()


class UvicornThread:
    """Run an ASGI app with uvicorn in a background thread."""

    def __init__(self, app, port: int) -> None:
        import uvicorn

        self.server = uvicorn.Server(uvicorn.Config(app, host="127.0.0.1", port=port, log_level="warning"))
        self.thread = threading.Thread(target=self.server.run, daemon=True)

    def start(self) -> None:
        self.thread.start()
        deadline = time.time() + 20
        while not self.server.started:
            if time.time() > deadline or not self.thread.is_alive():
                raise RuntimeError("server did not start")
            time.sleep(0.05)

    def stop(self) -> None:
        self.server.should_exit = True
        self.thread.join(timeout=10)


@unittest.skipUnless(DATABASE_URL, "set BUDDY_E2E_DATABASE_URL to run the end-to-end test")
class BuddyEndToEndTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        from buddy_mcp.data_api import create_app
        from buddy_mcp.server import create_server
        from buddy_mcp.settings import BuddySettings

        cls.issuer = LocalIssuer()
        data_port, mcp_port = free_port(), free_port()
        cls.data_api = UvicornThread(create_app(service_token=SERVICE_TOKEN), data_port)
        cls.data_api.start()

        cls.mcp_url = f"http://127.0.0.1:{mcp_port}{MCP_PATH}"
        settings = BuddySettings(
            data_api_url=f"http://127.0.0.1:{data_port}",
            data_api_token=SERVICE_TOKEN,
            issuer=cls.issuer.url,
            resource=cls.mcp_url,
            host="127.0.0.1",
            port=mcp_port,
        )
        app, cls.repository = create_server(settings)
        cls.mcp = UvicornThread(app.streamable_http_app(), mcp_port)
        cls.mcp.start()
        # Each run uses fresh subjects so reruns against the same database
        # never see earlier entries.
        cls.alice = f"alice-{uuid4()}"
        cls.bob = f"bob-{uuid4()}"

    @classmethod
    def tearDownClass(cls):
        cls.mcp.stop()
        cls.repository.close()
        cls.data_api.stop()
        cls.issuer.close()

    def call(self, subject: str, steps):
        """Open an authenticated MCP session and run steps(session)."""
        from mcp import ClientSession
        from mcp.client.streamable_http import streamable_http_client
        from mcp.shared._httpx_utils import create_mcp_http_client

        token = self.issuer.token(subject, audience=self.mcp_url)

        async def run():
            async with (
                create_mcp_http_client(headers={"Authorization": f"Bearer {token}"}) as http_client,
                streamable_http_client(self.mcp_url, http_client=http_client) as (read, write, _),
                ClientSession(read, write) as session,
            ):
                await session.initialize()
                return await steps(session)

        return asyncio.run(run())

    @staticmethod
    async def tool(session, name: str, arguments: dict):
        result = await session.call_tool(name, arguments)
        if result.isError:
            raise AssertionError(f"{name} failed: {result.content}")
        if result.structuredContent is not None:
            value = result.structuredContent
            return value.get("result", value) if isinstance(value, dict) and set(value) == {"result"} else value
        return json.loads(result.content[0].text) if result.content else None

    def test_protected_resource_metadata_matches_endpoint(self):
        import httpx

        # Clients reject a server whose advertised resource differs from the URL
        # they connect to, so this is what broke sign-in when Buddy moved hosts.
        metadata_url = self.mcp_url.replace(MCP_PATH, "/.well-known/oauth-protected-resource" + MCP_PATH)
        metadata = httpx.get(metadata_url).json()
        self.assertEqual(metadata["resource"], self.mcp_url)
        self.assertEqual(metadata["authorization_servers"], [self.issuer.url])

        response = httpx.post(self.mcp_url, json={"jsonrpc": "2.0", "id": 1, "method": "tools/list"})
        self.assertEqual(response.status_code, 401)
        self.assertIn("resource_metadata=", response.headers.get("www-authenticate", ""))

    def test_rejects_token_for_another_resource(self):
        import httpx

        token = self.issuer.token(self.alice, audience="https://example.com/other/mcp")
        response = httpx.post(
            self.mcp_url,
            headers={"Authorization": f"Bearer {token}", "accept": "application/json, text/event-stream"},
            json={"jsonrpc": "2.0", "id": 1, "method": "tools/list"},
        )
        self.assertEqual(response.status_code, 401)

    def test_memory_lifecycle_and_owner_isolation(self):
        project = "https://github.com/mcp-runtime/buddy"
        marker = f"e2e{uuid4().hex[:12]}"

        async def alice_steps(session):
            tools = {tool.name for tool in (await session.list_tools()).tools}
            self.assertTrue({"buddy_log", "buddy_search", "buddy_get", "buddy_update", "buddy_delete", "buddy_recent", "buddy_projects"} <= tools)

            created = await self.tool(session, "buddy_log", {
                "summary": f"Moved the Buddy resource to mcp.mcpruntime.org {marker}",
                "assistant": "claude",
                "section": "company",
                "project_url": "git@github.com:MCP-Runtime/Buddy.git",
                "entry_type": "work",
                "tags": ["oauth", "e2e"],
                "occurred_at": "2026-09-24T15:00:00+05:30",
            })
            entry_id = created["id"]
            self.assertEqual(created["project_url"], project)
            self.assertEqual(created["occurred_at"], "2026-09-24T15:00:00+05:30")

            found = await self.tool(session, "buddy_search", {"query": marker, "section": "company"})
            self.assertEqual([item["id"] for item in found], [entry_id])

            updated = await self.tool(session, "buddy_update", {"entry_id": entry_id, "outcome": "Clients can sign in again"})
            self.assertEqual(updated["outcome"], "Clients can sign in again")

            fetched = await self.tool(session, "buddy_get", {"entry_id": entry_id})
            self.assertEqual(fetched["outcome"], "Clients can sign in again")

            recent = await self.tool(session, "buddy_recent", {"project_url": project})
            self.assertIn(entry_id, [item["id"] for item in recent])

            projects = await self.tool(session, "buddy_projects", {"section": "company"})
            self.assertIn(project, [item["project_url"] for item in projects])
            return entry_id

        entry_id = self.call(self.alice, alice_steps)

        async def bob_steps(session):
            # Another subject sees nothing of Alice's, and cannot delete it.
            self.assertEqual(await self.tool(session, "buddy_search", {"query": marker}), [])
            self.assertIsNone(await self.tool(session, "buddy_get", {"entry_id": entry_id}))
            deleted = await self.tool(session, "buddy_delete", {"entry_id": entry_id})
            self.assertFalse(deleted["deleted"])

        self.call(self.bob, bob_steps)

        async def alice_delete(session):
            deleted = await self.tool(session, "buddy_delete", {"entry_id": entry_id})
            self.assertTrue(deleted["deleted"])
            self.assertIsNone(await self.tool(session, "buddy_get", {"entry_id": entry_id}))

        self.call(self.alice, alice_delete)

    def test_rejects_credentials_before_storing(self):
        async def steps(session):
            result = await session.call_tool("buddy_log", {
                "summary": "Rotated password=supersecretvalue for the data API",
                "assistant": "claude",
                "section": "company",
            })
            self.assertTrue(result.isError)
            found = await self.tool(session, "buddy_search", {"query": "supersecretvalue"})
            self.assertEqual(found, [])

        self.call(self.alice, steps)


if __name__ == "__main__":
    unittest.main()
