
# `omnisdk` staging

`omnisdk` results are cursor streams. When the query plan can preserve SQL semantics incrementally, batches can flow directly to the end user. Otherwise, batches can be staged in an RDBMS for relational operations such as ordering, aggregation, and set operations.

Unlike the eager, per-row RDBMS ingestion used by the `any-sdk` path, Omni staging is conditional and batched. The staging tablespace sits alongside StackQL-owned relations in the selected RDBMS, whether embedded SQLite or TCP-routed PostgreSQL.

## High level details of staging

Let $Q$ be a StackQL query and let $R_i$ be the result relation of its $i$th logical Omni-backed input. An Omni plan may contain multiple exchanges, but it is one producer of $R_i$; exchange boundaries do not create staging tables.

The cursor partitions each result into batches without changing its contents:

$$
R_i = B_{i,1} \mathbin{\|} B_{i,2} \mathbin{\|} \cdots \mathbin{\|} B_{i,n_i}
$$

Here $\|$ denotes concatenation in cursor order. Batch size affects transport and insertion cost, not the relational result. A downstream operator that can be evaluated incrementally with the available streaming state consumes these batches directly. If an operator needs its complete input to produce the required result, the planner stages only the logical input relations that operator needs:

$$
T_{Q,i} = \biguplus_{k=1}^{n_i} B_{i,k}
$$

The RDBMS then evaluates the blocking operation over the staged relations, and its result cursor becomes the output stream. For example, a global `ORDER BY` may require all candidate rows before the first output row is known; a grouping or set operation may likewise require an RDBMS boundary when it cannot be evaluated within the chosen streaming-state bound. Operators that do not require that boundary continue to stream. The planner, not the mere presence of an Omni exchange, determines where materialization is needed.

Each query owns a disjoint staging namespace $S_Q$ in the RDBMS. Names are stable for a logical relation within $Q$ and collision-resistant across concurrent queries. Thus rows from two queries cannot share a staging relation, and cleanup is scoped to the query that created those relations. A multi-exchange Omni plan does not require one table per exchange; separate tables are needed only for distinct logical inputs that the blocking relational operation must evaluate separately.

Reading a batch and inserting it form a back-pressured pipeline: the next batch is requested as the staging consumer is ready, rather than eagerly loading every result into application memory. Cursor exhaustion marks completion; query ownership tracks staged relations for cleanup on success, error, or cancellation. No per-exchange counters are required for correctness. When no blocking operation requires materialization, $S_Q$ remains unused and Omni rows flow directly to the result consumer.

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
