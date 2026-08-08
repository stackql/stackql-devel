// Package intrinsic implements the built-in "stackql_intrinsic" provider.
//
// Meta statements (USE, SHOW, DESCRIBE) are answered from static, in-process
// data; they never touch the provider registry, the anysdk hierarchy resolver
// or any HTTP machinery. The plan builder consults GeneratePrimitiveFunc ahead
// of its normal statement dispatch.
//
// Relations come in two flavours, both under the "audit" service. View
// relations ("info", "catalog", "methods") materialise as SQL literals and so
// compose anywhere a relation is legal; SELECT over them is left to the normal
// relational path. Data relations - one per omnisdk resource - stream the
// product of a method plan straight to the output writer, and SELECT over
// those IS handled here. See omnisdk.go for the trade-off that entails.
package intrinsic

import (
	"fmt"
	"strings"

	"github.com/stackql/any-sdk/pkg/dto"
	"github.com/stackql/any-sdk/public/formulation"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/typing"
	"github.com/stackql/stackql/internal/stackql/util"

	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

const (
	// ProviderName is the sole provider routed to intrinsic plan generation.
	ProviderName = "stackql_intrinsic"
	// ProviderVersion is reported by SHOW PROVIDERS.
	ProviderVersion = "internal"
)

const (
	serviceName      = "audit"
	serviceTitle     = "Intrinsic Audit"
	selectMethodName = "select"
	// columnType is the type reported by DESCRIBE for a column whose schema
	// declares none; view relations are all SQL literals, hence text.
	columnType = "text"
)

// column is one column of an intrinsic relation.
type column struct {
	name        string
	description string
	// dataType is the JSON Schema type name ("string", "boolean", "integer",
	// "number"). Empty means text, which is what every view relation is.
	dataType string
}

// table is one intrinsic relation, in exactly one of two flavours.
//
// rows != nil: a view relation. The values are rendered as SQL literals at
// registration time and materialise as a view, so the relation composes with
// arbitrary SQL.
//
// isData: a data relation. Rows are produced at query time by running an
// omnisdk method plan and are pushed straight to the output writer; see
// omnisdk.go for what that costs.
type table struct {
	name        string
	description string
	columns     []column
	rows        func() ([][]string, error)
	isData      bool
}

// tables is the intrinsic relation registry. Adding a relation is a matter of
// adding an entry: SELECT, SHOW RESOURCES, SHOW METHODS and DESCRIBE all read
// from here.
var tables = []table{ //nolint:gochecknoglobals // compile-time relation registry
	{
		name:        "info",
		description: "built-in intrinsic info resource",
		columns: []column{
			{name: "title", description: "intrinsic info title"},
			{name: "description", description: "intrinsic info description"},
		},
		rows: infoRows,
	},
	{
		name:        "catalog",
		description: "omnisdk resource catalog",
		columns: []column{
			{name: "path", description: "dot-path addressing the omnisdk resource"},
			{name: "relation", description: "stackql relation name for the resource"},
			{name: "summary", description: "human readable summary of the resource"},
		},
		rows: catalogRows,
	},
	{
		name:        "methods",
		description: "omnisdk method catalog",
		columns: []column{
			{name: "path", description: "dot-path addressing the omnisdk method"},
			{name: "relation", description: "stackql relation the method belongs to"},
			{name: "summary", description: "human readable summary of the method"},
			{name: "required_params", description: "comma separated required parameter names"},
		},
		rows: methodRows,
	},
}

// allTables is the view relations plus the omnisdk data relations. SHOW
// RESOURCES, SHOW METHODS and DESCRIBE all read from here.
func allTables() ([]table, error) {
	data, err := dataTables()
	if err != nil {
		return nil, err
	}
	return append(append([]table{}, tables...), data...), nil
}

// queryContext is the narrow slice of the handler context that intrinsic
// planning needs. It is declared here, rather than imported, so that this
// package remains a leaf and `handler` may import it without a cycle.
// handler.HandlerContext satisfies it structurally.
type queryContext interface {
	GetCurrentProvider() string
	SetCurrentProvider(string)
	GetTypingConfig() typing.Config
	GetAuthContext(providerName string) (*dto.AuthCtx, error)
}

// relationRegistrar is the narrow slice of sql_system.SQLSystem needed to
// register the intrinsic relations. As with queryContext, it is declared here
// to keep this package a leaf.
type relationRegistrar interface {
	CreateView(viewName string, rawDDL string, replaceAllowed bool, requiredParams []string) error
}

func infoRows() ([][]string, error) {
	return [][]string{
		{"intrinsic", "placeholder audit info row"},
		{"placeholder", "second placeholder audit info row"},
	}, nil
}

// RegisterRelations registers the intrinsic relations with the SQL backend. It
// is idempotent and is invoked once per handler context.
func RegisterRelations(registrar relationRegistrar) error {
	for _, tbl := range tables {
		if tbl.rows == nil {
			continue
		}
		ddl, err := tbl.ddl()
		if err != nil {
			return err
		}
		if err = registrar.CreateView(tbl.qualifiedName(), ddl, true, nil); err != nil {
			return err
		}
	}
	return nil
}

func (t table) qualifiedName() string {
	return fmt.Sprintf("%s.%s.%s", ProviderName, serviceName, t.name)
}

// ddl renders the relation as a UNION ALL of literal SELECTs. Only the first
// leg names the columns, which is what fixes the relation's column order.
func (t table) ddl() (string, error) {
	rows, err := t.rows()
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return t.emptyDDL(), nil
	}
	var builder strings.Builder
	for i, row := range rows {
		if len(row) != len(t.columns) {
			return "", fmt.Errorf(
				"intrinsic: relation '%s' row %d has %d values, want %d",
				t.name, i, len(row), len(t.columns))
		}
		if i > 0 {
			builder.WriteString(" UNION ALL ")
		}
		builder.WriteString("SELECT ")
		for j, value := range row {
			if j > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(sqlLiteral(value))
			if i == 0 {
				builder.WriteString(" AS " + t.columns[j].name)
			}
		}
	}
	return builder.String(), nil
}

