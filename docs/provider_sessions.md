# Provider sessions and secrets

Provider aliases bind a canonical provider to an explicit tenancy, credential and request defaults.
Configuration belongs to the client session; independent SQL connections do not share mutations.
MCP request contexts are isolated; use one SQL batch when configuring and querying through MCP.
These features use Omni's session APIs. SQL metadata lookup resolves the canonical provider while
execution retains the alias, credential and provider/tenancy concurrency controls.

## Configure and query two accounts

Use double-quoted configuration paths and JSON objects inside SQL string literals:

```sql
SET "providers.production" = '{
  "provider": "aws",
  "tenancy": "production-account",
  "auth": {
    "type": "aws_signing_v4",
    "access_key_id_env_var": "PRODUCTION_AWS_ACCESS_KEY_ID",
    "secret_access_key_env_var": "PRODUCTION_AWS_SECRET_ACCESS_KEY"
  },
  "params": {"region": "ap-southeast-2"},
  "tuning": {"max_concurrent": 4, "max_per_host": 4}
}';
SET "providers.development" = '{
  "provider": "aws",
  "tenancy": "development-account",
  "auth": {
    "type": "aws_signing_v4",
    "access_key_id_env_var": "DEVELOPMENT_AWS_ACCESS_KEY_ID",
    "secret_access_key_env_var": "DEVELOPMENT_AWS_SECRET_ACCESS_KEY"
  },
  "params": {"region": "ap-southeast-2"},
  "tuning": {"max_concurrent": 2}
}';
SHOW SERVICES IN production;
SELECT * FROM production.ec2.vpcs;
SELECT * FROM development.ec2.vpcs;
```

Provider and tenancy are required. Tenancy is an explicit grouping label, not an account identity
inferred or verified by StackQL. Aliases sharing a provider/tenancy share their HTTP request limit
within one session and must declare identical tuning. `max_concurrent` limits total in-flight HTTP
requests across hosts; `max_per_host` limits each host. Zero uses Omni's defaults; a negative
`max_concurrent` explicitly removes the total limit. SQL predicates and mutation assignments override
alias parameter defaults. Binding parameters satisfy declared required inputs during SQL analysis.

Alias auth uses Omni's execution `Auth` JSON fields. It must be explicit; it does not borrow a
canonical provider's global auth configuration. Each execution receives a detached credential
snapshot. Credential edits affect subsequent queries; existing executions retain their snapshot.
Changing structural bindings or tuning starts a new governor generation.

```sql
SET "providers.production.auth.credentials" = 'replacement-token';
UNSET "providers.production.auth.credentials";
UNSET "providers.production";
```

Aliases work with document-driven reads, metadata commands, star expansion, explicit EXEC and
ordinary mutations. They do not add alias support to Omni's native built-in catalog or IaC resource
addresses. Provider-specific asynchronous completion paths still require their supported canonical
provider contracts; this change does not extend those contracts.

## Persistent secret storage

Persistent secrets are **disabled by default**. An operator must start StackQL with
`--allow-persistent-secrets` to enable the OS backend. SQL cannot enable this permission.
Flat-file stores, paths and an omitted backend are rejected even with the flag enabled.
In-memory provider/auth configuration remains available without this flag.

```sql
SET "secret_store" = '{"backend":"os","namespace":"stackql/provider-auth"}';
SET "secrets.production" = '{"type":"bearer","credentials":"example-token"}';
SET "providers.production" = '{
  "provider":"example_provider",
  "tenancy":"production-account",
  "auth_secret":"production"
}';
UNSET "secrets.production";
```

The OS backend uses Keychain on macOS, Secret Service on Linux (`secret-tool` and a
working session service), or Credential Manager on Windows. It fails when unavailable;
there is no file fallback. Namespaces separate keys but are not per-principal ACLs.
The OS store operates under the server process's OS identity. Enabling it does not
provide per-secret authorization between shared-server callers.

