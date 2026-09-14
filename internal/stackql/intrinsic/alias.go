package intrinsic

// The alias providers present omnisdk's document-driven capabilities as
// relations: stackql_dynamic for a query spanning several exchanges, stackql_iac
// for an idempotent converge run. Both are deliberately thin. A predicate
// carries the specification, the SDK does the work, and the rows it yields
// stream through the same cursor a single-method select already uses.
//
// They are gated behind the same opt-in as the document-driven providers,
// because that is what they read: documents straight from disk, with none of
// the registry's curation behind them.

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/stackql-labs/omnisdk/pkg/omnisdk"
	"github.com/stackql/any-sdk/public/formulation"
	"github.com/stackql/psql-wire/pkg/sqldata"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/util"

	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

const (
	// DynamicProviderName addresses a query wired across several exchanges.
	DynamicProviderName = "stackql_dynamic"
	// IaCProviderName addresses an idempotent converge run and the blueprints
	// that render one.
	IaCProviderName = "stackql_iac"
)

const (
	graphService     = "graph"
	graphRelation    = "query"
	convergeService  = "converge"
	convergeRelation = "run"
	blueprintService = "blueprints"
)

const (
	specPredicate       = "spec"
	collectionPredicate = "collection"
	blueprintPredicate  = "blueprint"
	resourcesPredicate  = "resources"
	statePredicate      = "state"
	runIDPredicate      = "run_id"
)

// aliasProvider canonicalises an alias provider name, or reports false. The
// aliases read documents from disk, so they appear only once that was opted
// into, exactly as the unstable providers do.
func aliasProvider(name string) (string, bool) {
	if !IsUnstableEnabled() {
		return "", false
	}
	switch trimmed := strings.TrimSpace(name); {
	case strings.EqualFold(trimmed, DynamicProviderName):
		return DynamicProviderName, true
	case strings.EqualFold(trimmed, IaCProviderName):
		return IaCProviderName, true
	default:
		return "", false
	}
}

// aliasServices names the services an alias provider presents.
func aliasServices(provider string) []string {
	if provider == IaCProviderName {
		return []string{blueprintService, convergeService}
	}
	return []string{graphService}
}

// registryRoot is the provider-document root omnisdk resolves addresses
// against. It is the directory holding "<provider>/<version>/provider.yaml",
// which is where stackql's own document root keeps them.
func registryRoot(ctx queryContext) string {
	return filepath.Join(localDocRoot(ctx.GetRuntimeContext()), "src")
}

// popPredicate takes a predicate out of the parameter map. What remains after
// every control predicate is popped is the run's scope, which rides through to
// omnisdk untouched.
func popPredicate(params map[string]string, key string) string {
	value := params[key]
	delete(params, key)
	return value
}

// streamPlan opens a plan and hands its cursor to the caller. Neither alias
// declares an egress schema, so the columns are the ones the first row carries
// and the projection is applied over them.
func streamPlan(
	ctx queryContext,
	plan omnisdk.Plan,
	relation string,
	exprs sqlparser.SelectExprs,
) internaldto.ExecutorOutput {
	rows, openErr := plan.Open(context.Background())
	if openErr != nil {
		return internaldto.NewErroneousExecutorOutput(openErr)
	}
	input := previewCfg
	stream := &rowStream{
		rows:          rows,
		batchSize:     input.getBatchSize(),
		flushInterval: input.getFlushInterval(),
		table:         sqldata.NewSQLTable(0, relation),
		typCfg:        ctx.GetTypingConfig(),
		projection:    exprs,
	}
	primed, readErr := newPrimedStream(stream)
	if readErr != nil {
		return internaldto.NewErroneousExecutorOutput(readErr)
	}
	return internaldto.NewExecutorOutput(primed, nil, nil, nil, nil)
}

// aliasArgs assembles the SDK arguments common to both aliases: the scope left
// over after the control predicates, the credential for the cloud in play, and
// the backend tuning.
func aliasArgs(ctx queryContext, cloud string, params map[string]string) omnisdk.Args {
	input := previewCfg
	return omnisdk.Args{
		Params:                params,
		Auth:                  omnisdkAuth(providerAuthContext(ctx, cloud)),
		Endpoint:              input.getEndpoint(),
		InsecureSkipTLSVerify: input.getInsecureSkipTLSVerify(),
	}
}

// aliasSelectFunc routes a SELECT over an alias relation.
func aliasSelectFunc(
	ctx queryContext,
	node *sqlparser.Select,
	provider, service, resource string,
) (func() internaldto.ExecutorOutput, bool) {
	if provider == DynamicProviderName {
		return dynamicSelectFunc(ctx, node, service, resource)
	}
	return iacSelectFunc(ctx, node, service, resource)
}

// aliasPredicates reads a select's WHERE clause as the flat equality map both
// aliases take, refusing anything the streaming path cannot honour.
func aliasPredicates(
	node *sqlparser.Select,
	relation string,
) (map[string]string, func() internaldto.ExecutorOutput) {
	if unsupported := unsupportedClauses(node); len(unsupported) > 0 {
		return nil, refuse(fmt.Errorf(
			"relation '%s' streams its rows, so %s cannot be applied; remove %s from the query",
			relation, strings.Join(unsupported, ", "), pluralClause(len(unsupported))))
	}
	params, bad := equalityPredicates(node.Where)
	if len(bad) > 0 {
		return nil, refuse(fmt.Errorf(
			"relation '%s' streams its rows, so only equality predicates are applied; "+
				"%s cannot be honoured",
			relation, strings.Join(bad, ", ")))
	}
	return params, nil
}

