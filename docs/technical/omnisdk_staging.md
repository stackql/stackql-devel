
# `omnisdk` staging

`omnisdk` queries produce eagerly written, *mostly* unorded streams of data.  For queries lacking in post-hoc: ordering, aggregation and set operations; then output can be sent straight to the end user.

For query outputs that are not thus suitable for direct display, a staging into an RDBMS (eg: embedded `sqliute`, or `tcp`-routed `postgres`) is indicated.

Because the output is not staged strictly per-row eagerly like `any-sdk` output, there is scope for flexible and runtime adaptive batching policy.

## High level details of staging

TBA.
