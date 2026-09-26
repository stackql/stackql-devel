
# omnisdk shortcomings

## High Level Expectations

1. `omnisdk` will accept existing `any-sdk` auth structures and default configs in provider documents and behave as expected, can be observed in `any-sdk`.  Behviours should be implemented protocol agnostic and best practice.  Tech debt from `any-sdk` must not be inherited by copy or imitation.  **`omnisdk` will never import from `any-sdk`, directly or indirectly.**
2.  This is very important.  Whereas `nay-sdk` has been tightly coupled to `http`, with some subprocess variants, by contrast `omnisdk` is **protocol agsnostc**.  Full support for proticol buffers, `http`, subprocess calls, streaming protocols, whatever old and new transports, are intrinsic to `omnisdk`.
3. Any WAL, ledgers and the like should be configurable in location, substrate (eg: local vs, s3, different os...) and the like.  Must be extensible and abstacted, no excuses.
4. `omnisdk` will support all functionality in `any-sdk`, but with clean implementation.  Breaking changes will be specified ahead of time.  The api is wildly different, but functional coverage will not be lesser.
5. `omnisdk` does not stage results in RDBMS or otherwise, it eagerly streams generated records.  This is a significant and highly beneficial difference to `any-sdk`.  Any ordering, aggregation or set operations (union and the like) is imposed post `omnisdk` query fulfilment.
6. `omnisdk` is required to support SQL extension funstions, including: scalar, redord and table valued ones.  We do not  expect the very first version to have total coverage and so some queries may need to be routed away from `omnisdk` at times, by logic within `stackql`.
7. `omnisdk` is to support all of the request and response trandsform grammars and shorthands already in `any-sdk`.
8. `omnisdk` will at some point support both SQL style saga rollbacks and also an IAC saga variant with fine grained locking and abstracted latches that are objectively superior to `terraform` style crude locks and failure modes.  That said, early IAC forms are per stack locked.  Patience is our watchword.

## Migration plan at coarse grain

- (a) Support joins, sql functions consuming only `omnisdk`, in `stackql_unstable-<provider>` namespace.  It may take some time to acheive full covereage, in stages.
- (b) Cut versions and releases of both `omnisdk` and `stackql` along the way as useful milestones are reached.  


## Current issues

- (i) Auth has to function as expected.  No excuses.  The existing stackql patter with auth structures and docs simply **must** work.
- (ii) We need an orderly abstraction and catalogue of supported SQL extension functions in `omnisdk`.  This will be in a discrete `pkg`.  See below `Expected SQL extension functions` section.
- (iii) We need an orderly abstraction and catalogue of supported request and response processing grammars and shorthands in `omnisdk`.  These are expected to mirror the `main` branch of `any-sdk` but be cleanly implemented in a discrete `pkg` with minimal dependencies and **zero** relation to `any-sdk`.  See below `Expected Transformation Grammars and Shorthands` section.
- (iv) We want support for user/agent composed cross cloud rapid audit queries.
- (v) I want an SOC or whatever corporate audit query suite asap.

## Supporting information

### Expected SQL extension functions

The required SQL extension functions are precisely those that are robot tested anywhere in the `stackql` codebase.

### Expected Transformation Grammars and Shorthands

The required transformation grammars (eg for http request and response) are precisely those that are implemented in the `main` branch of the `any-sdk` codebase.

