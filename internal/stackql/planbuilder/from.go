package planbuilder

import (
	"fmt"

	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/parserutil"
	"github.com/stackql/stackql/internal/stackql/taxonomy"
	"vitess.io/vitess/go/vt/sqlparser"
)

func analyzeFrom(from sqlparser.TableExprs, tableMap parserutil.TableExprMap) ([]*taxonomy.ExtendedTableMetadata, error) {
	if len(from) > 1 {
		return nil, fmt.Errorf("cannot accomodate cartesian joins")
	}
	//      tableRoot := from[0]

	return nil, nil
}

func analyzeAliasedTable(handlerCtx *handler.HandlerContext, tb *sqlparser.AliasedTableExpr, tableMap parserutil.TableExprMap) (meta *taxonomy.ExtendedTableMetadata, err error) {
	switch expr := tb.Expr.(type) {
	case sqlparser.TableName:
		sm := tableMap.SingleTableMap(expr)
		params := sm.ToStringMap()
		_, _, err := taxonomy.GetHeirarchyFromStatement(handlerCtx, tb, params)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