Secret values are strings, including JSON strings. `auth_secret` identifies a JSON
`Auth` object resolved for a query snapshot. Rotation invalidates the session's credential
cache generation. Missing secrets fail rather than borrowing another credential.
Ordinary provider/auth SET statements never write the store. SQL revelation separately
requires `--allow-secret-dereference`. Avoid secret literals in command arguments/history.

Robot OS-store scenarios mock the Keychain/Secret Service command boundaries on Unix
so CI does not require an unlocked desktop keyring. Native Windows uses Credential Manager.
These tests verify the CLI/store contract, not the security of a real Unix keyring service.

## Legacy auth configuration and reset

Existing `auth.<provider>` paths retain the legacy StackQL AuthCtx field names:

```sql
SET "auth.example" = '{"type":"basic","username":"operator","password":"example"}';
SET "auth.example.password" = 'replacement';
UNSET "auth.example.password";
```

Whole-object SET replaces that object. Field SET updates one field. UNSET removes a field or
binding; whole legacy auth reset leaves an empty explicit DTO. Unknown fields and wrong types fail
without changing the object. Multiple assignments execute in order and stop at the first error;
the statement is not a transaction across assignments or persistent stores.

Parser UNSET normalizes to the existing SET AST with DEFAULT-valued expressions. Thus
`SET "providers.production" = DEFAULT` is equivalent to UNSET. The pinned parser supports unqualified
UNSET and SET SESSION. Scoped UNSET requires the companion parser changes to be released and pinned;
no repository go.mod replacement is supplied.

When caller access management is enabled, configuration changes require `session/config` `write`
permission as well as SQL statement permission. Alias execution additionally requires
`tenancy/<canonical-provider>/<tenancy>` `use` and the existing operation/credential permissions.
Session configuration remains opt-in; normal sessions are unaffected until configuration is set.

## Scope

Assignment targets omit `$`; that sigil is reserved for dereferencing values.
Both commands accept an optional scope word:

```sql
SET SESSION "auth.example.password" = 'replacement';
UNSET SESSION "auth.example.password";
```

Unqualified commands use SESSION. GLOBAL parses but is rejected by StackQL before changing
configuration. Principal/group scopes and their inheritance rules are not implemented.

## Scalar variables and dereferencing

Variables are local to the current session. Assign them without a dollar on the left:

```sql
SET "variables.region" = 'east';
SET "variables.attempts" = 3;
SET "variables.enabled" = true;
SET "variables.optional" = null;
SELECT upper($region) || '-primary' AS region, $attempts + 1 AS next_attempt,
       $enabled AS enabled, $optional IS NULL AS absent;
UNSET "variables.region";
```

`$name` reads `variables.name`. Quoted references can name configuration paths, for example
`"$providers.production.params.region"`. Ordinary string literals containing `$` are unchanged.
References are resolved once per statement, before provider dispatch, and become typed AST literals.
SQL functions, casts, comparisons and concatenation operate normally. Numbers retain their literal
precision at substitution; the selected SQL backend still determines arithmetic precision.
Only strings, numbers, booleans and null are supported. Objects/arrays are rejected, and a missing
or unset value is an error rather than null. Prepared statements resolve variables at execution.
Statements containing references and SET commands do not reuse cached execution plans.

Reading auth fields or secret-store values is disabled by default. Trusted host startup flag
`--allow-secret-dereference` opts into exposing these values through SQL, including results,
logs and other execution artifacts. SQL cannot enable this flag. For example, after trusted startup
opts in, `SELECT upper("$secrets.example");` reads the stored string; a structured secret can be read
at a scalar field such as `"$secrets.example.credentials"`. Internal provider credential resolution
continues to work without exposing credentials to SQL. Protected sessions additionally require
`session/variables` `read` policy permission.

## Verification

`test/robot/functional/stackql_provider_sessions.robot` runs the real CLI against a loopback provider.
It checks account credentials, required scope defaults, explicit SQL override, metadata, SELECT *,
secret rotation/deletion and in-memory auth reset. Handler race tests check detached snapshots and
isolation between clients. No live cloud or OS credential store is used by those tests.
