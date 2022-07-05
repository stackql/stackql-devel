package dataflow

import (
	"vitess.io/vitess/go/vt/sqlparser"
)

type DataFlowEdge interface {
}

type StandardDataFlowEdge struct {
	source, dest   DataFlowVertex
	comparisonExpr *sqlparser.ComparisonExpr
	destColumn     *sqlparser.ColName
	sourceExpr     sqlparser.Expr
}

func NewStandardDataFlowEdge(
	source DataFlowVertex,
	dest DataFlowVertex,
	comparisonExpr *sqlparser.ComparisonExpr,
	sourceExpr sqlparser.Expr,
	destColumn *sqlparser.ColName,
) DataFlowEdge {
	return &StandardDataFlowEdge{
		source:         source,
		dest:           dest,
		comparisonExpr: comparisonExpr,
		sourceExpr:     sourceExpr,
		destColumn:     destColumn,
	}
}
