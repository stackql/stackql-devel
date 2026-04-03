# Live Context: Extended Query Protocol Implementation

## Date: 2026-04-03

## Current State

### What's been done

1. **Indirect joins tests** (complete, merged-ready):
   - 8 new robot tests covering 3-way and 4-way INNER JOIN + LEFT OUTER JOIN across views, materialized views, subqueries, provider tables
   - All use `Should Stackql Exec Inline Equal Both Streams` with exact output matching
   - LEFT OUTER JOIN tests prove NULL behavior with partial matches
   - `docs/views.md` updated with supported combinations

2. **Extended query protocol stubs** (complete):
   - `basicStackQLDriver` implements `IExtendedQueryBackend` with compile-time check
   - `HandleParse`: passthrough (returns client OIDs as-is)
   - `HandleBind`: no-op
   - `HandleExecute`: `queryshape.SubstituteParams` replaces `$N` → `HandleSimpleQuery`
   - `HandleDescribeStatement/Portal`: delegates to `queryshape.Inferrer`
   - `HandleClose*`: no-ops
   - 407/407 robot tests pass

3. **`queryshape` package** (complete):
   - `internal/stackql/queryshape/queryshape.go`
   - Public `Inferrer` interface, private `standardInferrer` struct
   - `InferResultColumns(query)` with two paths:
     - `inferFromStoredRelation`: reads MV/table column metadata from stored DTOs (cheap)
     - `inferFromPlan`: builds plan via `planbuilder.BuildPlanFromContext` (no execution)
   - `SubstituteParams`: moved here from driver, replaces `$N` with bound values
   - `extractSingleTableName`: lightweight sqlparser-based single-table detection
   - Unit tests in `queryshape_test.go` with JSON testdata

4. **Plan column metadata** (complete):
   - `plan.Plan` has `GetColumnMetadata()`/`SetColumnMetadata()`
   - Set in `planbuilder/entrypoint.go` from `GetSelectPreparedStatementCtx().GetNonControlColumns()`
   - `planGraphBuilder` interface has `getRootPrimitiveGenerator()`

5. **OID fidelity + value coercion** (IN PROGRESS — 10 failures remaining):
   - Finer OID mapping in `typing/standard_column_metadata.go`:
     - `getOidForSchema`: `integer`→`T_int8`, `boolean`→`T_bool`, `number`→`T_numeric`
     - `getOidForParserColType`: split into `T_int2/T_int4/T_int8/T_float4/T_float8/T_json/T_jsonb`
   - Finer OID mapping in `typing/relayed_column_metadata.go`
   - Value coercion in `internal/stackql/psqlwire/psqlwire.go`:
     - `coerceForOID()` function converts string/[]byte from RDBMS to Go types pgtype expects
     - Applied in `ExtractRowElement` for non-text, non-numeric OIDs
     - Existing `shimNumericElement` preserved for `"numeric"` type
   - **397/407 pass, 10 failures remain** — need to check what those 10 are

### What's next (from the plan)

**Immediate**: Fix remaining 10 test failures from OID changes. Check what OID/coercion path they hit.

**Phase 2**: `paramresolver` package
- Resolve `$N` placeholder OIDs from method schemas during `HandleParse`
- Add `ParameterOIDs []uint32` to `plan.Plan`
- Populate during plan building in `entrypoint.go`

**Phase 3**: Stateful driver + `paramdecoder` package
- `paramdecoder`: decode binary-format params using `jackc/pgtype`
- Statement/portal caches in `basicStackQLDriver`:
  - `HandleParse`: resolve OIDs + infer columns → cache in `stmtCache`
  - `HandleDescribeStatement`: return from cache (no re-planning)
  - `HandleBind`: record portal→statement mapping
  - `HandleExecute`: look up portal → decode params → substitute → execute
  - `HandleClose*`: delete from caches

**Phase 4** (separate, psql-wire repo): Respect `resultFormats` from Bind instead of hardcoding `TextFormat`

### Key files modified

| File | Status | Description |
|------|--------|-------------|
| `internal/stackql/driver/driver.go` | Modified | IExtendedQueryBackend impl, shapeInferrer field |
| `internal/stackql/queryshape/queryshape.go` | New | Inferrer interface, SubstituteParams |
| `internal/stackql/queryshape/queryshape_test.go` | New | Unit tests with JSON testdata |
| `internal/stackql/queryshape/testdata/*.json` | New | Test cases |
| `internal/stackql/psqlwire/psqlwire.go` | Modified | coerceForOID() value coercion |
| `internal/stackql/typing/standard_column_metadata.go` | Modified | Finer OID mapping |
| `internal/stackql/typing/relayed_column_metadata.go` | Modified | Finer OID mapping |
| `internal/stackql/typing/oid_mapping_test.go` | New | OID mapping unit tests |
| `internal/stackql/plan/plan.go` | Modified | ColumnMetadata field + getter/setter |
| `internal/stackql/planbuilder/entrypoint.go` | Modified | Extract column metadata during plan build |
| `internal/stackql/planbuilder/plan_builder.go` | Modified | getRootPrimitiveGenerator() on interface |
| `test/robot/functional/stackql_mocked_from_cmd_line.robot` | Modified | 8 new indirect join tests + 3-way test |
| `docs/views.md` | Modified | Updated supported join combinations |

### Key architectural decisions

- `queryshape.Inferrer` is the single entry point for ahead-of-time schema inference
- Plan building (without execution) is the mechanism for inferring column types from provider schemas
- Value coercion in `psqlwire/psqlwire.go` bridges sqlite's string-heavy output to pgtype's typed encoders
- OID fidelity is now enabled: integer→T_int8, boolean→T_bool, fine-grained parser col types
- The `shimNumericElement`/`shimNumericTextBytes` hacks are preserved for backward compatibility with the "numeric" pgtype path

### Phase 3 Complete (2026-04-04)

All Handle* methods now flow through stateful caches:
- `HandleParse`: infers columns, caches in `stmtCache[stmtName]`
- `HandleDescribeStatement`: returns from cache (no re-planning)
- `HandleDescribePortal`: looks up portal→statement→columns
- `HandleBind`: records portal→statement in `portalCache`
- `HandleExecute`: looks up portal→OIDs, decodes params via `paramdecoder`, substitutes, executes
- `HandleClose*`: cleans up caches

New packages:
- `internal/stackql/paramdecoder/` — decodes text AND binary format params (int2/4/8, float4/8, bool, timestamp, text)
- Value coercion function `coerceForOID` in `psqlwire.go` — ready but not active (deferred to Phase 4 with OID fidelity)

### Remaining work

- **Phase 1 (OID fidelity)**: finer OIDs (integer→T_int8, bool→T_bool) break 10 tests because pgtype's text encoder formats values differently. Needs psql-wire change to bypass pgtype.Set() for text format and write strings directly. Deferred.
- **Phase 2 (paramresolver)**: resolve $N placeholder OIDs from method schemas. pgx works without this (defaults to text). Enhancement.
- **Phase 4 (psql-wire)**: respect resultFormats from Bind, enable binary result encoding. Separate repo.
