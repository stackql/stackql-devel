# Window Functions and CTE Implementation for StackQL Parser

## Overview

This document describes the experimental implementation of SQL window functions and CTEs (Common Table Expressions) in the stackql-parser fork. This is a proof-of-concept implementation to validate the approach before implementing it properly in the main stackql-parser repository.

## Summary of Changes

### Files Modified

The following files in `internal/stackql-parser-fork/go/vt/sqlparser/` were modified:

1. **ast.go** - Added AST types for window functions and CTEs
2. **sql.y** - Added grammar rules for parsing
3. **token.go** - Added keyword mappings
4. **constants.go** - Added constants for frame types
5. **external_visitor.go** - Added Accept methods for new types

### New Test Files Created

- `window_test.go` - Unit tests for window function parsing
- `cte_test.go` - Unit tests for CTE parsing

## Implementation Details

### Window Functions

#### AST Types Added (ast.go)

```go
// OverClause represents an OVER clause for window functions
OverClause struct {
    WindowName ColIdent
    WindowSpec *WindowSpec
}

// WindowSpec represents a window specification
WindowSpec struct {
    PartitionBy Exprs
    OrderBy     OrderBy
    Frame       *FrameClause
}

// FrameClause represents a frame clause (ROWS/RANGE)
FrameClause struct {
    Unit  string // ROWS or RANGE
    Start *FramePoint
    End   *FramePoint
}

// FramePoint represents a frame boundary
FramePoint struct {
    Type string // UNBOUNDED PRECEDING, CURRENT ROW, etc.
    Expr Expr   // for N PRECEDING or N FOLLOWING
}
```

The `FuncExpr` struct was extended with an `Over *OverClause` field.

#### Grammar Rules Added (sql.y)

- `over_clause_opt` - Optional OVER clause after function calls
- `window_spec` - Window specification (PARTITION BY, ORDER BY, frame)
- `partition_by_opt` - Optional PARTITION BY clause
- `frame_clause_opt` - Optional frame specification (ROWS/RANGE)
- `frame_point` - Frame boundary points

#### Tokens Added

- `OVER`, `ROWS`, `RANGE`, `UNBOUNDED`, `PRECEDING`, `FOLLOWING`, `CURRENT`, `ROW`

#### Supported Syntax

```sql
-- Simple window function
SELECT SUM(count) OVER () FROM t

-- With ORDER BY
SELECT RANK() OVER (ORDER BY count DESC) FROM t

-- With PARTITION BY
SELECT SUM(count) OVER (PARTITION BY category) FROM t

-- With PARTITION BY and ORDER BY
SELECT SUM(count) OVER (PARTITION BY category ORDER BY name) FROM t

-- With frame clause
SELECT SUM(count) OVER (ORDER BY id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) FROM t

-- Multiple window functions
SELECT SUM(x) OVER (), COUNT(*) OVER (ORDER BY y) FROM t

-- Window functions with aggregates
SELECT serviceName, COUNT(*) as count, SUM(COUNT(*)) OVER () as total FROM t GROUP BY serviceName
```

### CTEs (Common Table Expressions)

#### AST Types Added (ast.go)

```go
// With represents a WITH clause (CTE)
With struct {
    Recursive bool
    CTEs      []*CommonTableExpr
}

// CommonTableExpr represents a single CTE definition
CommonTableExpr struct {
    Name    TableIdent
    Columns Columns
    Subquery *Subquery
}
```

The `Select` struct was extended with a `With *With` field.

#### Grammar Rules Added (sql.y)

- `cte_list` - List of CTE definitions
- `cte` - Single CTE definition
- Extended `base_select` with WITH clause alternatives

#### Tokens Added

- `RECURSIVE`

#### Supported Syntax