// emptyDDL yields a correctly shaped relation with no rows.
func (t table) emptyDDL() string {
	projections := make([]string, 0, len(t.columns))
	for _, col := range t.columns {
		projections = append(projections, fmt.Sprintf("CAST(NULL AS TEXT) AS %s", col.name))
	}
	return "SELECT " + strings.Join(projections, ", ") + " WHERE 1 = 0"
}

// sqlLiteral renders s as a SQL string literal, doubling embedded quotes.
func sqlLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// GeneratePrimitiveFunc returns an executor for stmt, and true, when stmt is
// routed to the intrinsic provider. It returns false for every other
// statement, leaving the caller's normal dispatch untouched.
func GeneratePrimitiveFunc(
	ctx queryContext,
	stmt sqlparser.SQLNode,
) (func() internaldto.ExecutorOutput, bool) {
	current := ctx.GetCurrentProvider()
	switch node := stmt.(type) {
	case *sqlparser.Use:
		return useFunc(ctx, node)
	case *sqlparser.Show:
		return showFunc(ctx, node, current)
	case *sqlparser.DescribeTable:
		return describeTableFunc(ctx, node, current)
	case *sqlparser.DescribeMethod:
		return describeMethodFunc(ctx, node, current)
	}
	return nil, false
}

// GenerateStreamFunc returns a streaming executor for a SELECT over an omnisdk
// data relation. Such a relation is not backed by a view, so it must be planned
// before analysis tries to resolve it as a registry-backed provider relation;
// callers invoke this ahead of that analysis. Every other SELECT returns false.
func GenerateStreamFunc(
	ctx queryContext,
	node *sqlparser.Select,
) (func() internaldto.ExecutorOutput, bool) {
	return selectFunc(ctx, node, ctx.GetCurrentProvider())
}

// IsProvider reports whether name designates the intrinsic provider. The
// intrinsic provider has no registry document and no auth, so callers that
// eagerly resolve provider strings must skip it.
func IsProvider(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), ProviderName)
}

// resolveProvider prefers an explicitly qualified provider, falling back to the
// session provider set by USE.
func resolveProvider(providerName string, currentProvider string) string {
	if strings.TrimSpace(providerName) != "" {
		return providerName
	}
	return currentProvider
}

