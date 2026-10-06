# OAuth deployment

OAuth is optional. `cully-mcp` starts without it by default; use `cully-mcp --oauth` or `CULLY_MCP_AUTH_MODE=oauth` to enable bearer validation. Cully accepts a compatible authorization server that issues RS256 JWTs with an exact issuer, MCP resource audience, subject, expiration and `tools:read` or `tools:write` scopes. Your authorization server and identity provider determine login, users and grants.

Cully uses the [MCP Auth Go client SDK](https://github.com/Agent-Hellboy/mcp-auth/tree/main/auth-client/go) to verify OAuth tokens. The official MCP Go SDK handles MCP transport and the OAuth challenge. The MCP-to-data service token and Mem0 API key are separate from a caller's OAuth token.

## How hosted Cully uses OAuth

Cully's hosted MCP endpoint is configured to use MCP Auth deployed on MCP Runtime. MCP Auth is connected to an identity provider, currently Keycloak. The identity provider manages the realm, users and sign-in. MCP Auth handles the MCP authorization flow and issues an access token for Cully's public MCP resource. Cully validates that token before calling a tool. The data API and Mem0 use separate private service credentials; they do not use the user's OAuth token.

1. An agent connects to `https://mcp.mcpruntime.org/cully/mcp`. Without a token, Cully returns an OAuth challenge and protected-resource metadata pointing to MCP Auth.
2. MCP Auth sends the user to the configured identity provider to sign in, then issues a token whose audience is Cully's MCP URL.
3. Cully checks the token's signature, issuer, audience, expiry, subject and scopes. Reading requires `tools:read`; logging, updating and deleting require `tools:write`.

The platform operator configures the identity-provider connector for MCP Auth, including the Keycloak realm, client, callback, identity claims and supported MCP scopes. [Cully's MCP Runtime metadata](https://github.com/mcp-runtime/cully/blob/main/.mcp/servers.yaml) specifies its issuer and resource audience; it contains no Keycloak secret. MCP Runtime currently selects one connector for its MCP Auth instance rather than a separate connector in each server manifest. Tool descriptions and policy are configured per server; Cully also enforces scopes for each tool. See the [MCP Runtime OAuth guide](https://docs.mcpruntime.org/mcp-oauth/) for connector setup.

The [Docker Compose stack](/hosting#one-command-full-stack) starts Caddy for HTTPS in OAuth mode and replaces the fixed owner with token subjects. Set public DNS and allow ports 80/443 to reach the host before starting Caddy. Keep the data API and PostgreSQL private.

The **agent** only needs the MCP URL and `--oauth` at setup. The **operator** needs the issuer, exact resource URL, JWKS URL and HTTPS routing before starting the server. A new MCP Auth broker additionally needs a selected upstream connector, signing key, persistent state and upstream client secret. The full-stack `./setup.sh` script starts Mem0 and both databases in either OAuth option after these values are configured.

## Self-hosted Docker with an existing authorization server

Use `--oauth existing` when an MCP-compatible OAuth authorization server already issues resource-bound tokens for your MCP URL. This could be MCP Auth deployed elsewhere. A Keycloak realm used only for identity and login is not itself this option; use the bundled MCP Auth setup below to connect that realm to Cully.

Set these values in `deploy/self-hosted/.env`:

```dotenv
CULLY_MCP_HOST=mcp.example.com
CULLY_AUTH_ISSUER=https://auth.example.com/mcp-auth
CULLY_AUTH_RESOURCE=https://mcp.example.com/mcp
CULLY_JWKS_URL=https://auth.example.com/mcp-auth/.well-known/jwks.json
```

The issuer URL is only an example. `CULLY_AUTH_RESOURCE` must equal the public MCP URL and the token audience. Use the actual signing-key endpoint of your server for `CULLY_JWKS_URL`. Register the resource and grant `tools:read` and `tools:write` with that server. Configure its OAuth discovery and clients according to its own documentation. Then run:

```sh
cd deploy/self-hosted
./setup.sh --oauth existing
curl -fsS https://mcp.example.com/.well-known/oauth-protected-resource/mcp
```

`Caddyfile.existing-auth` routes only the MCP host. The existing authorization server and its TLS route remain independently operated. The base Compose file also binds port 8080 on loopback for local diagnostics; public traffic enters through Caddy.

`setup.sh` generates the database passwords and service tokens in `~/.cully/config.json` if they do not exist. With an installed Cully CLI, append `--mcp-url https://mcp.example.com/mcp codex` to configure the agent. Pass the exact `CULLY_AUTH_RESOURCE` value as `--mcp-url`.

## Self-hosted Docker with MCP Auth and a Keycloak connector

[MCP Auth](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md) can broker authorization for Cully. It is an authorization server in front of an upstream identity provider; it does not replace your identity management. Keycloak is one connector example. You can select another provider using MCP Auth's connector configuration, or choose a different authorization server.

1. Create a confidential client in your existing Keycloak realm. Register `https://auth.example.com/mcp-auth/identity/callback` as its redirect URI. Give it the identity claims you intend to use and keep its client secret private.
2. Copy `deploy/self-hosted/connectors.keycloak.example.json` to `deploy/self-hosted/connectors.keycloak.json`. Replace the realm issuer and endpoint URLs, client ID, callback URI and claim mapping. `identity_claims` must resolve to a stable, unique identity; the example uses `sub`. Put the secret in `KEYCLOAK_CLIENT_SECRET` in the private `.env`, not in JSON.
3. In `.env`, set the issuer, resource and JWKS URLs shown above, plus `CULLY_AUTH_HOST=auth.example.com`. Generate a persistent RSA signing key at `deploy/self-hosted/.secrets/signing-key.pem`; for example, run `mkdir -p .secrets` and `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out .secrets/signing-key.pem` from that directory. Make the file readable by the MCP Auth container. Back it up along with the Compose `auth-state` volume. Keep `.env`, connector settings and key out of Git.
4. Start the stack:

```sh
cd deploy/self-hosted
./setup.sh --oauth mcp-auth
curl -fsS https://auth.example.com/mcp-auth/.well-known/jwks.json
curl -fsS https://mcp.example.com/.well-known/oauth-protected-resource/mcp
```

The example pins `princekrroshan01/mcp-auth-server:0.3.0`, persists its SQLite state and signing key, enables dynamic client registration for Claude Code and Cursor, and enables trusted proxy TLS because Caddy is its only ingress. If you change the proxy or provider, adjust these settings. Preserve the signing key and state across upgrades: existing tokens, registrations and sessions depend on them.

With the CLI installed, append `--mcp-url https://mcp.example.com/mcp codex` for agent setup in the same command. `setup.sh` checks the OAuth values, connector JSON and signing-key file before starting this mode.

## Connect an agent

Register the public URL with `cully mcp add --agent codex --url https://mcp.example.com/mcp --oauth`, then follow that agent's sign-in instructions. The `--oauth` flag on `cully mcp add` changes setup instructions; the **server** must also enable OAuth. See [connect an agent](/agents).
