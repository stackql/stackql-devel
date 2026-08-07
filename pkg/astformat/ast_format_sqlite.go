package astformat

import (
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

func SQLiteSelectExprsFormatter(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {
	switch node := node.(type) {
	case *sqlparser.SQLVal:
		if node.Type == sqlparser.StrVal {
			formatStrVal(node, buf)
			return
		}
		node.Format(buf)
		return
	case sqlparser.ColIdent:
		formatColIdentCaseInsensitive(node, buf)
		return

	default:
		node.Format(buf)
		return
	}
}
