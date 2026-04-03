package queryshape

import (
	"github.com/stackql/psql-wire/pkg/sqldata"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/planbuilder"
	"github.com/stackql/stackql/internal/stackql/typing"
)

// InferResultColumns analyses a SQL query and returns result column metadata
// without executing the query.  It builds a query plan to resolve column
// projections and types, but does not run the plan.
//
// Returns nil when column metadata cannot be derived (e.g. DDL, mutations,
// or queries that fail planning).
func InferResultColumns(
	handlerCtx handler.HandlerContext,
	query string,
) []sqldata.ISQLColumn {
	clonedCtx := handlerCtx.Clone()
	clonedCtx.SetQuery(query)
	clonedCtx.SetRawQuery(query)
	pb := planbuilder.NewPlanBuilder(nil)
	qPlan, err := pb.BuildPlanFromContext(clonedCtx)
	if err != nil || qPlan == nil {
		return nil
	}
	colMeta := qPlan.GetColumnMetadata()
	if len(colMeta) == 0 {
		return nil
	}
	return ColumnMetadataToSQLColumns(colMeta)
}

// ColumnMetadataToSQLColumns converts internal column metadata to
// wire protocol ISQLColumn objects.
func ColumnMetadataToSQLColumns(cols []typing.ColumnMetadata) []sqldata.ISQLColumn {
	table := sqldata.NewSQLTable(0, "")
	result := make([]sqldata.ISQLColumn, len(cols))
	for i, col := range cols {
		result[i] = sqldata.NewSQLColumn(
			table,
			col.GetIdentifier(),
			0,
			uint32(col.GetColumnOID()),
			-1,
			0,
			"text",
		)
	}
	return result
}