func isAuditService(providerName, serviceStr, currentProvider string) bool {
	return IsProvider(resolveProvider(providerName, currentProvider)) &&
		strings.EqualFold(serviceStr, serviceName)
}

// lookupTable resolves a fully addressed intrinsic relation.
func lookupTable(providerName, serviceStr, resourceStr, currentProvider string) (table, bool) {
	if !isAuditService(providerName, serviceStr, currentProvider) {
		return table{}, false
	}
	registry, err := allTables()
	if err != nil {
		return table{}, false
	}
	for _, tbl := range registry {
		if strings.EqualFold(resourceStr, tbl.name) {
			return tbl, true
		}
	}
	return table{}, false
}

func isExtended(extended string) bool {
	return strings.EqualFold(strings.TrimSpace(extended), "extended")
}

func useFunc(ctx queryContext, node *sqlparser.Use) (func() internaldto.ExecutorOutput, bool) {
	if !IsProvider(node.DBName.GetRawVal()) {
		return nil, false
	}
	return func() internaldto.ExecutorOutput {
		ctx.SetCurrentProvider(ProviderName)
		return internaldto.NewExecutorOutput(nil, nil, nil, nil, nil)
	}, true
}

//nolint:gocritic // a switch on node.Type reads better than the alternatives
func showFunc(
	ctx queryContext,
	node *sqlparser.Show,
	currentProvider string,
) (func() internaldto.ExecutorOutput, bool) {
	extended := isExtended(node.Extended)
	switch strings.ToUpper(strings.TrimSpace(node.Type)) {
	case "SERVICES":
		// SHOW SERVICES IN <provider>
		if !IsProvider(resolveProvider(node.OnTable.Name.GetRawVal(), currentProvider)) {
			return nil, false
		}
		return func() internaldto.ExecutorOutput { return showServices(ctx, extended) }, true
	case "RESOURCES":
		// SHOW RESOURCES IN <provider>.<service>
		if !isAuditService(
			node.OnTable.Qualifier.GetRawVal(), node.OnTable.Name.GetRawVal(), currentProvider) {
			return nil, false
		}
		return func() internaldto.ExecutorOutput { return showResources(ctx, extended) }, true
	case "METHODS":
		// SHOW METHODS IN <provider>.<service>.<resource>
		tbl, ok := lookupTable(
			node.OnTable.QualifierSecond.GetRawVal(),
			node.OnTable.Qualifier.GetRawVal(),
			node.OnTable.Name.GetRawVal(),
			currentProvider,
		)
		if !ok {
			return nil, false
		}
		return func() internaldto.ExecutorOutput { return showMethods(ctx, tbl, extended) }, true
	}
	return nil, false
}

func describeTableFunc(
	ctx queryContext,
	node *sqlparser.DescribeTable,
	currentProvider string,
) (func() internaldto.ExecutorOutput, bool) {
	tbl, ok := lookupTable(
		node.Table.QualifierSecond.GetRawVal(),
		node.Table.Qualifier.GetRawVal(),
		node.Table.Name.GetRawVal(),
		currentProvider,
	)
	if !ok {
		return nil, false
	}
	extended := isExtended(node.Extended)
	return func() internaldto.ExecutorOutput { return describeTable(ctx, tbl, extended) }, true
}

func describeMethodFunc(
	ctx queryContext,
	node *sqlparser.DescribeMethod,
	currentProvider string,
) (func() internaldto.ExecutorOutput, bool) {
	tbl, ok := lookupTable(
		node.Provider.GetRawVal(),
		node.Service.GetRawVal(),
		node.Resource.GetRawVal(),
		currentProvider,
	)
	if !ok {
		return nil, false
	}
	methodName := node.Method.GetRawVal()
	// A data relation describes the named omnisdk method, whose response shape
	// is its own; a view relation has only the synthetic "select".
	columns, ok := tbl.methodColumns(methodName)
	if !ok {
		return func() internaldto.ExecutorOutput {
			return internaldto.NewErroneousExecutorOutput(
				fmt.Errorf(
					"relation '%s.%s.%s' has no method '%s'; run SHOW METHODS to list them",
					ProviderName, serviceName, tbl.name, methodName,
				),
			)
		}, true
	}
	extended := isExtended(node.Extended)
	return func() internaldto.ExecutorOutput { return describeMethod(ctx, columns, extended) }, true
}

