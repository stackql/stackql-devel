
# `omnisdk` staging

`omnisdk` results are cursor streams. When the query plan can preserve SQL semantics incrementally, batches can flow directly to the end user. Otherwise, batches can be staged in an RDBMS for relational operations such as ordering, aggregation, and set operations.

Unlike the eager, per-row RDBMS ingestion used by the `any-sdk` path, Omni staging is conditional and batched. The staging tablespace sits alongside StackQL-owned relations in the selected RDBMS, whether embedded SQLite or TCP-routed PostgreSQL.

## High level details of staging

Let $Q$ be a StackQL query with query ID $q$. Omni executes the joins and exchanges in its own query plan and produces a single final relation $R_Q$. An Omni plan may contain multiple exchanges, but exchange boundaries do not create staging tables or query IDs.

The cursor partitions $R_Q$ into batches without changing its contents:

$$
R_Q = B_{1} \mathbin{\|} B_{2} \mathbin{\|} \cdots \mathbin{\|} B_{n}
$$

Here $\|$ denotes concatenation in cursor order. Batch size affects transport and insertion cost, not the relational result. The SQL operators that remain after Omni planning (for example global `ORDER BY`, aggregation, `DISTINCT`, set operations, or a `LIMIT` that depends on ordering) are evaluated over $R_Q$. If each of them can be evaluated incrementally within the chosen streaming-state bound, batches flow directly to the result consumer and no table is created. Otherwise $R_Q$ is staged, as a multiset, in exactly one query-owned table:

$$
T_{q} = \biguplus_{k=1}^{n} B_{k}
$$

The RDBMS then evaluates the remaining SQL over $T_q$, and its result cursor becomes the output stream. The planner, not the mere presence of an Omni exchange, determines whether materialization is needed.

A blocking operation whose inputs Omni cannot combine into one relation (for example a set operation over two independent Omni plans) is not staged as several tables under one query ID. Pre-analysis instead splits $Q$ into child queries $q_1, \ldots, q_m$; each child owns at most one table $T_{q_j}$, and the parent evaluates the operation over those tables. Child IDs therefore correspond only to real pre-analysis splits, and the invariant is one table per query ID.

Query IDs are allocated by the RDBMS itself (a PostgreSQL sequence, or an SQLite `AUTOINCREMENT` table), so they are collision free across concurrent queries and across processes sharing a backend. Rows from two queries never share a table, and cleanup is scoped to a query and its descendants.

Reading a batch and inserting it form a back-pressured pipeline: the next batch is requested only after the previous one is written, rather than eagerly loading every result into application memory. Cursor exhaustion marks completion; query ownership tracks staged tables for cleanup on success, error, or cancellation.

## Concrete descisions

### Decision 1: Savage cut in transaction control counters

The prior `any-sdk` implementation is supported by counters for:

- `Generation ID`.
- `Sessions ID`.
- `Transaction ID`.
- `Insert ID`.

The new, `omnisdk` implementation will require only one counter, `Query ID`.  This is because staging happens only per query.  We do reserve the right to split queries, so shall maintain a concurrency safe hierarchy store for `Query ID` parent-child associations.

### Decision 2: New tablespace for omnisdk staging



```bash
./build/stackql exec \
  --sqlBackend='{"dsn":"file:./stackql.db"}' \
  "SELECT
     v.vpc_id,
     s.subnet_id
   FROM aws.ec2.vpcs AS v
   INNER JOIN aws.ec2.subnets AS s
     ON v.vpc_id = s.vpc_id
   WHERE v.region = 'ap-southeast-2'
     AND s.region = 'ap-southeast-2';"
```

**Figure MQ-1** Model query 1.  A simple working query.

---

**Table T-1**: Tablespace comparison for model query MQ-1.

| sdk | RDBMS | tables |
|---|---|---|
| any-sdk | sqlite | `"aws.ec2.vpcs.generation_<v>"`, <br/> `"aws.ec2.subnets.generation_<s>"`  |
| any-sdk | postgres |  `"<table_schema>"."aws.ec2.vpcs.generation_<v>"`, <br/> `"<table_schema>"."aws.ec2.subnets.generation_<s>"` |
| omnisdk | sqlite |  `"__iql__.queries.<Query ID>"`  |
| omnisdk | postgres | `"<query_schema>"."<Query ID>"`  |

`<query_schema>` is configured with `schemata.querySchema` in `--sqlBackend` and defaults to `stackql_queries`.

## Decision 3: GC simplification for omnisdk

For omnisdk, any materializations needed will be created eagerly, **in the same place for prior** and all query tables can be marked for deletion immediately where no cache is in operation, or at whatever future time if cacheing is in effect.  We will need a robust mechanism in place from day 1, default to no cache.

This does imply a keyval store to look up tombstone times and find query ID by query plaintext.  Intuitively, I favour a new GC mechanism with the old one phased out when `any-sdk` is decommissioned.


## Terse comparison prior any-sdk vs omnisdk

| Aspect | any-sdk | omnisdk | Commment |
|----|----|----|----|
| RDBMS Ingestion | Per API response "record" | Per query.  AOT configurable and runtime responsive batching | omnisdk clearly better performance best case |
| SQL translation of relation names | Recursive: global and per API call |  Per query and flat | omnisdk has appealing simplicity and debug property of thin layer queries  |
| SQL control counters | Mutiple counters and recursively applied: global and per API call |  Per query and single counter only |  |
| SQL secure storage | - | - |  We have decided not to address security or obfuscation at this time.  This will happen in future versions |
