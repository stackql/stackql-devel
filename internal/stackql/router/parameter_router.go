package router

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/parserutil"
	"github.com/stackql/stackql/internal/stackql/taxonomy"
	"vitess.io/vitess/go/vt/sqlparser"
)

// Parameter router supports
// mapping columnar input to
// tabular output.
// This is for dealing with parser objects, prior to assignment
// of openapi schemas.
// The storage medium for constituents is abstracted.
// As of now this is a multi stage object, violates Single functionality.
type ParameterRouter interface {

	// Obtains parameters that are unammbiguous (eg: aliased, unique col name)
	// or potential matches for a supplied table.
	// getAvailableParameters(tb sqlparser.TableExpr) *parserutil.TableParameterCoupling

	// Records the fact that parameters have been assigned to a table and
	// cannot be used elsewhere.
	// invalidateParams(params map[string]interface{}) error

	// First pass assignment of columnar objects
	// to tables, only for HTTP method parameters.  All data accrual is done herein:
	//   - SQL parser table objects mapped to hierarchy.
	//   - Data flow dependencies identified and persisted.
	//   - Hierarchies may be persisted for analysis.
	// Detects bi-directional data flow errors and returns error if so.
	// Returns:
	//   - Hierarchy.
	//   - Columnar objects definitely assigned as HTTP method parameters.
	//   - Error if applicable.
	Route(tb sqlparser.TableExpr, handler *handler.HandlerContext) (*taxonomy.ExtendedTableMetadata, map[string]interface{}, error)

	// Detects:
	//   - Dependency cycle.
	AnalyzeDependencies() error

	GetOnConditionsToRewrite() map[*sqlparser.ComparisonExpr]struct{}

	GetOnConditionDataFlows() (map[*taxonomy.ExtendedTableMetadata]*taxonomy.ExtendedTableMetadata, error)
}

type StandardParameterRouter struct {
	tablesAliasMap                parserutil.TableAliasMap
	tableMap                      parserutil.TableExprMap
	onParamMap                    parserutil.ParameterMap
	whereParamMap                 parserutil.ParameterMap
	colRefs                       parserutil.ColTableMap
	comparisonToTableDependencies parserutil.ComparisonTableMap
	tableToComparisonDependencies parserutil.ComparisonTableMap
	tableToMetadata               map[sqlparser.TableExpr]*taxonomy.ExtendedTableMetadata
	invalidatedParams             map[string]interface{}
}

func NewParameterRouter(
	tablesAliasMap parserutil.TableAliasMap,
	tableMap parserutil.TableExprMap,
	whereParamMap parserutil.ParameterMap,
	onParamMap parserutil.ParameterMap,
	colRefs parserutil.ColTableMap,
) ParameterRouter {
	return &StandardParameterRouter{
		tablesAliasMap:                tablesAliasMap,
		tableMap:                      tableMap,
		whereParamMap:                 whereParamMap,
		onParamMap:                    onParamMap,
		colRefs:                       colRefs,
		invalidatedParams:             make(map[string]interface{}),
		comparisonToTableDependencies: make(parserutil.ComparisonTableMap),
		tableToComparisonDependencies: make(parserutil.ComparisonTableMap),
		tableToMetadata:               make(map[sqlparser.TableExpr]*taxonomy.ExtendedTableMetadata),
	}
}

func (pr *StandardParameterRouter) AnalyzeDependencies() error {
	// for k, v := range pr.comparisonToTableDependencies {
	// }
	return nil
}

func (pr *StandardParameterRouter) GetOnConditionsToRewrite() map[*sqlparser.ComparisonExpr]struct{} {
	rv := make(map[*sqlparser.ComparisonExpr]struct{})
	for k, _ := range pr.comparisonToTableDependencies {
		rv[k] = struct{}{}
	}
	return rv
}

func (pr *StandardParameterRouter) extractDataFlowDependency(input sqlparser.Expr) (*taxonomy.ExtendedTableMetadata, error) {
	switch l := input.(type) {
	case *sqlparser.ColName:
		// leave unknown for now -- bit of a mess
		ref, err := parserutil.NewColumnarReference(l, parserutil.UnknownParam)
		if err != nil {
			return nil, err
		}
		tb, ok := pr.colRefs[ref]
		if !ok {
			return nil, fmt.Errorf("unassigned column in ON condition dataflow; please alias column '%s'", l.GetRawVal())
		}
		hr, ok := pr.tableToMetadata[tb]
		if !ok {
			return nil, fmt.Errorf("cannot assign hierarchy for column '%s'", l.GetRawVal())
		}
		return hr, nil
	default:
		return nil, fmt.Errorf("cannot accomodate ON condition of type = '%T'", l)
	}
}

