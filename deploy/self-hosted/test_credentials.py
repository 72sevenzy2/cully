import importlib.util
import json
import os
import stat
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch


SPEC = importlib.util.spec_from_file_location("credentials", Path(__file__).with_name("credentials.py"))
credentials = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(credentials)


class CredentialTests(unittest.TestCase):
    def test_generated_credentials_are_private_and_stable(self):
        with tempfile.TemporaryDirectory() as directory:
            config = Path(directory) / ".cully" / "config.json"
            env = Path(directory) / ".env"
            first = credentials.ensure(config, env)
            second = credentials.ensure(config, env)
            self.assertEqual(first, second)
            self.assertEqual(stat.S_IMODE(config.parent.stat().st_mode), 0o700)
            self.assertEqual(stat.S_IMODE(config.stat().st_mode), 0o600)
            self.assertNotEqual(first["data_api_token"], first["mem0_api_key"])
            self.assertNotEqual(first["db_password"], first["mem0_db_password"])
            self.assertEqual(json.loads(config.read_text())["self_hosted"], first)

    def test_existing_credentials_are_imported(self):
        with tempfile.TemporaryDirectory() as directory:
            config = Path(directory) / ".cully" / "config.json"
            env = Path(directory) / ".env"
            env.write_text(
                "CULLY_DB_PASSWORD=existing-db\n"
                "CULLY_DATA_API_TOKEN=existing-data\n"
                "CULLY_MEM0_DB_PASSWORD=existing-mem0-db\n"
                "CULLY_MEM0_API_KEY=existing-mem0-key\n"
                "CULLY_MEM0_JWT_SECRET=existing-jwt\n"
                "CULLY_MCP_OWNER=existing-owner\n"
            )
            current = credentials.ensure(config, env)
            self.assertEqual(current["db_password"], "existing-db")
            self.assertEqual(current["data_api_token"], "existing-data")
            self.assertEqual(current["mem0_api_key"], "existing-mem0-key")
            self.assertEqual(current["mcp_owner"], "existing-owner")

    def test_override_does_not_change_a_shared_directory(self):
        with tempfile.TemporaryDirectory() as directory:
            shared = Path(directory) / "shared"
            shared.mkdir(mode=0o755)
            with self.assertRaisesRegex(ValueError, "must be private"):
                credentials.ensure(shared / "config.json", shared / ".env")
            self.assertEqual(stat.S_IMODE(shared.stat().st_mode), 0o755)

    def test_oauth_connector_selects_organization_provider(self):
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, {}, clear=True):
            root = Path(directory)
            (root / ".env").write_text(
                "CULLY_MCP_HOST=mcp.acme.test\n"
                "CULLY_AUTH_HOST=auth.acme.test\n"
                "CULLY_AUTH_ISSUER=https://auth.acme.test/mcp-auth\n"
                "CULLY_AUTH_RESOURCE=https://mcp.acme.test/mcp\n"
                "CULLY_JWKS_URL=https://auth.acme.test/mcp-auth/.well-known/jwks.json\n"
                "CULLY_MCP_AUTH_CONNECTOR=acme\n"
                "MCP_AUTH_UPSTREAM_CLIENT_SECRET=private-value\n"
            )
            (root / "connectors.json").write_text(json.dumps({
                "acme": {"client_secret_env": "MCP_AUTH_UPSTREAM_CLIENT_SECRET"}
            }))
            credentials.validate_oauth("mcp-auth", "https://mcp.acme.test/mcp", root)
            (root / "connectors.json").write_text(json.dumps({
                "other": {"client_secret_env": "MCP_AUTH_UPSTREAM_CLIENT_SECRET"}
            }))
            with self.assertRaisesRegex(ValueError, "selected connector 'acme'"):
                credentials.validate_oauth("mcp-auth", "https://mcp.acme.test/mcp", root)

    def test_oauth_urls_and_single_connector_are_derived(self):
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, {}, clear=True):
            root = Path(directory)
            (root / ".env").write_text(
                "CULLY_MCP_HOST=mcp.acme.test\n"
                "CULLY_AUTH_HOST=auth.acme.test\n"
                "MCP_AUTH_UPSTREAM_CLIENT_SECRET=private-value\n"
            )
            (root / "connectors.json").write_text(json.dumps({
                "org": {"client_secret_env": "MCP_AUTH_UPSTREAM_CLIENT_SECRET"}
            }))
            values = credentials.validate_oauth("mcp-auth", "https://mcp.acme.test/mcp", root)
            self.assertEqual(values["CULLY_AUTH_RESOURCE"], "https://mcp.acme.test/mcp")
            self.assertEqual(values["CULLY_AUTH_ISSUER"], "https://auth.acme.test/mcp-auth")
            self.assertEqual(values["CULLY_JWKS_URL"],
                             "https://auth.acme.test/mcp-auth/.well-known/jwks.json")
            self.assertEqual(values["CULLY_MCP_AUTH_CONNECTOR"], "org")

if __name__ == "__main__":
    unittest.main()
