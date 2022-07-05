package dataflow

import (
	"github.com/stackql/stackql/internal/stackql/taxonomy"
	"vitess.io/vitess/go/vt/sqlparser"
)

type DataFlowVertex interface {
	DataFlowUnit
	GetAnnotation() taxonomy.AnnotationCtx
	GetTableExpr() sqlparser.TableExpr
}

type StandardDataFlowVertex struct {
	collection DataFlowCollection
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

func (dv *StandardDataFlowVertex) GetAnnotation() taxonomy.AnnotationCtx {
	return dv.annotation
}

func (dv *StandardDataFlowVertex) GetTableExpr() sqlparser.TableExpr {
	return dv.tableExpr
}
