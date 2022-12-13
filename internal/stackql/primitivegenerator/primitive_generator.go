package primitivegenerator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/stackql/stackql/internal/stackql/constants"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internaldto"
	"github.com/stackql/stackql/internal/stackql/iqlutil"
	"github.com/stackql/stackql/internal/stackql/logging"
	"github.com/stackql/stackql/internal/stackql/metadatavisitors"
	"github.com/stackql/stackql/internal/stackql/planbuilderinput"
	"github.com/stackql/stackql/internal/stackql/primitive"
	"github.com/stackql/stackql/internal/stackql/primitivecomposer"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/relational"
	"github.com/stackql/stackql/internal/stackql/symtab"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
	"github.com/stackql/stackql/internal/stackql/taxonomy"
	"github.com/stackql/stackql/internal/stackql/util"

	"github.com/stackql/stackql/pkg/prettyprint"
	"github.com/stackql/stackql/pkg/sqltypeutil"

	"github.com/stackql/go-openapistackql/openapistackql"

	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	_ PrimitiveGenerator = &standardPrimitiveGenerator{}
)

type PrimitiveGenerator interface {
	AnalyzeInsert(pbi planbuilderinput.PlanBuilderInput) error
	AnalyzeNop(pbi planbuilderinput.PlanBuilderInput) error
	AnalyzeUpdate(pbi planbuilderinput.PlanBuilderInput) error
	AnalyzePGInternal(pbi planbuilderinput.PlanBuilderInput) error
	AnalyzeRegistry(pbi planbuilderinput.PlanBuilderInput) error
	AnalyzeStatement(pbi planbuilderinput.PlanBuilderInput) error
	GetPrimitiveComposer() primitivecomposer.PrimitiveComposer
	IsShowResults() bool
}

type standardPrimitiveGenerator struct {
	Parent            PrimitiveGenerator
	Children          []PrimitiveGenerator
	PrimitiveComposer primitivecomposer.PrimitiveComposer
}

func NewRootPrimitiveGenerator(ast sqlparser.SQLNode, handlerCtx handler.HandlerContext, graph *primitivegraph.PrimitiveGraph) PrimitiveGenerator {
	tblMap := make(taxonomy.TblMap)
	symTab := symtab.NewHashMapTreeSymTab()
	return &standardPrimitiveGenerator{
		PrimitiveComposer: primitivecomposer.NewPrimitiveComposer(nil, ast, handlerCtx.GetDrmConfig(), handlerCtx.GetTxnCounterMgr(), graph, tblMap, symTab, handlerCtx.GetSQLEngine(), handlerCtx.GetSQLDialect(), handlerCtx.GetASTFormatter()),
	}
}

func (pb *standardPrimitiveGenerator) GetPrimitiveComposer() primitivecomposer.PrimitiveComposer {
	return pb.PrimitiveComposer
}

func (pb *standardPrimitiveGenerator) addChildPrimitiveGenerator(ast sqlparser.SQLNode, leaf symtab.SymTab) *standardPrimitiveGenerator {
	tables := pb.PrimitiveComposer.GetTables()
	switch node := ast.(type) {
	case sqlparser.Statement:
		logging.GetLogger().Infoln(fmt.Sprintf("creating new table map for node = %v", node))
		tables = make(taxonomy.TblMap)
	}
	retVal := &standardPrimitiveGenerator{
		Parent: pb,
		PrimitiveComposer: primitivecomposer.NewPrimitiveComposer(
			pb.PrimitiveComposer,
			ast,
			pb.PrimitiveComposer.GetDRMConfig(),
			pb.PrimitiveComposer.GetTxnCounterManager(),
			pb.PrimitiveComposer.GetGraph(),
			tables,
			leaf,
			pb.PrimitiveComposer.GetSQLEngine(),
			pb.PrimitiveComposer.GetSQLDialect(),
			pb.PrimitiveComposer.GetASTFormatter(),
		),
	}
	pb.Children = append(pb.Children, retVal)
	pb.PrimitiveComposer.AddChild(retVal.PrimitiveComposer)
	return retVal
}