func (pr *StandardParameterRouter) GetOnConditionDataFlows() (map[*taxonomy.ExtendedTableMetadata]*taxonomy.ExtendedTableMetadata, error) {
	rv := make(map[*taxonomy.ExtendedTableMetadata]*taxonomy.ExtendedTableMetadata)
	for k, v := range pr.comparisonToTableDependencies {
		selfTableCited := false
		v2, ok := pr.tableToMetadata[v]
		if !ok {
			return nil, fmt.Errorf("table expression '%s' has not been assigned to hierarchy", sqlparser.String(v))
		}
		var dependency *taxonomy.ExtendedTableMetadata
		switch l := k.Left.(type) {
		case *sqlparser.ColName:
			lhr, err := pr.extractDataFlowDependency(l)
			if err != nil {
				return nil, err
			}
			if v2 == lhr {
				selfTableCited = true
			} else {
				dependency = lhr
			}
		}
		switch r := k.Right.(type) {
		case *sqlparser.ColName:
			rhr, err := pr.extractDataFlowDependency(r)
			if err != nil {
				return nil, err
			}
			if v2 == rhr {
				if selfTableCited {
					return nil, fmt.Errorf("table join ON comparison '%s' is self referencing", sqlparser.String(k))
				}
				selfTableCited = true
			} else {
				dependency = rhr
			}
		}
		if !selfTableCited {
			return nil, fmt.Errorf("table join ON comparison '%s' referencing incomplete", sqlparser.String(k))
		}
		rv[dependency] = v2
	}
	return rv, nil
}

func (pr *StandardParameterRouter) getAvailableParameters(tb sqlparser.TableExpr) parserutil.TableParameterCoupling {
	rv := parserutil.NewTableParameterCoupling()
	for k, v := range pr.whereParamMap.GetMap() {
		key := k.String()
		tableAlias := k.Alias()
		foundTable, ok := pr.tablesAliasMap[tableAlias]
		if ok && foundTable != tb {
			continue
		}
		if pr.isInvalidated(key) {
			continue
		}
		ref, ok := pr.colRefs[k]
		if ok && ref != tb {
			continue
		}
		rv.Add(k, v, parserutil.WhereParam)
	}
	for k, v := range pr.onParamMap.GetMap() {
		key := k.String()
		tableAlias := k.Alias()
		foundTable, ok := pr.tablesAliasMap[tableAlias]
		if ok && foundTable != tb {
			continue
		}
		if pr.isInvalidated(key) {
			continue
		}
		ref, ok := pr.colRefs[k]
		if ok && ref != tb {
			continue
		}
		val := v.GetVal()
		switch val := val.(type) {
		case *sqlparser.ColName:
			log.Debugf("%v\n", val)
			rhsAlias := val.Qualifier.GetRawVal()
			log.Debugf("%v\n", rhsAlias)
			foundTable, ok := pr.tablesAliasMap[rhsAlias]
			if ok && foundTable != tb {
				//
			}
		}
		rv.Add(k, v, parserutil.JoinOnParam)
	}
	return rv
}

