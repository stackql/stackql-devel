package parserutil

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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
	GetAvailableParameters(tb sqlparser.TableExpr) *TableParameterCoupling
	// Records the fact that parameters have been assigned to a table and
	// cannot be used elsewhere.
	InvalidateParams(params map[string]interface{}) error
	// First pass, tentative assignment of columnar objects
	// to tables.
	Route(tb sqlparser.TableExpr) error
}

type IParameterRouter struct {
	tablesAliasMap    TableAliasMap
	tableMap          TableExprMap
	onParamMap        ParameterMap
	whereParamMap     ParameterMap
	colRefs           ColTableMap
	invalidatedParams map[string]interface{}
}

func NewParameterRouter(tablesAliasMap TableAliasMap, tableMap TableExprMap, whereParamMap ParameterMap, onParamMap ParameterMap, colRefs ColTableMap) ParameterRouter {
	return &IParameterRouter{
		tablesAliasMap:    tablesAliasMap,
		tableMap:          tableMap,
		whereParamMap:     whereParamMap,
		onParamMap:        onParamMap,
		colRefs:           colRefs,
		invalidatedParams: make(map[string]interface{}),
	}
}

func (pr *IParameterRouter) GetAvailableParameters(tb sqlparser.TableExpr) *TableParameterCoupling {
	rv := NewTableParameterCoupling()
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
		rv.Add(k, v)
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
		rv.Add(k, v)
	}
	return rv
}

func (pr *IParameterRouter) InvalidateParams(params map[string]interface{}) error {
	for k, v := range params {
		err := pr.invalidate(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pr *IParameterRouter) isInvalidated(key string) bool {
	_, ok := pr.invalidatedParams[key]
	return ok
}

func (pr *IParameterRouter) invalidate(key string, val interface{}) error {
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
func (pr *IParameterRouter) Route(tb sqlparser.TableExpr) error {
	for k, v := range pr.whereParamMap.GetMap() {
		log.Infof("%v\n", v)
		alias := k.Alias()
		if alias == "" {
			continue
		}
		t, ok := pr.tablesAliasMap[alias]
		if !ok {
			return fmt.Errorf("alias '%s' does not map to any table expression", alias)
		}
		if t == tb {
			ref, ok := pr.colRefs[k]
			if ok && ref != t {
				return fmt.Errorf("failed parameter routing, cannot re-assign")
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
			return fmt.Errorf("alias '%s' does not map to any table expression", alias)
		}
		if t == tb {
			ref, ok := pr.colRefs[k]
			if ok && ref != t {
				return fmt.Errorf("failed parameter routing, cannot re-assign")
			}
			pr.colRefs[k] = t
		}
	}
	return nil
}
