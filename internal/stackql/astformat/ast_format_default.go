package astformat

import (
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

func DefaultSelectExprsFormatter(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {
	switch node := node.(type) {
	case sqlparser.ColIdent:
		formatColIdent(node, buf)
		return
	case *sqlparser.IntervalExpr:
		sb := sqlparser.NewTrackedBuffer(PostgresSelectExprsFormatter)
		sb.AstPrintf(node, "%s '%v %s'", "INTERVAL", node.Expr, node.Unit)
	default:
		node.Format(buf)
		return
	}
}