func (pb *standardPrimitiveGenerator) comparisonExprToFilterFunc(table openapistackql.ITable, parentNode *sqlparser.Show, expr *sqlparser.ComparisonExpr) (func(openapistackql.ITable) (openapistackql.ITable, error), error) {
	qualifiedName, ok := expr.Left.(*sqlparser.ColName)
	if !ok {
		return nil, fmt.Errorf("unexpected: %v", sqlparser.String(expr))
	}
	if !qualifiedName.Qualifier.IsEmpty() {
		return nil, fmt.Errorf("unsupported qualifier for column: %v", sqlparser.String(qualifiedName))
	}
	colName := qualifiedName.Name.GetRawVal()
	tableContainsKey := table.KeyExists(colName)
	if !tableContainsKey {
		return nil, fmt.Errorf("col name = '%s' not found in table name = '%s'", colName, table.GetName())
	}
	_, lhsValErr := table.GetKeyAsSqlVal(colName)
	if lhsValErr != nil {
		return nil, lhsValErr
	}
	var resolved sqltypes.Value
	var rhsStr string
	switch right := expr.Right.(type) {
	case *sqlparser.SQLVal:
		if right.Type != sqlparser.IntVal && right.Type != sqlparser.StrVal {
			return nil, fmt.Errorf("unexpected: %v", sqlparser.String(expr))
		}
		pv, err := sqlparser.NewPlanValue(right)
		if err != nil {
			return nil, err
		}
		rhsStr = string(right.Val)
		resolved, err = pv.ResolveValue(nil)
		if err != nil {
			return nil, err
		}
	case sqlparser.BoolVal:
		var resErr error
		resolved, resErr = sqltypeutil.InterfaceToSQLType(right == true)
		if resErr != nil {
			return nil, resErr
		}
	default:
		return nil, fmt.Errorf("unexpected: %v", sqlparser.String(right))
	}
	var retVal func(openapistackql.ITable) (openapistackql.ITable, error)
	if expr.Operator == sqlparser.LikeStr || expr.Operator == sqlparser.NotLikeStr {
		likeRegexp, err := regexp.Compile(iqlutil.TranslateLikeToRegexPattern(rhsStr))
		if err != nil {
			return nil, err
		}
		retVal = relational.ConstructLikePredicateFilter(colName, likeRegexp, expr.Operator == sqlparser.NotLikeStr)
		pb.PrimitiveComposer.SetColVisited(colName, true)
		return retVal, nil
	}
	operatorPredicate, preErr := relational.GetOperatorPredicate(expr.Operator)

	if preErr != nil {
		return nil, preErr
	}

	pb.PrimitiveComposer.SetColVisited(colName, true)
	return relational.ConstructTablePredicateFilter(colName, resolved, operatorPredicate), nil
}

func getProviderServiceMap(item openapistackql.ProviderService, extended bool) map[string]interface{} {
	retVal := map[string]interface{}{
		"id":    item.ID,
		"name":  item.Name,
		"title": item.Title,
	}
	if extended {
		retVal["description"] = item.Description
		retVal["version"] = item.Version
	}
	return retVal
}

func convertProviderServicesToMap(services map[string]*openapistackql.ProviderService, extended bool) map[string]map[string]interface{} {
	retVal := make(map[string]map[string]interface{})
	for k, v := range services {
		retVal[k] = getProviderServiceMap(*v, extended)
	}
	return retVal
}

func filterResources(resources map[string]*openapistackql.Resource, tableFilter func(openapistackql.ITable) (openapistackql.ITable, error)) (map[string]*openapistackql.Resource, error) {
	var err error
	if tableFilter != nil {
		filteredResources := make(map[string]*openapistackql.Resource)
		for k, rsc := range resources {
			filteredResource, filterErr := tableFilter(rsc)
			if filterErr == nil && filteredResource != nil {
				filteredResources[k] = filteredResource.(*openapistackql.Resource)
			}
			if filterErr != nil {
				err = filterErr
			}
		}
		resources = filteredResources
	}
	return resources, err
}

