import os
import unittest
from unittest import mock

from buddy_mcp.settings import DEFAULT_ISSUER, DEFAULT_RESOURCE, BuddySettings

REQUIRED = {"BUDDY_DATA_API_URL": "https://data.example", "BUDDY_DATA_API_TOKEN": "token"}


def settings_from(env):
    with mock.patch.dict(os.environ, {**REQUIRED, **env}, clear=True):
        return BuddySettings.from_env()


class SettingsTests(unittest.TestCase):
    def test_platform_injected_resource_and_issuer_win(self):
        settings = settings_from(
            {
                "MCP_AUTH_RESOURCE": "https://mcp.example.org/buddy/mcp/",
                "MCP_AUTH_ISSUER": "https://auth.example.org/mcp-auth/",
                "BUDDY_MCP_URL": "https://stale.example.org/buddy/mcp",
            }
        )
        self.assertEqual(settings.resource, "https://mcp.example.org/buddy/mcp")
        self.assertEqual(settings.issuer, "https://auth.example.org/mcp-auth")

    def test_standalone_override_used_without_platform_resource(self):
        settings = settings_from({"BUDDY_MCP_URL": "https://local.example/buddy/mcp"})
        self.assertEqual(settings.resource, "https://local.example/buddy/mcp")

    def test_blank_values_fall_back_to_defaults(self):
        settings = settings_from({"MCP_AUTH_RESOURCE": " ", "MCP_AUTH_ISSUER": ""})
        self.assertEqual(settings.resource, DEFAULT_RESOURCE)
        self.assertEqual(settings.issuer, DEFAULT_ISSUER)


if __name__ == "__main__":
    unittest.main()
