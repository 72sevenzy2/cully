# OAuth for the MCP endpoint

OAuth is optional for Cully's MCP endpoint. `cully-mcp` starts without it by default; use `cully-mcp --oauth` or `CULLY_MCP_AUTH_MODE=oauth` to enable bearer validation. Cully validates tokens issued for its public MCP URL. The authorization server and your identity provider handle sign-in, users and grants; the Cully data API and Mem0 use separate private service credentials.

Cully uses the [MCP Auth Go client SDK](https://github.com/mcp-runtime/mcp-auth/tree/main/auth-client/go) to verify RS256 signatures, exact issuer and resource audience, expiration, subject and scopes. The official MCP Go SDK handles MCP transport and the OAuth challenge. Read tools require `tools:read`; writes require `tools:write`.

## Example: the maintainer's deployment on MCP Runtime

[MCP Runtime](https://mcpruntime.org) is a platform where you can deploy an MCP server. The Cully maintainer uses it for a personal Cully deployment at `https://mcp.mcpruntime.org/cully/mcp`. MCP Runtime provides OAuth through MCP Auth, so this Cully deployment only declares its public resource URL and `tools:read` and `tools:write` scopes. Its manifest contains no identity-provider secret.

1. An agent connects to `https://mcp.mcpruntime.org/cully/mcp`. Without a token, Cully returns an OAuth challenge and protected-resource metadata pointing to MCP Auth.
2. MCP Runtime's MCP Auth service sends the user to the configured identity provider, then issues a token for Cully's MCP URL.
3. Cully validates the token and its tool scope before handling the request.

MCP Runtime configures the authorization service and passes the issuer and resource URL to Cully. See [publish an MCP server on MCP Runtime](https://docs.mcpruntime.org/publish-mcp-server/), the [platform OAuth guide](https://docs.mcpruntime.org/mcp-oauth/), and [this deployment's server manifest](https://github.com/mcp-runtime/cully/blob/main/.mcp/servers.yaml).

You can use the same pattern in your environment: deploy Cully MCP, connect an authorization server, and configure that server to use your organization's identity provider. [MCP Auth](https://github.com/mcp-runtime/mcp-auth) is a provider-neutral OAuth broker that supports OIDC and OAuth 2.0 connectors; it does not replace the organization's identity provider. An agent only needs the MCP URL and `--oauth` when connecting.

## Self-hosted Docker with an existing authorization server

Use `--oauth existing` when an MCP-compatible authorization server already issues resource-bound tokens for your MCP URL. This can be MCP Auth deployed elsewhere. An identity provider used only for sign-in still needs an authorization server such as MCP Auth in front of Cully.

Set these values in `deploy/self-hosted/.env`:

```dotenv
CULLY_MCP_HOST=mcp.example.com
CULLY_AUTH_ISSUER=https://auth.example.com/mcp-auth
```

`setup.sh` derives the MCP resource URL as `https://mcp.example.com/mcp` and defaults the signing-key URL to `<issuer>/.well-known/jwks.json`. Set `CULLY_JWKS_URL` in `.env` only when your authorization server uses another endpoint. Register that exact resource and grant `tools:read` and `tools:write` with your authorization server. Configure its OAuth discovery and clients according to its documentation. Then run:

```sh
cd deploy/self-hosted
./setup.sh --oauth existing codex
curl -fsS https://mcp.example.com/.well-known/oauth-protected-resource/mcp
```

`Caddyfile.existing-auth` routes only the MCP host. The existing authorization server and its TLS route remain independently operated. The base Compose file also binds port 8080 on loopback for local diagnostics; public traffic enters through Caddy.

`setup.sh` generates database passwords and service tokens in `~/.cully/config.json`, starts the Docker stack and, when the CLI is installed, configures Codex with the derived MCP URL. Use `claude` or `cursor` instead for those agents. Set public DNS and allow ports 80/443 to reach Caddy; keep the data API and PostgreSQL private.

## Self-hosted Docker with MCP Auth

[MCP Auth](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md) can run beside Cully in the same Docker Compose stack. Configure its connector to use an identity provider supported by your organization. Follow the MCP Auth documentation for [OIDC or OAuth 2.0 connector settings](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md#oidc-or-plain-oauth-20) and [the redirect URI to register with the identity provider](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md#the-redirect-uri-you-register-with-your-identity-provider). Those settings make the identity provider work with MCP Auth's MCP OAuth flow; Cully only verifies the resulting resource-bound token. The provided JSON is one Keycloak example.

1. Register an MCP Auth client with your identity provider. Set its redirect URI to your MCP Auth callback, for example `https://auth.example.com/mcp-auth/identity/callback`.
2. Create `deploy/self-hosted/connectors.json` with a named connector for that provider. For Keycloak, copy `connectors.keycloak.example.json` to `connectors.json` and replace the example realm, client and claim values. If the file has one connector, setup selects it automatically; set `CULLY_MCP_AUTH_CONNECTOR` only if it has more than one. Keep the upstream client secret in the private `.env`, not in JSON.
3. Set these values in `.env`:

   ```dotenv
   CULLY_MCP_HOST=mcp.example.com
   CULLY_AUTH_HOST=auth.example.com
   MCP_AUTH_UPSTREAM_CLIENT_SECRET=your-private-client-secret
   ```

   `setup.sh` derives the resource URL, issuer and JWKS URL from the two hostnames. Generate a persistent RSA signing key at `deploy/self-hosted/.secrets/signing-key.pem`; for example, run `mkdir -p .secrets` and `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out .secrets/signing-key.pem` from that directory. Make the file readable by the MCP Auth container. Back it up along with the Compose `auth-state` volume. Keep `.env`, connector settings and key out of Git.

4. Start the stack and configure an agent:

   ```sh
   cd deploy/self-hosted
   ./setup.sh --oauth mcp-auth codex
   curl -fsS https://auth.example.com/mcp-auth/.well-known/jwks.json
   curl -fsS https://mcp.example.com/.well-known/oauth-protected-resource/mcp
   ```

This stack pins `princekrroshan01/mcp-auth-server:0.4.4` and assigns Cully's read and write scopes to its resource URL. Keep the signing key and SQLite state across upgrades so existing tokens, registrations and sessions continue working. Your organization can also run MCP Auth separately and use the `--oauth existing` setup above.

`setup.sh` checks the OAuth values, connector JSON and signing-key file before starting this mode. The CLI must be installed to configure the agent in the same command.

## Connect an agent

For an already-running server, use `cully setup codex --mcp-url https://mcp.example.com/mcp --oauth` to install the skill and register the MCP connection together. Then follow the agent's sign-in instructions. See [connect an agent](/agents).
