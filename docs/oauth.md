# OAuth deployment

OAuth is optional. `cully-mcp` starts without it by default; use `cully-mcp --oauth` or `CULLY_MCP_AUTH_MODE=oauth` to enable bearer validation. Cully accepts a compatible authorization server that issues RS256 JWTs with an exact issuer, MCP resource audience, subject, expiration and `tools:read` or `tools:write` scopes. Your authorization server and identity provider determine login, users and grants.

The Compose examples reuse the [single-user stack](/hosting#fresh-single-user-compose-installation). They add Caddy for HTTPS and replace the fixed owner with token subjects. Set public DNS and allow ports 80/443 to reach the host before starting Caddy. Keep the data API and PostgreSQL private.

## Existing authorization server

Set these values in `deploy/self-hosted/.env`:

```dotenv
CULLY_MCP_HOST=mcp.example.com
CULLY_AUTH_ISSUER=https://auth.example.com/mcp-auth
CULLY_AUTH_RESOURCE=https://mcp.example.com/mcp
CULLY_JWKS_URL=https://auth.example.com/.well-known/jwks.json
CULLY_CADDYFILE=Caddyfile.existing-auth
```

The issuer URL is only an example. `CULLY_AUTH_RESOURCE` must equal the public MCP URL and the token audience. Use the actual signing-key endpoint of your server for `CULLY_JWKS_URL`. Register the resource and grant `tools:read` and `tools:write` with that server. Configure its OAuth discovery and clients according to its own documentation. Then run:

```sh
cd deploy/self-hosted
docker compose up -d db
docker compose -f compose.yaml -f compose.oauth.yaml --profile ops run --rm migrate
docker compose -f compose.yaml -f compose.oauth.yaml up -d --build data-api mcp caddy
curl -fsS https://mcp.example.com/.well-known/oauth-protected-resource/mcp
```

`Caddyfile.existing-auth` routes only the MCP host. The existing authorization server and its TLS route remain independently operated. The base Compose file also binds port 8080 on loopback for local diagnostics; public traffic enters through Caddy.

## New MCP Auth with a Keycloak connector

[MCP Auth](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md) can broker authorization for Cully. It is an authorization server in front of an upstream identity provider; it does not replace your identity management. Keycloak is one connector example. You can select another provider using MCP Auth's connector configuration, or choose a different authorization server.

1. Create a confidential client in your existing Keycloak realm. Register `https://auth.example.com/mcp-auth/identity/callback` as its redirect URI. Give it the identity claims you intend to use and keep its client secret private.
2. Copy `deploy/self-hosted/connectors.keycloak.example.json` to `deploy/self-hosted/connectors.keycloak.json`. Replace the realm issuer and endpoint URLs, client ID, callback URI and claim mapping. `identity_claims` must resolve to a stable, unique identity; the example uses `sub`. Put the secret in `KEYCLOAK_CLIENT_SECRET` in the private `.env`, not in JSON.
3. In `.env`, set the issuer, resource and JWKS URLs shown above, plus `CULLY_AUTH_HOST=auth.example.com` and `CULLY_CADDYFILE=Caddyfile.new-auth`. Generate a persistent RSA signing key at `deploy/self-hosted/.secrets/signing-key.pem`; for example, run `mkdir -p .secrets` and `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out .secrets/signing-key.pem` from that directory. Make the file readable by the MCP Auth container. Back it up along with the Compose `auth-state` volume. Keep `.env`, connector settings and key out of Git.
4. Start the stack:

```sh
cd deploy/self-hosted
docker compose up -d db
docker compose -f compose.yaml -f compose.oauth.yaml -f compose.mcp-auth.yaml --profile ops run --rm migrate
docker compose -f compose.yaml -f compose.oauth.yaml -f compose.mcp-auth.yaml up -d --build data-api mcp mcp-auth caddy
curl -fsS https://auth.example.com/.well-known/jwks.json
curl -fsS https://mcp.example.com/.well-known/oauth-protected-resource/mcp
```

The example pins `princekrroshan01/mcp-auth-server:0.3.0`, persists its SQLite state and signing key, enables dynamic client registration for Claude Code and Cursor, and enables trusted proxy TLS because Caddy is its only ingress. If you change the proxy or provider, adjust these settings. Preserve the signing key and state across upgrades: existing tokens, registrations and sessions depend on them.

## Connect an agent

Register the public URL with `cully mcp add --agent codex --url https://mcp.example.com/mcp --oauth`, then follow that agent's sign-in instructions. The `--oauth` flag on `cully mcp add` changes setup instructions; the **server** must also enable OAuth. See [connect an agent](/agents).

## MCP Runtime metadata

The repository's `.mcp/servers.yaml` is the personal **public OAuth deployment**, with `auth.mode: oauth`, `CULLY_MCP_AUTH_MODE=oauth`, an explicit JWKS URL and no fixed owner. Its ingress, audience and issuer must match. For a private single-user deployment, use separate metadata with `auth.mode: none`, `CULLY_MCP_AUTH_MODE=none`, a fixed owner from a secret and private ingress or network policy. Relevant server-entry fields:

```yaml
ingressHost: private-mcp.example.internal
envVars:
  - name: CULLY_MCP_AUTH_MODE
    value: none
  - name: CULLY_DATA_API_URL
    value: http://data-api:8083
secretEnvVars:
  - name: CULLY_MCP_OWNER
    secretKeyRef:
      name: cully-single-owner
      key: CULLY_MCP_OWNER
auth:
  mode: none
gateway:
  enabled: false
```

Do not point a public gateway at this no-OAuth entry. Runtime `auth.mode` and the Cully process mode must match. In either mode, supply the private API token as a secret and keep PostgreSQL and Mem0 off public ingress.
