#!/usr/bin/env python3
"""Smoke-check the public OAuth contract used by the personal Cully release."""

import argparse
import json
import time
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

ISSUER = "https://auth.mcpruntime.org/mcp-auth"
RESOURCE = "https://mcp.mcpruntime.org/cully/mcp"
DISCOVERY = "https://auth.mcpruntime.org/.well-known/oauth-authorization-server/mcp-auth"
JWKS = "https://auth.mcpruntime.org/mcp-auth/.well-known/jwks.json"
RESOURCE_METADATA = "https://mcp.mcpruntime.org/.well-known/oauth-protected-resource/cully/mcp"
SCOPES = {"tools:read", "tools:write"}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def get_json(url):
    request = Request(url, headers={"Accept": "application/json"})
    with urlopen(request, timeout=10) as response:
        return json.load(response)


def check_auth_server():
    metadata = get_json(DISCOVERY)
    require(metadata.get("issuer") == ISSUER, "MCP Auth issuer does not match Cully metadata")
    require(metadata.get("jwks_uri") == JWKS, "MCP Auth JWKS URL does not match Cully metadata")
    for endpoint in ("authorization_endpoint", "token_endpoint", "registration_endpoint"):
        require(str(metadata.get(endpoint, "")).startswith(ISSUER + "/"),
                f"MCP Auth {endpoint} is missing or points outside the configured issuer")
    require(SCOPES.issubset(set(metadata.get("scopes_supported", []))),
            "MCP Auth must advertise tools:read and tools:write for Cully")
    keys = get_json(JWKS).get("keys", [])
    require(any(key.get("kty") == "RSA" and key.get("kid") and
                key.get("alg") in (None, "RS256") for key in keys),
            "MCP Auth has no RS256 signing key in its JWKS")
    print("MCP Auth discovery, scopes and signing keys are ready for Cully")


def check_resource_once():
    metadata = get_json(RESOURCE_METADATA)
    require(metadata.get("resource") == RESOURCE, "Cully protected resource URL is wrong")
    require(ISSUER in metadata.get("authorization_servers", []),
            "Cully protected resource does not advertise MCP Auth")
    advertised_scopes = metadata.get("scopes_supported")
    if advertised_scopes is not None:
        require(SCOPES.issubset(set(advertised_scopes)),
                "Cully protected resource advertises incomplete tool scopes")

    body = json.dumps({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {
        "protocolVersion": "2025-06-18", "capabilities": {},
        "clientInfo": {"name": "cully-release-check", "version": "1"}}}).encode()
    request = Request(RESOURCE, data=body, headers={
        "Accept": "application/json, text/event-stream",
        "Content-Type": "application/json"}, method="POST")
    try:
        with urlopen(request, timeout=10) as response:
            raise ValueError(f"Unauthenticated Cully MCP call returned HTTP {response.status}")
    except HTTPError as error:
        require(error.code == 401, f"Unauthenticated Cully MCP call returned HTTP {error.code}")
        challenge = error.headers.get("WWW-Authenticate", "")
        require("Bearer" in challenge and RESOURCE_METADATA in challenge,
                "Cully MCP did not return the expected OAuth challenge")
    print("Cully MCP route, protected resource metadata and OAuth challenge are ready")


def check_resource():
    for attempt in range(12):
        try:
            check_resource_once()
            return
        except (HTTPError, URLError, TimeoutError, ValueError) as error:
            if attempt == 11:
                raise SystemExit(f"Cully OAuth route did not become ready: {error}") from error
            time.sleep(5)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("phase", choices=("auth", "resource"))
    args = parser.parse_args()
    if args.phase == "auth":
        check_auth_server()
    else:
        check_resource()


if __name__ == "__main__":
    main()