func filterServices(services map[string]*openapistackql.ProviderService, tableFilter func(openapistackql.ITable) (openapistackql.ITable, error), useNonPreferredAPIs bool) (map[string]*openapistackql.ProviderService, error) {
	var err error
	if tableFilter != nil {
		filteredServices := make(map[string]*openapistackql.ProviderService)
		for k, svc := range services {
			if useNonPreferredAPIs || svc.Preferred {
				filteredService, filterErr := tableFilter(svc)
				if filterErr == nil && filteredService != nil {
					filteredServices[k] = (filteredService.(*openapistackql.ProviderService))
				}
				if filterErr != nil {
					err = filterErr
				}
			}
		}
		services = filteredServices
	}
	return services, err
}

func filterMethods(methods openapistackql.Methods, tableFilter func(openapistackql.ITable) (openapistackql.ITable, error)) (openapistackql.Methods, error) {
	var err error
	if tableFilter != nil {
		filteredMethods := make(openapistackql.Methods)
		for k, m := range methods {
			filteredMethod, filterErr := tableFilter(&m)
			if filterErr == nil && filteredMethod != nil {
				filteredMethods[k] = m
			}
			if filterErr != nil {
				err = filterErr
			}
		}
		methods = filteredMethods
	}
	return methods, err
}

func (pb *standardPrimitiveGenerator) inferProviderForShow(node *sqlparser.Show, handlerCtx handler.HandlerContext) error {
	nodeTypeUpperCase := strings.ToUpper(node.Type)
	switch nodeTypeUpperCase {
	case "AUTH":
		prov, err := handlerCtx.GetProvider(node.OnTable.Name.GetRawVal())
		if err != nil {
			return err
		}
		pb.PrimitiveComposer.SetProvider(prov)
	case "INSERT":
		prov, err := handlerCtx.GetProvider(node.OnTable.QualifierSecond.GetRawVal())
		if err != nil {
			return err
		}
		pb.PrimitiveComposer.SetProvider(prov)

	case "METHODS":
		prov, err := handlerCtx.GetProvider(node.OnTable.QualifierSecond.GetRawVal())
		if err != nil {
			return err
		}
		pb.PrimitiveComposer.SetProvider(prov)
	case "PROVIDERS":
		// no provider, might create some dummy object dunno
	case "RESOURCES":
		prov, err := handlerCtx.GetProvider(node.OnTable.Qualifier.GetRawVal())
		if err != nil {
			return err
		}
		pb.PrimitiveComposer.SetProvider(prov)
	case "SERVICES":
		prov, err := handlerCtx.GetProvider(node.OnTable.Name.GetRawVal())
		if err != nil {
			return err
		}
		pb.PrimitiveComposer.SetProvider(prov)
	default:
		return fmt.Errorf("unsuported node type: '%s'", node.Type)
	}
	return nil
}

