package dataflow

import (
	"github.com/stackql/stackql/internal/stackql/taxonomy"
	"vitess.io/vitess/go/vt/sqlparser"
)

type DataFlowVertex interface {
	DataFlowUnit
}

type StandardDataFlowVertex struct {
	annotation taxonomy.AnnotationCtx
	tableExpr  sqlparser.TableExpr
}

func NewStandardDataFlowVertex(
	annotation taxonomy.AnnotationCtx,
	tableExpr sqlparser.TableExpr) DataFlowVertex {
	return &StandardDataFlowVertex{
		annotation: annotation,
		tableExpr:  tableExpr,
	}
}

func (dv *StandardDataFlowVertex) iDataFlowUnit() {}
