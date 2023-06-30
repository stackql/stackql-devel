package internaldto

import (
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

var (
	_ FromSubtree = &simpleFromSubtree{}
)

type tblLookasideMap struct {
	hoistedTables []sqlparser.SQLNode
}

type FromSubtree interface {
}

type simpleFromSubtree struct {
}

func NewFromSubtree() (FromSubtree, error) {
	return &simpleFromSubtree{}, nil
}

func (f *simpleFromSubtree) Render() string {
	return ""
}
