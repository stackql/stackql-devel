# Caller access: four executable walkthroughs

These walkthroughs make no cloud requests. Each provisions a fresh SQLite principal directory,
configures one authentication mode, then runs:

```sql
SELECT whoami() AS principal, groups() AS memberships;
```

Expected identity: principal `reader`, directory group `readers`. Requirements: StackQL on `PATH`,
Python 3 and OpenSSL. The script uses only Python's standard library. Run from the repository root.
Use a different empty directory for each run; existing credentials are never overwritten.

## 1. JWT

```sh
python3 docs/examples/access-management/quickstart.py --mode jwt --directory /tmp/access-jwt
```

Creates an Ed25519 signing key, pins its public key and `kid` in `access.json`, issues a short-lived
JWT with the configured issuer/audience, then supplies it through `STACKQL_CALLER_TOKEN` to `exec`.
The directory maps `(access-demo, reader)` to principal `reader`. Token group claims grant nothing.
In a deployment, obtain tokens from your issuer and configure its trusted public keys instead.

## 2. Static bearer token

```sh
python3 docs/examples/access-management/quickstart.py --mode bearer --directory /tmp/access-bearer
```

Generates a random token. Only its SHA-256 digest and subject go into `access.json`. The credential
is passed through the child process environment; the same issuer/subject directory mapping supplies
permissions. A deployment provisions token digests administratively and supplies plaintext tokens
through its secret-management environment. Restart after changing configured digests.

## 3. Local OS identity

```sh
python3 docs/examples/access-management/quickstart.py --mode local_os --directory /tmp/access-os
```

Maps the effective Unix UID or Windows SID to principal `reader`. No token is required. OS usernames,
OS groups and client-supplied names do not grant permissions: directory memberships remain authoritative.
Docker uses the container process's effective UID. HTTP callers cannot use the server's OS identity.

## 4. Mutual TLS

```sh
python3 docs/examples/access-management/quickstart.py --mode mtls --directory /tmp/access-mtls
```

Creates a demonstration CA, server certificate and client certificate. It starts an HTTPS MCP server,
connects with the client certificate and verifies the server CA. The client URI SAN
`spiffe://access-demo/reader` maps to directory principal `reader`. No bearer token is sent. The server
requires a verified client chain and exactly one configured SAN, then stops after the query.
In a deployment, use your managed CA and certificates; configure `client_ca_pem` and `uri_san` or `dns_san`.
Common Name is never an identity fallback.

## Inspect or deploy the configuration

Each directory contains `access.json`, `backend.json` and `principals.db`. The mTLS walkthrough also
writes `mcp.json`. The generated private keys are for this local demonstration. Tokens are never printed
or stored as plaintext configuration. The script's child token disappears when that child exits.

The actual CLI configuration form is:

```sh
stackql --access.config '{"path":"/etc/stackql/access.json"}' \
  --sqlBackend '{"dsn":"/var/lib/stackql/principals.db"}' \
  exec 'SELECT whoami(), groups();'
```

JWT/bearer deployments must set `STACKQL_CALLER_TOKEN` before this command. Shell and MCP stdio use
that same environment. For HTTPS, send the token in `Authorization: Bearer ...`, or use a verified
client certificate when `type` is `mtls`. See [the configuration reference](../../access_management.md)
for complete mode examples, PostgreSQL provisioning and transport restrictions.

The same walkthrough script is executed by four individually named tests in
[`stackql_access.robot`](../../../test/robot/functional/stackql_access.robot). Additional scenarios
exercise each mode against real CLI/MCP transports and SQLite/PostgreSQL, including invalid credentials,
revocation, isolation and denied effects. Docker tests run the script against the actual tested image
through a host-side command adapter; the standard StackQL image does not acquire a Python dependency.
