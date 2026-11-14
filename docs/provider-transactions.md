# Provider transactions

The opt-in `$.transaction` path delegates scheduling, journal/ledger writes,
locking, compensation and recovery to Omni. stackql retains SQL translation,
authentication and connection-bound handle storage.

```sql
SET "$.transaction" = '{"collection":"demo","mode":"lazy","keys":["item"],"targets":{"stackql_unstable_demo.service.items":{"key":"item","identity":"id","addressed_by":"id"}}}';
BEGIN;
-- provider INSERT / SELECT
SELECT * FROM stackql_preview.txn.status;
COMMIT; -- or ROLLBACK
```

`mode` accepts `eager` or `lazy`. If omitted, existing session `rollback_type`
configuration chooses execution timing. Reads and RETURNING flush preceding writes.
Rollback discards queued work and compensates executed work under retained locks.
`state` names an optional directory under the same caller/tenant IaC state root.

Recovery: reconnect, configure the same collection/state with `resume_id`, then BEGIN.
Read status in the resumed session before proceeding; resume advances its version. For an uncertain statement, verify the provider state,
then call `stackql_preview.txn.recover` with `version`, JSON `resources` containing
`key`, `outcome` (`applied`/`absent`), observed resource `version`, and verified identity
for applied creates. `continue = 'true'` explicitly attests statement completion;
false aborts it for subsequent rollback. Recovery sends no provider request.

Current limits: authenticated callers are required; only creates with a declared delete
inverse are supported. Unsupported mutations fail before dispatch. Configuration cannot
change inside a transaction; nested transactions are rejected. Consume read cursors before
COMMIT/ROLLBACK. Provider aliases are unsupported; each resource address currently binds
one logical target. Leases default to 300 seconds; execution lifetime defaults to 900 seconds.
Stop or isolate the previous executor before takeover: leases cannot fence provider
requests already in flight. Legacy unconfigured sessions retain their previous coordinators.

The new adapter requires an Omni release exposing `TransactionSession`. Local validation
uses a build-only modfile; pin the published release before normal CI. Protocol and
invariants: [Omni transactions](https://github.com/stackql-labs/omnisdk/blob/main/docs/developer/transactions.md).