func (pr *StandardParameterRouter) invalidateParams(params map[string]interface{}) error {
	for k, v := range params {
		err := pr.invalidate(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pr *StandardParameterRouter) isInvalidated(key string) bool {
	_, ok := pr.invalidatedParams[key]
	return ok
}

func (pr *StandardParameterRouter) invalidate(key string, val interface{}) error {
	if pr.isInvalidated(key) {
		return fmt.Errorf("parameter '%s' already invalidated", key)
	}
	pr.invalidatedParams[key] = val
	return nil
}

// Route will map columnar input to a supplied
// parser table object.
// Columnar input may come from either where clause
// or on conditions.
func (pr *StandardParameterRouter) Route(tb sqlparser.TableExpr, handlerCtx *handler.HandlerContext) (*taxonomy.ExtendedTableMetadata, map[string]interface{}, error) {
	for k, v := range pr.whereParamMap.GetMap() {
		log.Infof("%v\n", v)
		alias := k.Alias()
		if alias == "" {
			continue
		}
		t, ok := pr.tablesAliasMap[alias]
		if !ok {
			return nil, nil, fmt.Errorf("alias '%s' does not map to any table expression", alias)
		}
		if t == tb {
			ref, ok := pr.colRefs[k]
			if ok && ref != t {
				return nil, nil, fmt.Errorf("failed parameter routing, cannot re-assign")
			}
			pr.colRefs[k] = t
		}
	}
	for k, v := range pr.onParamMap.GetMap() {
		log.Infof("%v\n", v)
		alias := k.Alias()
		if alias == "" {
			continue
		}
		t, ok := pr.tablesAliasMap[alias]
		if !ok {
			return nil, nil, fmt.Errorf("alias '%s' does not map to any table expression", alias)
		}
		if t == tb {
			ref, ok := pr.colRefs[k]
			if ok && ref != t {
				return nil, nil, fmt.Errorf("failed parameter routing, cannot re-assign")
			}
			pr.colRefs[k] = t
		}
	}
	// These are "available parameters"
	tpc := pr.getAvailableParameters(tb)
	// After executing GetHeirarchyFromStatement(), we know:
	//   - Any remaining param is not required.
	//   - Any "on" param that was consumed:
	//      - Can / must be from removed join conditions in a rewrite. [Requires Join in router for later rewrite].
	//      - Defines a sequencing and data flow dependency unless RHS is a literal. [Create new object to represent].
	// TODO: In order to do this, we can, for each table:
	//   1. [*] Subtract the remaining parameters returned by GetHeirarchyFromStatement()
	//      from the available parameters.  Will need reversible string to object translation.
	//   2. [*] Identify "on" parameters that were consumed as per item #1.
	//      We are free to change the "table parameter coupling" API to accomodate
	//      items #1 and #2.
	//   3. [*] If #2 is consumed, then:
	//        - [*] Tag the "on" comparison as being incident to the table.
	//        - [*] Tag the "on" comparison for later rewrite to NOP.
	//      Probably some
	//      new data structure to accomodate this.
	// And then, once all tables are done and also therefore, all hierarchies are present:
	//   a) [ ] Assign all remaining on parameters based on schema.
	//   b) [ ] Represent assignments as edges from table to on condition.
	//   d) [ ] Throw error for disallowed scenarios:
	//        - Dual outgoing from ON object.
	//   e) [ ] Rewrite NOP on clauses.
	//   f) [ ] Catalogue and return dataflows (somehow)
	stringParams := tpc.GetStringified()
	hr, remainingParams, err := taxonomy.GetHeirarchyFromStatement(handlerCtx, tb, stringParams)
	log.Infof("hr = '%+v', remainingParams = '%+v', err = '%+v'", hr, remainingParams, err)
	if err != nil {
		return nil, nil, err
	}
	reconstitutedConsumedParams, err := tpc.ReconstituteConsumedParams(remainingParams)
	if err != nil {
		return nil, nil, err
	}
	abbreviatedConsumedMap, err := reconstitutedConsumedParams.AbbreviateMap()
	if err != nil {
		return nil, nil, err
	}
	onConsumed := reconstitutedConsumedParams.GetOnCoupling()
	pms := onConsumed.GetAllParameters()
	log.Infof("onConsumed = '%+v'", onConsumed)
	for _, kv := range pms {
		// In this stanza:
		//   1. [*] mark comparisons for rewriting
		//   2. [*] some sequencing data to be stored
		p := kv.V.GetParent()
		existingTable, ok := pr.comparisonToTableDependencies[p]
		if ok {
			return nil, nil, fmt.Errorf("data flow violation detected: ON comparison expression '%s' is a  dependency for tables '%s' and '%s'", sqlparser.String(p), sqlparser.String(existingTable), sqlparser.String(tb))
		}
		pr.comparisonToTableDependencies[p] = tb
		// this can be done, not sure if it is the best way
		// rewriteComparisonExpr(p)
		log.Infof("%v", kv)
	}
	m := taxonomy.NewExtendedTableMetadata(hr, taxonomy.GetAliasFromStatement(tb))
	// store relationship from sqlparser table expression to
	// hierarchy.  This enables e2e relationship
	// from expression to hierarchy.
	// eg: "on" clause to openapi method
	pr.tableToMetadata[tb] = m
	return m, abbreviatedConsumedMap, nil
}

func rewriteComparisonExpr(ex *sqlparser.ComparisonExpr) {
	ex = &sqlparser.ComparisonExpr{
		Left:     &sqlparser.SQLVal{Type: sqlparser.IntVal, Val: []byte("1")},
		Right:    &sqlparser.SQLVal{Type: sqlparser.IntVal, Val: []byte("1")},
		Operator: ex.Operator,
		Escape:   ex.Escape,
	}
}
