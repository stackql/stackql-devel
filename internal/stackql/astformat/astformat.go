package astformat

import (
	"fmt"

	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	_ sqlparser.NodeFormatter = postgresFormatter
)

func jj() {
	fmt.Printf("\n")
}

func postgresFormatter(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {
	switch node := node.(type) {
	default:
		node.Format(buf)
	}
}
