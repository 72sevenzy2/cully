# OAuth deployment

OAuth is optional. `cully-mcp` starts without it by default; use `cully-mcp --oauth` or `CULLY_MCP_AUTH_MODE=oauth` to enable bearer validation. Cully accepts a compatible authorization server that issues RS256 JWTs with an exact issuer, MCP resource audience, subject, expiration and `tools:read` or `tools:write` scopes. Your authorization server and identity provider determine login, users and grants.

Cully uses the [MCP Auth Go client SDK](https://github.com/Agent-Hellboy/mcp-auth/tree/main/auth-client/go) to verify OAuth tokens. The official MCP Go SDK handles MCP transport and the OAuth challenge. The MCP-to-data service token and Mem0 API key are separate from a caller's OAuth token.

## Personal hosted deployment on MCP Runtime

This deployment uses the platform's existing MCP Auth at `https://auth.mcpruntime.org/mcp-auth`, backed by its Keycloak identity provider. Cully runs as an MCP workload on MCP Runtime; its data API, PostgreSQL and Mem0 run separately on the personal VM. The platform's MCP Auth connector authenticates the user through Keycloak and issues an MCP access token. Cully validates that token before running a tool. The private MCP-to-data API token and data-to-Mem0 API key are unrelated to user sign-in.

1. **Configure the identity provider in MCP Runtime.** The platform operator keeps Keycloak's realm issuer, authorization/token/JWKS endpoints, confidential `mcp-auth` client ID, client secret reference, stable identity claim and exact `https://auth.mcpruntime.org/mcp-auth/identity/callback` URI in the selected MCP Auth connector. Reuse the platform's working connector; add `"tools:write"` beside `"tools:read"` in its `mcp_scopes`. Its upstream `scopes` (`openid`, `profile`, and so on) serve Keycloak login and are separate. Keep the client secret in the platform's connector Secret. The historical Buddy MCP manifest contained only the platform issuer and Buddy's resource audience; it did not contain Keycloak settings.
2. **Publish Cully's resource.** [`.mcp/servers.yaml`](https://github.com/mcp-runtime/cully/blob/main/.mcp/servers.yaml) sets `auth.mode: oauth`, `issuerURL: https://auth.mcpruntime.org/mcp-auth`, `audience: https://mcp.mcpruntime.org/cully/mcp`, and `CULLY_MCP_AUTH_MODE=oauth`. Its `CULLY_JWKS_URL` points to MCP Auth's signing keys. The Runtime operator reconciles this audience into MCP Auth and injects `MCP_AUTH_ISSUER`, `MCP_AUTH_RESOURCE` and `MCP_PATH` into Cully. Keycloak configuration does not belong in this manifest.
3. **Deploy the services.** A `v*` release runs [the personal deployment workflow](https://github.com/mcp-runtime/cully/blob/main/.github/workflows/personal-deploy.yml). It deploys PostgreSQL, Mem0 and the data API on the VM, then builds, pushes and deploys the Cully MCP image through the pinned `mcp-runtime` CLI and the platform API key. The MCP workload receives its private `CULLY_DATA_API_TOKEN`; it does not receive PostgreSQL or Mem0 credentials. The [personal deployment runbook](/personal-deployment) lists the VM and Actions settings.
4. **Connect an agent.** Register `https://mcp.mcpruntime.org/cully/mcp` with `cully mcp add --agent codex --url https://mcp.mcpruntime.org/cully/mcp --oauth`, then run `codex mcp login cully`. The client follows Cully's protected-resource metadata to MCP Auth, which sends the browser to Keycloak. MCP Auth returns a resource-bound token; read tools require `tools:read` and write/delete tools require `tools:write`.

The release workflow checks MCP Auth discovery and JWKS, both tool scopes, Cully's protected-resource metadata and an unauthenticated `401` challenge. After deployment, complete one interactive sign-in and authorized tool call from an agent to verify the Keycloak callback and claim mapping; metadata checks alone cannot prove those. The [MCP Runtime OAuth guide](https://docs.mcpruntime.org/mcp-oauth/) describes the platform connector setup. The Docker options below apply when you operate the stack yourself.

The [Docker Compose stack](/hosting#one-command-full-stack) starts Caddy for HTTPS in OAuth mode and replaces the fixed owner with token subjects. Set public DNS and allow ports 80/443 to reach the host before starting Caddy. Keep the data API and PostgreSQL private.

The **agent** only needs the MCP URL and `--oauth` at setup. The **operator** needs the issuer, exact resource URL, JWKS URL and HTTPS routing before starting the server. A new MCP Auth broker additionally needs a selected upstream connector, signing key, persistent state and upstream client secret. The full-stack `./setup.sh` script starts Mem0 and both databases in either OAuth option after these values are configured.

## Self-hosted Docker with an existing authorization server

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

## MCP Runtime metadata

The active [`.mcp/servers.yaml`](https://github.com/mcp-runtime/cully/blob/main/.mcp/servers.yaml) is the personal **public OAuth deployment**, with `auth.mode: oauth`, `CULLY_MCP_AUTH_MODE=oauth`, an explicit issuer, audience and JWKS URL, and no fixed owner. MCP Runtime derives and injects the matching MCP resource and issuer values into the container. Its ingress, audience and issuer must match. An [isolated no-OAuth metadata example](https://github.com/mcp-runtime/cully/blob/main/.mcp/examples/no-oauth/servers.yaml) shows the same image with `auth.mode: none`, a fixed owner from a secret, a private ingress host and the private data API token. Copy and adapt that file only for a private single-user deployment; MCP Runtime network policy must keep the hostname inaccessible to untrusted callers. Relevant server-entry fields:

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
