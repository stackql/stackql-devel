package intrinsic

import (
	"slices"
	"strings"

	"github.com/stackql-labs/omnisdk/pkg/omnisdk"
	"github.com/stackql-labs/omnisdk/pkg/query"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

// resolvedStarColumns shares operation selection with streamed execution. The
// host only turns the resulting projection metadata back into SQL AST columns.
func resolvedStarColumns(ctx queryContext, source translatedSource, star *sqlparser.StarExpr,
	relations map[string]omnisdk.Relation) (sqlparser.SelectExprs, error) {
	tables, err := starTables(ctx, source, relations)
	if err != nil {
		return nil, err
	}
	outputs := []query.Output{query.NewOutput("", query.NewStar(star.TableName.Name.GetRawVal()))}
	q, err := query.New(source.joins(), source.where(), outputs)
	if err != nil {
		return nil, err
	}
	q, err = bindServerEnvironment(q, tables)
	if err != nil {
		return nil, err
	}
	dialect, err := backendDialect(ctx)
	if err != nil {
		return nil, err
	}
	resolved, err := omnisdk.ResolveIn(q, tables, dialect.catalogue())
	if err != nil {
		return nil, err
	}
	return starSQLColumns(resolved.Graph(), tables, starBoundInputs(resolved)), nil
}

func starTables(
	ctx queryContext, source translatedSource, relations map[string]omnisdk.Relation,
) (map[string]omnisdk.Table, error) {
	tables := map[string]omnisdk.Table{}
	for _, join := range source.joins() {
		resource := join.Resource()
		if _, function := resource.(query.TableFunction); function {
			continue
		}
		if relation, present := relations[resource.Handle()]; present {
			tables[resource.Alias()] = relation
			continue
		}
		table, err := omnisdk.DescribeTable(registryRoot(ctx), resource.Handle())
		if err != nil {
			return nil, err
		}
		tables[resource.Alias()] = table
	}
	return tables, nil
}

func starBoundInputs(resolved omnisdk.Resolution) map[string]map[string]bool {
	bound := map[string]map[string]bool{}
	for _, node := range resolved.Graph().Nodes() {
		inputs := map[string]bool{}
		for name := range resolved.Params() {
			inputs[name] = true
		}
		for name := range node.Params() {
			inputs[name] = true
		}
		for name := range node.Fanout() {
			inputs[name] = true
		}
		bound[node.Alias()] = inputs
	}
	for _, wiring := range resolved.Graph().Wirings() {
		for _, input := range wiring.Inbound() {
			bound[wiring.To()][input.As()] = true
		}
	}
	return bound
}

func starScopeNames(table omnisdk.Table, inputs map[string]bool) []string {
	if table == nil {
		return nil
	}
	var names []string
	for _, method := range table.Methods() {
		for _, input := range method.Params() {
			if input.In() == "path" && inputs[input.Name()] {
				names = append(names, input.Name())
			}
		}
	}
	return names
}

func starSQLColumns(
	graph omnisdk.Graph, tables map[string]omnisdk.Table, bound map[string]map[string]bool,
) sqlparser.SelectExprs {
	var expressions sqlparser.SelectExprs
	for _, projection := range graph.Projections() {
		var names []string
		for _, column := range projection.Columns() {
			if !strings.HasPrefix(column.Out(), "\x00") {
				names = append(names, column.Out())
			}
		}
		if len(names) == 0 {
			continue
		}
		// Scope columns describe host output; they do not introduce value producers.
		names = append(names, starScopeNames(tables[projection.Alias()], bound[projection.Alias()])...)
		slices.Sort(names)
		for _, name := range slices.Compact(names) {
			expressions = append(expressions, &sqlparser.AliasedExpr{Expr: &sqlparser.ColName{
				Name:      sqlparser.NewColIdent(name),
				Qualifier: sqlparser.TableName{Name: sqlparser.NewTableIdent(projection.Alias())},
			}})
		}
	}
	return expressions
}
