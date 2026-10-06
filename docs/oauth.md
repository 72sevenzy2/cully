---
title: Add sign-in to Cully
description: Set up OAuth when several people use a Cully server over HTTPS.
---

# Add sign-in to Cully

OAuth is optional for a company team deployment. You do not need it for `cully setup codex` on your laptop. When an ops team deploys one Cully server for several people over HTTPS, OAuth lets each person sign in and access their own records. [MCP Auth](https://github.com/mcp-runtime/mcp-auth) can connect Cully to the company's OIDC or OAuth 2.0 identity provider.

For the provided Docker Compose stack, prepare two public hostnames pointing to your machine: one for Cully MCP and one for MCP Auth. Ports 80 and 443 must be reachable so Caddy can serve HTTPS. You also need an identity provider where you can register a client, plus the [Cully CLI](/installation), Docker Compose and Python 3.

## Self-hosted Docker with MCP Auth

1. Create the editable files without starting containers:

   ```sh
   cully setup --prepare
   ```

2. Register a client with your identity provider. Set its redirect URI to your MCP Auth callback, such as `https://auth.example.com/mcp-auth/identity/callback`. Follow the [MCP Auth connector guide](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md#connect-an-organizations-identity-provider) for provider settings and identity claims.

3. Edit `~/.cully/self-hosted/config/.env` with your hostnames and the client secret:

   ```dotenv
   CULLY_MCP_HOST=mcp.example.com
   CULLY_AUTH_HOST=auth.example.com
   MCP_AUTH_UPSTREAM_CLIENT_SECRET=your-private-client-secret
   ```

   Use hostnames without `https://` or a path. Setup derives the public MCP URL, issuer and signing-key URL from them. Keep this file private.

4. Create `~/.cully/self-hosted/config/connectors.json` for your provider. For Keycloak, copy `connectors.keycloak.example.json` from the same directory and replace its example realm, client and claim values. Keep the client secret in `.env`, not in the JSON file. If you define more than one connector, set `CULLY_MCP_AUTH_CONNECTOR` in `.env`. See the [MCP Auth connector settings](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md#oidc-or-plain-oauth-20) for other providers.

5. Create the signing key:

   ```sh
   openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out ~/.cully/self-hosted/config/.secrets/signing-key.pem
   ```

   Make it readable by the MCP Auth container. Keep it private and back it up with the Compose `auth-state` volume.

6. Start the stack and connect your agent:

   ```sh
   cully setup codex --oauth
   ```

   Use `claude` or `cursor` instead if needed. Setup starts Cully, MCP Auth and Caddy, then registers the public MCP URL with your agent. Restart the agent. For Codex, run `codex mcp login cully`; in Claude Code, use `/mcp`; in Cursor, sign in from MCP settings.

Check the public routes after setup, replacing the example hostnames:

```sh
curl -fsS https://auth.example.com/mcp-auth/.well-known/jwks.json
curl -fsS https://mcp.example.com/.well-known/oauth-protected-resource/mcp
```

## Connect to a server that already uses OAuth

Run `cully agent setup codex --mcp-url https://mcp.example.com/mcp --oauth`, restart Codex and run `codex mcp login cully`. See [connect an agent](/agents) for Claude Code and Cursor.

Cully checks token signature, issuer, exact MCP resource URL, expiration, user and tool scopes. Read tools require `tools:read`; writes require `tools:write`. The data API and Mem0 use separate private service credentials. For another container platform or authorization server, see [team deployment](/team-deployment) and the [configuration reference](/configuration#manual-mcp-service-settings).