```sql
-- Simple CTE
WITH cte AS (SELECT id FROM t) SELECT * FROM cte

-- CTE with column list
WITH cte (col1, col2) AS (SELECT id, name FROM t) SELECT * FROM cte

-- Multiple CTEs
WITH cte1 AS (SELECT id FROM t1), cte2 AS (SELECT id FROM t2) SELECT * FROM cte1 JOIN cte2

-- Recursive CTE
WITH RECURSIVE cte AS (SELECT 1 AS n UNION ALL SELECT n + 1 FROM cte WHERE n < 10) SELECT * FROM cte

-- CTE with window function
WITH sales AS (SELECT product, amount FROM orders)
SELECT product, SUM(amount) OVER (ORDER BY product) FROM sales
```

## Key Design Decisions

### Window Functions

1. **OVER clause placement**: Added `over_clause_opt` to the `function_call_generic` rule to allow OVER on any generic function call.

2. **Frame specification**: Supports both ROWS and RANGE frame types with:
   - UNBOUNDED PRECEDING
   - UNBOUNDED FOLLOWING
   - CURRENT ROW
   - N PRECEDING
   - N FOLLOWING

3. **Named windows**: The grammar supports `OVER window_name` syntax for referencing named windows (though WINDOW clause definition is not yet implemented).

### CTEs

1. **Grammar approach**: Instead of using an optional `with_clause_opt` rule that includes an empty alternative (which caused grammar conflicts), we directly added WITH alternatives to the `base_select` rule.

2. **Recursive CTEs**: Supported via the `WITH RECURSIVE` syntax.

3. **Column lists**: Optional column list specification for CTEs is supported.

## Parser Conflicts

The implementation increases reduce/reduce conflicts from 461 to 464. This is acceptable for an experimental implementation.

## Testing

### Unit Tests

All parser unit tests pass:
- 8 window function tests
- 5 CTE tests

### Running Tests

```bash
cd /home/user/stackql-devel/internal/stackql-parser-fork/go/vt/sqlparser
go test -run "TestWindowFunctions|TestCTEs" -v
```

## Next Steps for Production Implementation

1. **Upstream the changes**: Apply these changes to the main `stackql-parser` repository.

2. **Execution layer**: Implement window function and CTE execution in the SQLite backend:
   - SQLite already supports window functions and CTEs natively
   - Need to ensure the parsed AST is correctly converted to SQL for execution

3. **Named Windows**: Add support for the `WINDOW` clause to define named windows:
   ```sql
   SELECT SUM(x) OVER w FROM t WINDOW w AS (PARTITION BY y ORDER BY z)
   ```

4. **Additional window functions**: The parser supports any function name with OVER. Consider adding specific handling for:
   - ROW_NUMBER()
   - RANK()
   - DENSE_RANK()
   - LEAD()
   - LAG()
   - FIRST_VALUE()
   - LAST_VALUE()
   - NTH_VALUE()
   - NTILE()

5. **Robot tests**: Add integration tests that verify window functions and CTEs work end-to-end with actual cloud provider data.

## Known Limitations

1. **No execution support**: This implementation only adds parsing support. The execution layer still needs to be updated to handle window functions and CTEs.

2. **Pre-existing test failures**: The parser has some pre-existing test failures unrelated to window functions/CTEs (table name quoting, OR operator rendering). These should be addressed separately.

3. **Build complexity**: The local fork approach with replace directive in go.mod can cause directory conflicts when building the main stackql binary. For production, the changes should be upstreamed to stackql-parser.

## Files to Review

- `internal/stackql-parser-fork/go/vt/sqlparser/ast.go` - AST type definitions
- `internal/stackql-parser-fork/go/vt/sqlparser/sql.y` - Grammar rules (lines ~3074-3410)
- `internal/stackql-parser-fork/go/vt/sqlparser/token.go` - Keyword mappings
- `internal/stackql-parser-fork/go/vt/sqlparser/constants.go` - Frame type constants
- `internal/stackql-parser-fork/go/vt/sqlparser/window_test.go` - Window function tests
- `internal/stackql-parser-fork/go/vt/sqlparser/cte_test.go` - CTE tests
