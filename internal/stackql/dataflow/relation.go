package dataflow

import (
	"fmt"

	"vitess.io/vitess/go/vt/sqlparser"

	"github.com/stackql/go-openapistackql/openapistackql"
)

type DataFlowRelation interface {
	GetProjection() (string, string, error)
	GetSelectExpr() (sqlparser.SelectExpr, error)
	GetColumnDescriptor() (openapistackql.ColumnDescriptor, error)
}

type StandardDataFlowRelation struct {
	comparisonExpr *sqlparser.ComparisonExpr
	destColumn     *sqlparser.ColName
	sourceExpr     sqlparser.Expr
}

func NewStandardDataFlowRelation(
	comparisonExpr *sqlparser.ComparisonExpr,
	destColumn *sqlparser.ColName,
	sourceExpr sqlparser.Expr,
) DataFlowRelation {
	return &StandardDataFlowRelation{
		comparisonExpr: comparisonExpr,
		destColumn:     destColumn,
		sourceExpr:     sourceExpr,
	}
}

func (dr *StandardDataFlowRelation) GetProjection() (string, string, error) {
	switch se := dr.sourceExpr.(type) {
	case *sqlparser.ColName:
		return se.Name.GetRawVal(), dr.destColumn.Name.GetRawVal(), nil
	default:
		return "", "", fmt.Errorf("cannot project from expression type = '%T'", se)
	}
}

func (dr *StandardDataFlowRelation) GetSelectExpr() (sqlparser.SelectExpr, error) {
	rv := &sqlparser.AliasedExpr{
		Expr: dr.sourceExpr,
		As:   dr.destColumn.Name,
	}
	return rv, nil
}

func (dr *StandardDataFlowRelation) GetColumnDescriptor() (openapistackql.ColumnDescriptor, error) {
	decoratedColumn := fmt.Sprintf(`%s AS %s`, sqlparser.String(dr.sourceExpr), dr.destColumn.Name.GetRawVal())
	cd := openapistackql.NewColumnDescriptor("", "", decoratedColumn, nil, nil)
	return cd, nil
}
