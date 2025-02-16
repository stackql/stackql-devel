package astformat

import (
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

func DefaultSelectExprsFormatter(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {
	switch node := node.(type) {
	case sqlparser.ColIdent:
		formatColIdent(node, buf)
		return
	case *sqlparser.SQLVal:
		// switch node.Type {
		// case sqlparser.StrVal:
		// 	buf.Myprintf("%s", node.Val)
		// 	return
		// default:
		// 	node.Format(buf)
		// 	return
		// }
		node.Format(buf)
		return
	default:
		node.Format(buf)
		return
	}
}
