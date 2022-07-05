package dataflow

import (
	"github.com/stackql/stackql/internal/stackql/taxonomy"
	"vitess.io/vitess/go/vt/sqlparser"
)

type DataFlowVertex interface {
}

type StandardDataFlowVertex struct {
	hierarchy *taxonomy.ExtendedTableMetadata
	tableExpr sqlparser.TableExpr
}

func NewStandardDataFlowVertex(
	hierarchy *taxonomy.ExtendedTableMetadata,
	tableExpr sqlparser.TableExpr) DataFlowVertex {
	return &StandardDataFlowVertex{
		hierarchy: hierarchy,
		tableExpr: tableExpr,
	}
}
