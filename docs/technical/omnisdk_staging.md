
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




## Terse comaprison

| Aspect | any-sdk | omnisdk | Commment |
|----|----|----|----|
| RDBMS Ingestion | Per API response "record" | Per query.  AOT configurable and runtime responsive batching | omnisdk clearly better performance best case |
| SQL translation of relation names | Recursive: global and per API call |  Per query and flat | omnisdk has appealing simplicity and debug property of thin layer queries  |
| SQL control counters | Mutiple counters and recursively applied: global and per API call |  Per query and single counter only |  |
| SQL secure storage | - | - |  We have decided not to address security or obfuscation at this time.  This will happen in future versions |