func prepare(
	ctx queryContext,
	columnOrder []string,
	rows map[string]map[string]interface{},
	rowSort func(map[string]map[string]interface{}) []string,
) internaldto.ExecutorOutput {
	return util.PrepareResultSet(
		internaldto.NewPrepareResultSetDTO(
			nil, rows, columnOrder, rowSort, nil, nil, ctx.GetTypingConfig(),
		),
	)
}

func showServices(ctx queryContext, extended bool) internaldto.ExecutorOutput {
	row := map[string]interface{}{
		"id":    serviceName,
		"name":  serviceName,
		"title": serviceTitle,
	}
	if extended {
		row["description"] = "built-in intrinsic service"
		row["version"] = ProviderVersion
		row["preferred"] = true
	}
	return prepare(
		ctx,
		formulation.GetServicesHeader(extended),
		map[string]map[string]interface{}{"000001": row},
		util.DefaultRowSort,
	)
}

func showResources(ctx queryContext, extended bool) internaldto.ExecutorOutput {
	registry, err := allTables()
	if err != nil {
		return internaldto.NewErroneousExecutorOutput(err)
	}
	rows := make(map[string]map[string]interface{}, len(registry))
	for i, tbl := range registry {
		row := map[string]interface{}{
			"id":   tbl.name,
			"name": tbl.name,
		}
		if extended {
			row["description"] = tbl.description
		}
		rows[fmt.Sprintf("%06d", i)] = row
	}
	return prepare(ctx, formulation.GetResourcesHeader(extended), rows, util.DefaultRowSort)
}

// showMethods lists a relation's methods. A view relation has the one
// synthetic "select"; a data relation reports the omnisdk methods it can run,
// with their real required parameters, since those are what a caller must
// supply as WHERE predicates.
func showMethods(ctx queryContext, tbl table, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"MethodName", "RequiredParams", "SQLVerb"}
	if extended {
		columnOrder = append(columnOrder, "description")
	}
	rows := make(map[string]map[string]interface{})
	for i, meth := range tbl.methods() {
		row := map[string]interface{}{
			"MethodName":     meth.name,
			"RequiredParams": strings.Join(meth.requiredParams, ", "),
			"SQLVerb":        strings.ToUpper(selectMethodName),
		}
		if extended {
			row["description"] = meth.description
		}
		rows[fmt.Sprintf("%06d", i)] = row
	}
	return prepare(ctx, columnOrder, rows, util.DefaultRowSort)
}

func describeTable(ctx queryContext, tbl table, extended bool) internaldto.ExecutorOutput {
	return prepare(
		ctx,
		formulation.GetDescribeHeader(extended),
		columnRows(tbl.columns, extended, nil),
		util.DescribeRowSort,
	)
}

func describeMethod(ctx queryContext, cols []column, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"name", "type", "param_type", "shape"}
	if extended {
		columnOrder = append(columnOrder, "description")
	}
	rows := columnRows(cols, extended, func(row map[string]interface{}) {
		row["param_type"] = "response"
		row["shape"] = row["type"]
	})
	return prepare(ctx, columnOrder, rows, util.DescribeRowSort)
}

// columnRows renders a relation's column set as description rows, applying
// decorate (when non-nil) to each row, so that DESCRIBE and DESCRIBE METHOD
// share one source of truth for the relation shape.
func columnRows(
	cols []column,
	extended bool,
	decorate func(map[string]interface{}),
) map[string]map[string]interface{} {
	rows := make(map[string]map[string]interface{}, len(cols))
	for i, col := range cols {
		row := map[string]interface{}{
			"name": col.name,
			"type": col.reportedType(),
		}
		if extended {
			row["description"] = col.description
		}
		if decorate != nil {
			decorate(row)
		}
		rows[fmt.Sprintf("%06d", i)] = row
	}
	return rows
}
