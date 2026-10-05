import asyncio
import time
import unittest
from unittest import mock

import jwt
from cryptography.hazmat.primitives.asymmetric import rsa

from buddy_mcp.auth import BuddyTokenVerifier

ISSUER = "https://auth.example.org/mcp-auth"
RESOURCE = "https://mcp.example.org/buddy/mcp"
PRIVATE_KEY = rsa.generate_private_key(public_exponent=65537, key_size=2048)


def make_token(**overrides):
    claims = {"iss": ISSUER, "aud": RESOURCE, "sub": "user-1", "exp": int(time.time()) + 300}
    claims.update(overrides)
    claims = {key: value for key, value in claims.items() if value is not None}
    return jwt.encode(claims, PRIVATE_KEY, algorithm="RS256")


def verify(token):
    verifier = BuddyTokenVerifier(ISSUER, RESOURCE)
    signing_key = mock.Mock(key=PRIVATE_KEY.public_key())
    with mock.patch.object(verifier.jwks, "get_signing_key_from_jwt", return_value=signing_key):
        return asyncio.run(verifier.verify_token(token))


class TokenVerifierScopeTests(unittest.TestCase):
    def test_empty_scope_defaults_to_tools_read(self):
        result = verify(make_token(scope=""))
        self.assertIsNotNone(result)
        self.assertEqual(result.scopes, ["tools:read"])
        self.assertEqual(result.subject, "user-1")

    def test_missing_scope_defaults_to_tools_read(self):
        result = verify(make_token(scope=None))
        self.assertEqual(result.scopes, ["tools:read"])

    def test_explicit_tools_read_accepted(self):
        result = verify(make_token(scope="openid tools:read"))
        self.assertEqual(result.scopes, ["openid", "tools:read"])

    def test_list_scope_accepted(self):
        result = verify(make_token(scope=["tools:read"]))
        self.assertEqual(result.scopes, ["tools:read"])

    def test_other_scopes_without_tools_read_rejected(self):
        self.assertIsNone(verify(make_token(scope="openid profile")))

    def test_wrong_audience_rejected_even_with_empty_scope(self):
        self.assertIsNone(verify(make_token(scope="", aud="https://other.example/mcp")))


if __name__ == "__main__":
    unittest.main()
