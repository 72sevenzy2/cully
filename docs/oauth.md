# OAuth for the MCP endpoint

OAuth is optional for Cully's MCP endpoint. `cully-mcp` starts without it by default; use `cully-mcp --oauth` or `CULLY_MCP_AUTH_MODE=oauth` to enable bearer validation. Cully validates tokens issued for its public MCP URL. The authorization server and your identity provider handle sign-in, users and grants; the Cully data API and Mem0 use separate private service credentials.

Cully uses the [MCP Auth Go client SDK](https://github.com/mcp-runtime/mcp-auth/tree/main/auth-client/go) to verify RS256 signatures, exact issuer and resource audience, expiration, subject and scopes. The official MCP Go SDK handles MCP transport and the OAuth challenge. Read tools require `tools:read`; writes require `tools:write`.

## Example: a personal Cully deployment on MCP Runtime

The Cully deployment at `https://mcp.mcpruntime.org/cully/mcp` is a personal deployment on MCP Runtime. It uses the platform's OAuth service. MCP Runtime runs MCP Auth with an identity-provider connector; the identity provider manages users and sign-in. MCP Auth issues a token for this deployment's MCP URL, which Cully checks before running a tool. Its server manifest declares the resource URL and `tools:read` and `tools:write` scopes. It contains no identity-provider configuration or secret.

1. An agent connects to `https://mcp.mcpruntime.org/cully/mcp`. Without a token, Cully returns an OAuth challenge and protected-resource metadata pointing to MCP Auth.
2. MCP Runtime's MCP Auth service sends the user to the configured identity provider, then issues a token for Cully's MCP URL.
3. Cully validates the token and its tool scope before handling the request.

MCP Runtime configures the authorization service and passes the issuer and resource URL to this Cully deployment. The platform operator configures the identity-provider connector separately. See the [MCP Runtime OAuth guide](https://docs.mcpruntime.org/mcp-oauth/) for that platform setup and [the example Cully server manifest](https://github.com/mcp-runtime/cully/blob/main/.mcp/servers.yaml) for its resource and scopes.

The [Docker Compose stack](/hosting#one-command-full-stack) starts Caddy for HTTPS in OAuth mode and replaces the fixed owner with token subjects. Set public DNS and allow ports 80/443 to reach the host before starting Caddy. Keep the data API and PostgreSQL private.

An agent needs only the MCP URL and `--oauth` when connecting. For a self-hosted endpoint, its operator supplies the issuer, exact resource URL, JWKS URL and HTTPS routing. The full-stack `./setup.sh` script starts Cully, Mem0 and both databases after those values are configured.

## Self-hosted Docker with an existing authorization server

Use `--oauth existing` when an MCP-compatible authorization server already issues resource-bound tokens for your MCP URL. This can be MCP Auth deployed elsewhere. An identity provider used only for sign-in still needs an authorization server such as MCP Auth in front of Cully.

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

## Self-hosted Docker with MCP Auth

[MCP Auth](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md) can run beside Cully in the same Docker Compose stack. Configure its connector to use your organization's identity provider. The provided JSON is a Keycloak example; use the [MCP Auth connector guide](https://github.com/mcp-runtime/mcp-auth/blob/main/docs/auth-server.md) for your provider's endpoints, claims and callback settings.

1. Register an MCP Auth client with your identity provider. Set its redirect URI to your MCP Auth callback, for example `https://auth.example.com/mcp-auth/identity/callback`.
2. Create `deploy/self-hosted/connectors.json` with a named connector for that provider. For Keycloak, copy `connectors.keycloak.example.json` to `connectors.json` and replace the example realm, client and claim values. The connector name must match `CULLY_MCP_AUTH_CONNECTOR` in `.env` (default `keycloak`). Put the upstream client secret in `MCP_AUTH_UPSTREAM_CLIENT_SECRET` in the private `.env`, not in JSON.
3. In `.env`, set the issuer, resource and JWKS URLs shown above, plus `CULLY_AUTH_HOST=auth.example.com`. Generate a persistent RSA signing key at `deploy/self-hosted/.secrets/signing-key.pem`; for example, run `mkdir -p .secrets` and `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out .secrets/signing-key.pem` from that directory. Make the file readable by the MCP Auth container. Back it up along with the Compose `auth-state` volume. Keep `.env`, connector settings and key out of Git.
4. Start the stack:

```sh
cd deploy/self-hosted
./setup.sh --oauth mcp-auth
curl -fsS https://auth.example.com/mcp-auth/.well-known/jwks.json
curl -fsS https://mcp.example.com/.well-known/oauth-protected-resource/mcp
```

This stack pins `princekrroshan01/mcp-auth-server:0.4.4` and assigns Cully's read and write scopes to its resource URL. Keep the signing key and SQLite state across upgrades so existing tokens, registrations and sessions continue working. Your organization can also run MCP Auth separately and use the `--oauth existing` setup above.

With the CLI installed, append `--mcp-url https://mcp.example.com/mcp codex` for agent setup in the same command. `setup.sh` checks the OAuth values, connector JSON and signing-key file before starting this mode.

## Connect an agent

Register the public URL with `cully mcp add --agent codex --url https://mcp.example.com/mcp --oauth`, then follow that agent's sign-in instructions. The `--oauth` flag on `cully mcp add` changes setup instructions; the **server** must also enable OAuth. See [connect an agent](/agents).
