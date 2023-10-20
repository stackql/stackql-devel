



# ACID

`stackql` supports transactional semantics, up to a point.  As such, implementations of all of the [ACID](https://en.wikipedia.org/wiki/ACID) characteristics underpin query execution and application lifecycle: Atomicity, Consistency, Isolation and Durability.  `stackql` ACID characteristics are configurable along the following dimensions:

- Evaluation.  When statements in the transaction are executed.  Effectively an enumeration of `{ Eager, Lazy }`.  This affects both [atomicity](#atomicity) and [durability](#durability).
- Recovery. How to prosecute redo or undo actions.  This affects primarily [durability](#durability).

## __`stackql`__ and the CAP Theorem

In a general sense, `stackql` is a distributed data store.  With regard to the [CAP theorem](https://en.wikipedia.org/wiki/CAP_theorem), `stackql` does not meaningfully address partition tolerance.  If we defer / ignore user space constraints and offload the bulk of consistency to the provider systems, then that leaves availability (the `A` in `CAP`), which can be optimised using standard approaches for failover, disaster recovery, load balancing, etc.  This effectively dictates that prolonged network partitioning or provider defects supporting consistencey may require admin intervention.

## Atomicity

For each reversible action, the reversal operation defines an edge on an abstract DAG (potentially a DAG of DAGs of low level operations) from the current state to the prior state.  For convenience, let us call this the "undo DAG". Execution of this DAG in full constitutes a successful `ROLLBACK` operation.  

In similar vein, execution of all unactioned edges on an abstract DAG (the "redo DAG") from current to target state constitutes a successful `COMMIT` operation.

Serializations of undo and redo DAGs are written to [WAL](#wal).

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

## Implementations of ACID

### __Implementation A__: Naive 

Locks, latches and lookup tables can probably handle most of this.  

Per [this `stackoverflow` post](https://stackoverflow.com/questions/3111403/what-is-the-difference-between-a-lock-and-a-latch-in-the-context-of-concurrent-a#:~:text=Locks%20ensure%20that%20same%20record,consistency%20of%20the%20memory%20area.):

> From CMU 15-721 (Spring 2016), lecture 6 presentation, slides 25 and 26, which cites [A Survey of B-Tree Locking Techniques by Goetz Graefe](https://15721.courses.cs.cmu.edu/spring2016/papers/a16-graefe.pdf):
> 
> Locks
> → Protects the index’s logical contents from other txns.
> → Held for txn duration.
> → Need to be able to rollback changes.
> 
> Latches
> → Protects the critical sections of the index’s internal data structure from other threads.
> → Held for operation duration.
> → Do not need to be able to rollback changes.




