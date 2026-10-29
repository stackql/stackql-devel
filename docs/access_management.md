# Opt-in caller access and IaC

Caller access is disabled by default. Enable it with `--access.config /path/to/access.json`.
Caller credentials are independent of provider credentials and SQL `AUTH` statements.

## Configuration

The initial host integration validates JWTs using an explicit issuer, audience and pinned public
keys. For example:

```json
{
  "issuer": "https://identity.example",
  "audience": "stackql",
  "max_session_seconds": 300,
  "keys": [
    {"kid": "primary", "algorithm": "EdDSA", "public_key_pem": "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----\n"}
  ],
  "rules": [
    {"subject": "group:readers", "resource": "operation/demo/list/listItems", "action": "execute"},
    {"subject": "group:readers", "resource": "credential/demo", "action": "use"}
  ]
}
```

Use the real compiled method and operation identifiers from trusted provider documents. An operation
wildcard can grant mutation methods as well as reads. Keys are bound to their configured algorithms;
unknown key IDs, invalid signatures, missing/expired expiry and invalid temporal/issuer/audience
claims fail closed. Token-supplied URLs and groups are not used for trust or directory membership.
Configuration is loaded at startup; rotate policy or keys through a controlled restart. Sessions are
bounded by both token expiry and `max_session_seconds` (default 300; maximum 86400). In-flight SDK and
backend execution receives that deadline. A new statement resolves membership again.

## Principal directory

The configured SQLite or PostgreSQL backend stores:

- `stackql_access_principals`: ID and JSON object of string attributes.
- `stackql_access_identities`: verified `(issuer, subject)` to principal ID.
- `stackql_access_groups`: group IDs.
- `stackql_access_memberships`: principal/group bindings.

Schema creation is explicit: `accessmanagement.InitializeSchema(ctx, engine)` is an administration
API, not an authentication side effect. Provision tables, identities and memberships through a
trusted database administration path before serving protected queries. Unknown identities, absent
schema, malformed data and directory outages deny access. Ordinary protected SQL cannot address
directory tables. SQLite runtime connections enable foreign keys. Protect the database file and
use separate administrative/runtime database privileges where supported.

The directory produces an immutable per-statement snapshot. Omni receives that snapshot and an
authorizer; database connections, drivers and SQL remain in StackQL.

## Entry points and permissions

Local `exec`, shell and MCP stdio obtain the JWT from `STACKQL_CALLER_TOKEN`. Supply it through the
process environment, never a SQL literal or command-line argument. Protected `exec` failures return
exit status 1 after presenting the error.

HTTPS MCP obtains the bearer token from the actual request. TLS certificate/key configuration is
mandatory; the caller verifier replaces the legacy shared HTTP bearer gate when access is enabled.
Missing request credentials never fall back to the server's environment token.

Protected PostgreSQL-wire startup and MCP reverse proxy currently refuse explicitly. The current
wire dependency does not expose verified per-connection authentication evidence to the backend.
They must gain that integration before protection can be enabled on those transports. Unprotected
legacy deployments continue to operate as before. Provider SQL `AUTH`/`AUTH REVOKE` administration also refuses in protected mode; provision provider
credentials through trusted host configuration instead. Other standalone helper/registry commands refuse
when caller access is enabled, rather than bypassing the protected SQL path.

The `readers` directory group grants the host's safe-read SQL capability. SDK operations and provider
credential use still require explicit policy. Other statements require `execute` on
`host/sql/<statement-kind>`; reads outside the implicit reader group require `host/sql/read`.
Additional host checks include:

| Resource | Action | Scope |
| --- | --- | --- |
| `host/iac/run`, `host/iac/destroy` | `execute` | Effectful IaC relations, despite SELECT syntax |
| `host/sql/exec` | `execute` | Native EXEC, including nested EXEC |
| `host/function/<name>` | `execute` | Functions outside the small known-pure allowlist |
| `host/backend` | `read` | Opaque stored/external backend relations |
| `host/credentials` | `reload` | MCP credential reload administration |

SDK collection, operation, credential, relation and transport checks are additional requirements.
A collection grant does not grant its underlying methods. Stored query expansion is checked before
execution. Non-integrated legacy execution is refused while protection is enabled. Legacy plan and
row-cache reuse is disabled on protected paths.

Inspect the established snapshot using:

```sql
SELECT whoami() AS principal, groups() AS memberships;
```

## IaC surface

Use `stackql_preview.iac.run` and `.destroy` with the existing preview opt-in. Resource JSON uses
Omni's shared parser: `write_once` carries initial request inputs, `desired` carries maintained
state, and typed `input`, `env`, `self` and sibling bindings remain intact. Retired `params` and
`correlationParam` resource fields are rejected. Resource keys must be unique.

`inputs` and `environment` predicates accept JSON objects. JSON numbers retain their precision;
strings are literal values, not SQL or templates. Remaining predicates are provider execution scope.
Always supply explicit collection and cloud scope. The host-configured state root confines relative
`state` names, including checks against existing symlink escapes. Protected state is additionally
namespaced by the directory's `tenant` attribute, or by principal ID when no tenant is configured.
Only trusted directory administrators should assign tenant membership.

```sql
-- Explicit partial deletion. An empty selection is an error.
SELECT * FROM stackql_preview.iac.destroy
WHERE collection = 'deployment' AND state = 'production'
  AND keys = '["resource-key"]' AND run_id = 'prune-001';

-- Full teardown: omit keys.
SELECT * FROM stackql_preview.iac.destroy
WHERE collection = 'deployment' AND state = 'production' AND run_id = 'destroy-001';
```

Use unique run IDs. After pruning, remove the deleted declarations before applying again; omission
alone does not delete, and reapplying the original declaration requests recreation. Omni blocks
selected deletion when retained resources depend on it. Failed/error/outstanding result rows become
SQL failures and CLI failure status, while diagnostic rows remain available. `applied` can mean no-op.
Cancellation and authenticated context propagate to SDK execution, completion polls and compensation;
compensation does not retry anonymously after access expires or fails.

The black-box scenario is `test/robot/functional/stackql_access.robot`; it uses locally signed test
JWTs, a local API and the configured SQLite/PostgreSQL directory. Test tokens remain outside Robot
arguments and artifacts.
