

# The Transaction Storage Manager (TSM)

The TSM [^1] is conventionally regarded as a monolith for co-ordinating:

- Recovery via a `log manager`.  WAL is the canonical example.
- Concurrency via a `lock manager`.  Two classes are: (i) strict two phase locking (S2PL) and (ii) multi version concurrency control (MVCC).
- Data access via `access methods`.  Traditionally: B+ tree and heap file(s).
- Buffer pool management via a `buffer manager`.  

For the most part, these components are tightly coupled.  Some illustrative examples of why this is the case and necessary:

- Canonical B+ tree data structure includes latches for concurrent access.  Such latches are aspects of the lock manager.
- `physical` logs better support atomic disk updates and therefore the log manager requires access method internals.
- The log manager requires lock lifetime coupling in order to safely support undo / redo.

Anecdotally, the buffer manager component is not so tightly coupled to the remainder and can potentially be replaced in modular fashion.


# ACID

`stackql` supports transactional semantics, up to a point.  As such, implementations of all of the [ACID](https://en.wikipedia.org/wiki/ACID) characteristics underpin query execution and application lifecycle: Atomicity, Consistency, Isolation and Durability.  `stackql` ACID characteristics are configurable along the following dimensions:

- Evaluation.  When statements in the transaction are executed.  Effectively an enumeration of `{ Eager, Lazy }`.  This affects both [atomicity](#atomicity) and [durability](#durability).
- Recovery. How to prosecute redo or undo actions.  This affects primarily [durability](#durability).

## __`stackql`__ and the CAP Theorem

In a general sense, `stackql` is a distributed data store.  With regard to the [CAP theorem](https://en.wikipedia.org/wiki/CAP_theorem), `stackql` does not meaningfully address partition tolerance.  If we defer / ignore user space constraints and offload the bulk of consistency to the provider systems, then that leaves availability (the `A` in `CAP`), which can be optimised using standard approaches for failover, disaster recovery, load balancing, etc.  This effectively dictates that prolonged network partitioning or provider defects supporting consistency may require admin intervention.

## Atomicity

For each reversible action, the reversal operation defines an edge on an abstract DAG (potentially a DAG of DAGs of low level operations) from the current state to the prior state.  For convenience, let us call this the "undo DAG". Execution of this DAG in full constitutes a successful `ROLLBACK` operation.  

In similar vein, execution of all unactioned edges on an abstract DAG (the "redo DAG") from current to target state constitutes a successful `COMMIT` operation.

Serializations of undo and redo DAGs are written to [WAL](#wal).

### Two phase commit

[Two phased commit (2PC)](https://en.wikipedia.org/wiki/Two-phase_commit_protocol) is supported in `stackql`, through the `transact.Coordinator` interface.  Briefly, 2PC is (affirmative; `COMMIT`,  or negative; `ROLLBACK`) transaction finalization by sequential phases of:

- (i) Voting.
- (ii) Completion.

The role of the transaction coordinator, sometimes called a transaction manager, is to tally all required votes and then mark the transaction as complete.

For now, we do not expand upon exactly how `transact.Coordinator` implementations should do this.  Configurable commitment behaviour is supported in `stackql`.

## Consistency

Transactions will accept only permissible operations, and runtime errors can trigger undo, thus impacting [atomicity](#atomicity), depending upon recovery configuration. 

A naive implementation of consistency does not support SQL-style constraints (eg: foreign key, unique) and simply offloads such consistency aspects to the providers.  An example aspect that can be handled this way is referential integrity, treated as implicit and presumed enforced by the provider.

## Isolation

Records in the relational algebra backend will contain identifiers for transaction ID and other session information.  Only those records that are relevant to the query and user are exposed.

### Read Uncommitted

For the `read uncommitted` isolation level, any available update on the target relation is considered.  The definition of "available" will need to be expanded to cover that for which an update "step" is completed.

### Read committed

For the `read committed` isolation level, there is an additional requirement that the update transaction has been completed.

### Repeatable read 

In the case of `repeatable read`, all of the preceding apply, plus the `{ transaction, step }` for a given relation must be repeated throughout all the `select` queries in the consuming transaction.

### Serializable

All above apply.  Reads on subsequent transactions cannot be initiated until preceding dependency transactions are committed.  Some transactions will have to be stalled or pre-empted in order to avoid deadlock situations.

## Durability

`stackql` atomicity is configurable along the following dimensions, which also relate to [durability](#durability):

- Evaluation.  When statements in the transaction are executed.  Effectively an enumeration of `{ Eager, Lazy }`.
- Recovery. How to prosecute redo or undo actions 

## WAL

[Write Ahead Logging (WAL)](https://en.wikipedia.org/wiki/Write-ahead_logging) is the cornerstone of atomicity and durability.

### Implementation of WAL

Modern WAL is often comprised of a combination of physical and logical logging, dubbed "physiological" [^1].  For `stackql`, the closest one can imagine coming to physical logging is fully expanded database update operations, ie: independent and fully parameterized RDBMS queries.  This is scarcely the sort of tight coupling between WAL, buffer pool pages and disk sectors that traditional RDBMSs implement.  Let us then, at least for early versions of `stackql`, consider all logging to be logical.

Feasibly, `stackql` can model logical logging as:

1. HTTP requests, exclusive of secrets, with some indirection to credentials and token generation flow.  This latter is generally core `stackql` code or SDK code. ~~Because request multiplicity may be dependent on upstream data flows, this will operate at some sort of templated level in many cases.~~
2. RDBMS transactions.  These could be quite large in the context of `rollback` operations. 

The consideration of security for inputs to RDBMS transactions is delicate.  A naive implementation is to regard logs as insecure and wrap in encryption machinery as required.

In order to simplify logical logging, we postulate some mnemonic representation (MR) to describe the collections of DAGs that comprise `stackql` transactions.  This MR, desirably, should be:

- (a) Deterministic and repeatable.  This simplifies reasoning about executions and estimating costs.
- (b) Secure.  We do not want to expose credentials or sensitive information.
- (c) Suitable for all recovery scenarios.  Foreseeably, undo and redo operations may need to run in different orders, eg: depending on network partitioning or provider degradation.

To this end, we propose `WAL v1`.  The key design points are:

- (i) Logs are at primitive level, for their finest grain, in most cases.  If data caches are warranted and absent any security concerns, then caches may be used.
- (ii) DAG ordering of primitives will be in WAL.  Point to note: there is no reason that DAGs cannot be re sorted or dependencies added / subtracted for any reason (eg: performance, handling network partition events). 
- (iii) As a matter of principle, credentials and any other ephemeral tokens are not persisted in WAL.  `v1` will not include any secure secret storage, so this is sensible.
- (iv) SERDE between WAL and DAGs should be a simple and modular as possible.  Also extensible.
- (v) Let us re-use design points from [the existing __`postgres`__ WAL implementation](https://www.postgresql.org/docs/16/wal.html).  We do not want to re-invent the wheel:
    - A devoted WAL directory.
    - A control data structure records the most recent checkpoint.
    - Individual files ("segment" files in `postgres`) are discarded / renamed once obsoleted.
- (vi) Depending on ACID configuration, UNDO logs for committed transactions have the  same lifetime as REDO logs for the same transaction.

## SQL and VMs

We *can* incorporate some of the thinking and opcodes from [SQLite bytecode](https://www.sqlite.org/opcode.html), although it need not be so fulsome, in light of heavy pushdown to RDBMS.

## Implementations of ACID

### __Implementation A__: Naive 

Locks, latches and lookup tables can probably handle most of this.  

Per [this `stackoverflow` post](https://stackoverflow.com/questions/3111403/what-is-the-difference-between-a-lock-and-a-latch-in-the-context-of-concurrent-a#:~:text=Locks%20ensure%20that%20same%20record,consistency%20of%20the%20memory%20area.), which sites a paper by Graefe ([^2]):

> Locks
> → Protects the index’s logical contents from other txns.
> → Held for txn duration.
> → Need to be able to rollback changes.
> 
> Latches
> → Protects the critical sections of the index’s internal data structure from other threads.
> → Held for operation duration.
> → Do not need to be able to rollback changes.


[^1]: Hellerstein JM, Stonebraker M, Hamilton J; "Architecture of a Database System"; Foundations and Trends in Databases; Vol. 1, No. 2 (2007); 141–259
[^2]: Graefe G; "A Survey of B-Tree Locking Techniques"; ACM Transactions on Database Systems; Vol. 35, No. 3 (2010); Article 16