func (pb *standardPrimitiveGenerator) ShowInstructionExecutor(node *sqlparser.Show, handlerCtx handler.HandlerContext) internaldto.ExecutorOutput {
	extended := strings.TrimSpace(strings.ToUpper(node.Extended)) == "EXTENDED"
	nodeTypeUpperCase := strings.ToUpper(node.Type)
	var keys map[string]map[string]interface{}
	var columnOrder []string
	var err error
	var filter func(interface{}) (openapistackql.ITable, error)
	logging.GetLogger().Infoln(fmt.Sprintf("filter type = %T", filter))
	switch nodeTypeUpperCase {
	case "AUTH":
		logging.GetLogger().Infoln(fmt.Sprintf("Show For node.Type = '%s'", node.Type))
		if err == nil {
			authCtx, err := handlerCtx.GetAuthContext(pb.PrimitiveComposer.GetProvider().GetProviderString())
			if err == nil {
				var authMeta *openapistackql.AuthMetadata
				authMeta, err = pb.PrimitiveComposer.GetProvider().ShowAuth(authCtx)
				if err == nil {
					keys = map[string]map[string]interface{}{
						"1": authMeta.ToMap(),
					}
					columnOrder = authMeta.GetHeaders()
				}
			}
		}
	case "INSERT":
		ppCtx := prettyprint.NewPrettyPrintContext(
			handlerCtx.GetRuntimeContext().OutputFormat == constants.PrettyTextStr,
			constants.DefaultPrettyPrintIndent,
			constants.DefaultPrettyPrintBaseIndent,
			"'",
			logging.GetLogger(),
		)
		tbl, err := pb.PrimitiveComposer.GetTable(node)
		if err != nil {
			return util.GenerateSimpleErroneousOutput(err)
		}
		meth, err := tbl.GetMethod()
		if err != nil {
			tblName, tblErr := tbl.GetStackQLTableName()
			if tblErr != nil {
				return util.GenerateSimpleErroneousOutput(fmt.Errorf("Cannot find matching operation, possible causes include missing required parameters or an unsupported method for the resource, to find required parameters for supported methods run SHOW METHODS: %s", err.Error()))
			}
			return util.GenerateSimpleErroneousOutput(fmt.Errorf("Cannot find matching operation, possible causes include missing required parameters or an unsupported method for the resource, to find required parameters for supported methods run SHOW METHODS IN %s: %s", tblName, err.Error()))
		}
		svc, err := tbl.GetService()
		if err != nil {
			return util.GenerateSimpleErroneousOutput(err)
		}
		pp := prettyprint.NewPrettyPrinter(ppCtx)
		requiredOnly := pb.PrimitiveComposer.GetCommentDirectives() != nil && pb.PrimitiveComposer.GetCommentDirectives().IsSet("REQUIRED")
		insertStmt, err := metadatavisitors.ToInsertStatement(node.Columns, meth, svc, extended, pp, requiredOnly)
		tableName, _ := tbl.GetTableName()
		if err != nil {
			return util.GenerateSimpleErroneousOutput(fmt.Errorf("error creating insert statement for %s: %s", tableName, err.Error()))
		}
		stmtStr := fmt.Sprintf(insertStmt, tableName)
		keys = map[string]map[string]interface{}{
			"1": {
				"insert_statement": stmtStr,
			},
		}
	case "METHODS":
		var rsc *openapistackql.Resource
		rsc, err = pb.PrimitiveComposer.GetProvider().GetResource(node.OnTable.Qualifier.GetRawVal(), node.OnTable.Name.GetRawVal(), handlerCtx.GetRuntimeContext())
		methods := rsc.GetMethodsMatched()
		tbl, err := pb.PrimitiveComposer.GetTable(node.OnTable)
		var filter func(openapistackql.ITable) (openapistackql.ITable, error)
		if err != nil {
			logging.GetLogger().Infoln(fmt.Sprintf("table and therefore filter not found for AST, shall procede nil filter"))
		} else {
			filter = tbl.GetTableFilter()
		}
		methods, err = filterMethods(methods, filter)
		if err != nil {
			return util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, keys, columnOrder, nil, err, nil))
		}
		mOrd, err := methods.OrderMethods()
		if err != nil {
			return util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, keys, columnOrder, nil, err, nil))
		}
		methodKeys := make(map[string]map[string]interface{})
		for i, k := range mOrd {
			method := k
			methMap := method.ToPresentationMap(extended)
			methodKeys[fmt.Sprintf("%06d", i)] = methMap
			columnOrder = method.GetColumnOrder(extended)
		}
		keys = methodKeys
	case "PROVIDERS":
		keys = handlerCtx.GetSupportedProviders(extended)
		rv := util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, keys, columnOrder, nil, err, nil))
		if len(keys) == 0 {
			rv = util.EmptyProtectResultSet(
				rv,
				[]string{"name", "version"},
			)
		}
		return rv
	case "RESOURCES":
		svcName := node.OnTable.Name.GetRawVal()
		if svcName == "" {
			return prepareErroneousResultSet(keys, columnOrder, fmt.Errorf("no service designated from which to resolve resources"))
		}
		var resources map[string]*openapistackql.Resource
		resources, err = pb.PrimitiveComposer.GetProvider().GetResourcesRedacted(svcName, handlerCtx.GetRuntimeContext(), extended)
		if err != nil {
			return prepareErroneousResultSet(keys, columnOrder, err)
		}
		columnOrder = openapistackql.GetResourcesHeader(extended)
		var filter func(openapistackql.ITable) (openapistackql.ITable, error)
		if err != nil {
			logging.GetLogger().Infoln(fmt.Sprintf("table and therefore filter not found for AST, shall procede nil filter"))
		} else {
			filter = pb.PrimitiveComposer.GetTableFilter()
		}
		resources, err = filterResources(resources, filter)
		if err != nil {
			return prepareErroneousResultSet(keys, columnOrder, err)
		}
		keys = make(map[string]map[string]interface{})
		for k, v := range resources {
			keys[k] = v.ToMap(extended)
		}
	case "SERVICES":
		logging.GetLogger().Infoln(fmt.Sprintf("Show For node.Type = '%s': Displaying services for provider = '%s'", node.Type, pb.PrimitiveComposer.GetProvider().GetProviderString()))
		var services map[string]*openapistackql.ProviderService
		services, err = pb.PrimitiveComposer.GetProvider().GetProviderServicesRedacted(handlerCtx.GetRuntimeContext(), extended)
		if err != nil {
			return prepareErroneousResultSet(keys, columnOrder, err)
		}
		columnOrder = openapistackql.GetServicesHeader(extended)
		services, err = filterServices(services, pb.PrimitiveComposer.GetTableFilter(), handlerCtx.GetRuntimeContext().UseNonPreferredAPIs)
		if err != nil {
			return prepareErroneousResultSet(keys, columnOrder, err)
		}
		keys = convertProviderServicesToMap(services, extended)
	}
	return util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, keys, columnOrder, nil, err, nil))
}