// showAliasFunc answers SHOW for the alias providers.
func showAliasFunc(
	ctx queryContext,
	node *sqlparser.Show,
	currentProvider string,
	extended bool,
) (func() internaldto.ExecutorOutput, bool) {
	switch strings.ToUpper(strings.TrimSpace(node.Type)) {
	case "SERVICES":
		provider, isAlias := aliasProvider(resolveProvider(node.OnTable.Name.GetRawVal(), currentProvider))
		if !isAlias {
			return nil, false
		}
		return func() internaldto.ExecutorOutput {
			return showAliasServices(ctx, provider, extended)
		}, true
	case "RESOURCES":
		provider, isAlias := aliasProvider(
			resolveProvider(node.OnTable.Qualifier.GetRawVal(), currentProvider))
		if !isAlias {
			return nil, false
		}
		service := node.OnTable.Name.GetRawVal()
		return func() internaldto.ExecutorOutput {
			return showAliasResources(ctx, provider, service, extended)
		}, true
	case "METHODS":
		// Every alias relation is select-only, so the method list does not vary
		// by service or resource.
		if _, isAlias := aliasProvider(
			resolveProvider(node.OnTable.QualifierSecond.GetRawVal(), currentProvider)); !isAlias {
			return nil, false
		}
		return func() internaldto.ExecutorOutput { return showAliasMethods(ctx, extended) }, true
	}
	return nil, false
}

func showAliasServices(ctx queryContext, provider string, extended bool) internaldto.ExecutorOutput {
	services := aliasServices(provider)
	rows := make(map[string]map[string]interface{}, len(services))
	for i, service := range services {
		row := map[string]interface{}{
			"id":    fmt.Sprintf("%s.%s", provider, service),
			"name":  service,
			"title": service,
		}
		if extended {
			row["description"] = aliasServiceDescription(provider, service)
			row["version"] = ProviderVersion
			row["preferred"] = nil
		}
		rows[fmt.Sprintf("%06d", i)] = row
	}
	return prepare(ctx, formulation.GetServicesHeader(extended), rows, util.DefaultRowSort)
}

func aliasServiceDescription(provider, service string) string {
	switch {
	case provider == DynamicProviderName:
		return "a query wired across several document-declared exchanges"
	case service == blueprintService:
		return "precanned deployments, and the inputs each declares"
	default:
		return "an idempotent converge run over a collection of resources"
	}
}

func showAliasResources(
	ctx queryContext, provider, service string, extended bool) internaldto.ExecutorOutput {
	tables, err := aliasTables(provider, service)
	if err != nil {
		return internaldto.NewErroneousExecutorOutput(err)
	}
	rows := make(map[string]map[string]interface{}, len(tables))
	for i, tbl := range tables {
		row := map[string]interface{}{
			"id":   fmt.Sprintf("%s.%s.%s", provider, tbl.service, tbl.name),
			"name": tbl.name,
		}
		if extended {
			row["description"] = tbl.description
		}
		rows[fmt.Sprintf("%06d", i)] = row
	}
	return prepare(ctx, formulation.GetResourcesHeader(extended), rows, util.DefaultRowSort)
}

func showAliasMethods(ctx queryContext, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"MethodName", "RequiredParams", "SQLVerb"}
	if extended {
		columnOrder = append(columnOrder, "description")
	}
	row := map[string]interface{}{
		"MethodName":     selectMethodName,
		"RequiredParams": "",
		"SQLVerb":        strings.ToUpper(selectMethodName),
	}
	if extended {
		row["description"] = "select-only intrinsic method"
	}
	return prepare(ctx, columnOrder,
		map[string]map[string]interface{}{"000001": row}, util.DefaultRowSort)
}

// aliasTables presents an alias service's relations.
func aliasTables(provider, service string) ([]table, error) {
	switch {
	case provider == DynamicProviderName && strings.EqualFold(service, graphService):
		return []table{{
			service:     graphService,
			name:        graphRelation,
			isData:      true,
			description: "a graph query; the '" + specPredicate + "' predicate carries its wiring",
		}}, nil
	case provider == IaCProviderName && strings.EqualFold(service, convergeService):
		return []table{{
			service:     convergeService,
			name:        convergeRelation,
			isData:      true,
			description: "an idempotent converge run over a named collection",
		}}, nil
	case provider == IaCProviderName && strings.EqualFold(service, blueprintService):
		return blueprintTables(), nil
	}
	return nil, fmt.Errorf("provider '%s' has no service '%s'", provider, service)
}

// describeAliasTableFunc answers DESCRIBE for the alias providers. Only a
// blueprint has columns worth describing: its relations take a specification
// rather than a column list, and yield whatever the run reports.
func describeAliasTableFunc(
	ctx queryContext,
	node *sqlparser.DescribeTable,
	currentProvider string,
) (func() internaldto.ExecutorOutput, bool) {
	provider, isAlias := aliasProvider(
		resolveProvider(node.Table.QualifierSecond.GetRawVal(), currentProvider))
	if !isAlias {
		return nil, false
	}
	service := node.Table.Qualifier.GetRawVal()
	resource := node.Table.Name.GetRawVal()
	extended := isExtended(node.Extended)
	if provider == IaCProviderName && strings.EqualFold(service, blueprintService) {
		blueprint, ok := blueprintFor(resource)
		if !ok {
			return refuse(fmt.Errorf(
				"'%s.%s' has no blueprint '%s'; run SHOW RESOURCES IN %s.%s to list them",
				IaCProviderName, blueprintService, resource, IaCProviderName, blueprintService)), true
		}
		return func() internaldto.ExecutorOutput {
			return describeTable(ctx, table{columns: blueprintColumns(blueprint)}, extended)
		}, true
	}
	return refuse(fmt.Errorf(
		"relation '%s.%s.%s' takes a specification rather than columns; "+
			"run SHOW RESOURCES IN %s.%s for what it accepts",
		provider, service, resource, provider, service)), true
}
