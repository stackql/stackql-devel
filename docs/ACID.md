
# ACID

`stackql` supports transactional semantics, up to a point.  As such, implementations of all of the [ACID](https://en.wikipedia.org/wiki/ACID) characteristics underpin query execution and application lifecycle: Atomicity, Consistency, Isolation and Durability.  `stackql` ACID characteristics are configurable along the following dimensions:

- Evaluation.  When statements in the transaction are executed.  Effectively an enumeration of `{ Eager, Lazy }`.  This affects both [atomicity](#atomicity) and [durability](#durability).
- Recovery. How to prosecute redo or undo actions.  This affects primarily [durability](#durability).

## Atomicity

For each reversible action, the reversal operation defines an edge on an abstract DAG (potentially a DAG of DAGs of low level operations) from the current state to the prior state.  For convenience, let us call this the "undo DAG". Execution of this DAG in full constitutes a successful `ROLLBACK` operation.  

In similar vein, execution of all unactioned edges on an abstract DAG (the "redo DAG") from current to target state constitutes a successful `COMMIT` operation.

Serializations of undo and redo DAGs are written to [WAL](#wal).

## Consistency

Transactions will accept only permissible operations, and runtime errors can trigger undo, thus impacting [atomicity](#atomicity), depending upon recovery configuration. 

A naive implementation of consistency does not support SQL-style constraints (eg: foreign key, unique) and simply offloads such consistency aspects to the providers.  An example aspect that can be handled this way is referential integrity, treated as implicit and presumed enforced by the provider.

## Isolation

Records in the relational algebra backend will contain identifiers for transaction ID and other session information.  Only those records that are relevant to the query and user are exposed.

## Durability

`stackql` atomicity is configurable along the following dimensions, which also relate to [durability](#durability):

- Evaluation.  When statements in the transaction are executed.  Effectively an enumeration of `{ Eager, Lazy }`.
- Recovery. How to prosecute redo or undo actions 

## WAL

[Write Ahead Logging (WAL)](https://en.wikipedia.org/wiki/Write-ahead_logging) is the cornerstone of atomicity and durability.
