package dataflow

import (
	"gonum.org/v1/gonum/graph"
	"vitess.io/vitess/go/vt/sqlparser"
)

type DataFlowEdge interface {
	graph.WeightedEdge
	GetDest() DataFlowVertex
	GetSource() DataFlowVertex
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

func (de *StandardDataFlowEdge) From() graph.Node {
	return de.source
}

func (de *StandardDataFlowEdge) To() graph.Node {
	return de.dest
}

func (de *StandardDataFlowEdge) ReversedEdge() graph.Edge {
	// Reversal is invalid given the assymetric
	// expressions, therefore returning unaltered
	// as per library recommmendation.
	return de
}

func (de *StandardDataFlowEdge) Weight() float64 {
	return 1.0
}

func (de *StandardDataFlowEdge) GetSource() DataFlowVertex {
	return de.source
}

func (de *StandardDataFlowEdge) GetDest() DataFlowVertex {
	return de.dest
}
