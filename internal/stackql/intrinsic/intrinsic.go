// Package intrinsic implements the built-in "stackql_intrinsic" provider.
//
// Statements addressed to this provider are answered from static, in-process
// data; they never touch the provider registry, the anysdk hierarchy resolver
// or any HTTP machinery. The plan builder consults GeneratePrimitiveFunc ahead
// of its normal statement dispatch, so this package is the sole plan generator
// for the intrinsic provider.
//
// The surface is deliberately a placeholder: a single service ("audit") holding
// a single resource ("info") with a select-only method.
package intrinsic

import (
	"fmt"
	"strings"

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
	resourceName     = "info"
	selectMethodName = "select"

	titleColumn       = "title"
	descriptionColumn = "description"
	columnType        = "text"
)

// queryContext is the narrow slice of the handler context that intrinsic
// planning needs. It is declared here, rather than imported, so that this
// package remains a leaf and `handler` may import it without a cycle.
// handler.HandlerContext satisfies it structurally.
type queryContext interface {
	GetCurrentProvider() string
	SetCurrentProvider(string)
	GetTypingConfig() typing.Config
}

// relationRegistrar is the narrow slice of sql_system.SQLSystem needed to
// register the intrinsic relations. As with queryContext, it is declared here
// to keep this package a leaf.
type relationRegistrar interface {
	CreateView(viewName string, rawDDL string, replaceAllowed bool, requiredParams []string) error
}

// infoDDL is the body of the "info" relation. Registering it as a view, rather
// than answering SELECT from a bespoke primitive, is what lets the relation be
// referenced anywhere a relation is legal: subqueries, joins, CTEs, views over
// it, and so on. The SQL backend then applies projection, predicates and
// ordering natively.
const infoDDL = `SELECT 'intrinsic' AS title, ` +
	`'placeholder audit info row' AS description ` +
	`UNION ALL ` +
	`SELECT 'placeholder', 'second placeholder audit info row'`

// RegisterRelations registers the intrinsic relations with the SQL backend. It
// is idempotent and is invoked once per handler context.
func RegisterRelations(registrar relationRegistrar) error {
	return registrar.CreateView(
		fmt.Sprintf("%s.%s.%s", ProviderName, serviceName, resourceName),
		infoDDL,
		true,
		nil,
	)
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
	// SELECT is deliberately absent: the "info" relation is registered as a
	// view (see RegisterRelations) so that it composes with arbitrary SQL.
	return nil, false
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

func isAuditInfo(providerName, serviceStr, resourceStr, currentProvider string) bool {
	return IsProvider(resolveProvider(providerName, currentProvider)) &&
		strings.EqualFold(serviceStr, serviceName) &&
		strings.EqualFold(resourceStr, resourceName)
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
		if !IsProvider(resolveProvider(node.OnTable.Qualifier.GetRawVal(), currentProvider)) ||
			!strings.EqualFold(node.OnTable.Name.GetRawVal(), serviceName) {
			return nil, false
		}
		return func() internaldto.ExecutorOutput { return showResources(ctx, extended) }, true
	case "METHODS":
		// SHOW METHODS IN <provider>.<service>.<resource>
		if !isAuditInfo(
			node.OnTable.QualifierSecond.GetRawVal(),
			node.OnTable.Qualifier.GetRawVal(),
			node.OnTable.Name.GetRawVal(),
			currentProvider,
		) {
			return nil, false
		}
		return func() internaldto.ExecutorOutput { return showMethods(ctx, extended) }, true
	}
	return nil, false
}

func describeTableFunc(
	ctx queryContext,
	node *sqlparser.DescribeTable,
	currentProvider string,
) (func() internaldto.ExecutorOutput, bool) {
	if !isAuditInfo(
		node.Table.QualifierSecond.GetRawVal(),
		node.Table.Qualifier.GetRawVal(),
		node.Table.Name.GetRawVal(),
		currentProvider,
	) {
		return nil, false
	}
	extended := isExtended(node.Extended)
	return func() internaldto.ExecutorOutput { return describeInfo(ctx, extended) }, true
}

func describeMethodFunc(
	ctx queryContext,
	node *sqlparser.DescribeMethod,
	currentProvider string,
) (func() internaldto.ExecutorOutput, bool) {
	if !isAuditInfo(
		node.Provider.GetRawVal(),
		node.Service.GetRawVal(),
		node.Resource.GetRawVal(),
		currentProvider,
	) {
		return nil, false
	}
	methodName := node.Method.GetRawVal()
	if !strings.EqualFold(methodName, selectMethodName) {
		return func() internaldto.ExecutorOutput {
			return internaldto.NewErroneousExecutorOutput(
				fmt.Errorf(
					"provider '%s' supports method '%s' only; got '%s'",
					ProviderName, selectMethodName, methodName,
				),
			)
		}, true
	}
	extended := isExtended(node.Extended)
	return func() internaldto.ExecutorOutput { return describeMethod(ctx, extended) }, true
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
	row := map[string]interface{}{
		"id":   resourceName,
		"name": resourceName,
	}
	if extended {
		row["description"] = "built-in intrinsic info resource"
	}
	return prepare(
		ctx,
		formulation.GetResourcesHeader(extended),
		map[string]map[string]interface{}{"000001": row},
		util.DefaultRowSort,
	)
}

func showMethods(ctx queryContext, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"MethodName", "RequiredParams", "SQLVerb"}
	row := map[string]interface{}{
		"MethodName":     selectMethodName,
		"RequiredParams": "",
		"SQLVerb":        strings.ToUpper(selectMethodName),
	}
	if extended {
		columnOrder = append(columnOrder, "description")
		row["description"] = "select-only intrinsic method"
	}
	return prepare(ctx, columnOrder, map[string]map[string]interface{}{"000001": row}, util.DefaultRowSort)
}

func describeInfo(ctx queryContext, extended bool) internaldto.ExecutorOutput {
	return prepare(
		ctx,
		formulation.GetDescribeHeader(extended),
		columnRows(extended, nil),
		util.DescribeRowSort,
	)
}

func describeMethod(ctx queryContext, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"name", "type", "param_type", "shape"}
	if extended {
		columnOrder = append(columnOrder, "description")
	}
	rows := columnRows(extended, func(row map[string]interface{}) {
		row["param_type"] = "response"
		row["shape"] = columnType
	})
	return prepare(ctx, columnOrder, rows, util.DescribeRowSort)
}

// columnRows renders the "info" column set as description rows, applying
// decorate (when non-nil) to each row, so that DESCRIBE and DESCRIBE METHOD
// share one source of truth for the table shape.
func columnRows(
	extended bool,
	decorate func(map[string]interface{}),
) map[string]map[string]interface{} {
	descriptions := map[string]string{
		titleColumn:       "intrinsic info title",
		descriptionColumn: "intrinsic info description",
	}
	columnNames := []string{titleColumn, descriptionColumn}
	rows := make(map[string]map[string]interface{}, len(columnNames))
	for i, columnName := range columnNames {
		row := map[string]interface{}{
			"name": columnName,
			"type": columnType,
		}
		if extended {
			row["description"] = descriptions[columnName]
		}
		if decorate != nil {
			decorate(row)
		}
		rows[fmt.Sprintf("%06d", i)] = row
	}
	return rows
}