func prepareErroneousResultSet(rowMap map[string]map[string]interface{}, columnOrder []string, err error) internaldto.ExecutorOutput {
	return util.PrepareResultSet(
		internaldto.NewPrepareResultSetDTO(
			nil,
			rowMap,
			columnOrder,
			nil,
			err,
			nil,
		),
	)
}

func (pb *standardPrimitiveGenerator) DescribeInstructionExecutor(handlerCtx handler.HandlerContext, tbl tablemetadata.ExtendedTableMetadata, extended bool, full bool) internaldto.ExecutorOutput {
	schema, err := tbl.GetSelectableObjectSchema()
	if err != nil {
		return internaldto.NewErroneousExecutorOutput(err)
	}
	columnOrder := openapistackql.GetDescribeHeader(extended)
	descriptionMap := schema.ToDescriptionMap(extended)
	keys := make(map[string]map[string]interface{})
	for k, v := range descriptionMap {
		switch val := v.(type) {
		case map[string]interface{}:
			keys[k] = val
		}
	}
	return util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, keys, columnOrder, util.DescribeRowSort, err, nil))
}

func (pb *standardPrimitiveGenerator) LocalSelectExecutor(handlerCtx handler.HandlerContext, node *sqlparser.Select, rowSort func(map[string]map[string]interface{}) []string) (primitive.IPrimitive, error) {
	return primitive.NewLocalPrimitive(
		func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
			var columnOrder []string
			keys := make(map[string]map[string]interface{})
			row := make(map[string]interface{})
			for idx := range pb.PrimitiveComposer.GetValOnlyColKeys() {
				col := pb.PrimitiveComposer.GetValOnlyCol(idx)
				if col != nil {
					var alias string
					var val interface{}
					for k, v := range col {
						alias = k
						val = v
						break
					}
					if alias == "" {
						alias = "val_" + strconv.Itoa(idx)
					}
					row[alias] = val
					columnOrder = append(columnOrder, alias)
				}
			}
			keys["0"] = row
			return util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, keys, columnOrder, rowSort, nil, nil))
		}), nil
}

func (pb *standardPrimitiveGenerator) IsShowResults() bool {
	return pb.isShowResults()
}

func (pb *standardPrimitiveGenerator) isShowResults() bool {
	return pb.PrimitiveComposer.GetCommentDirectives() != nil && pb.PrimitiveComposer.GetCommentDirectives().IsSet("SHOWRESULTS")
}
