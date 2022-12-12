package astindirect

import (
	"fmt"

	"github.com/stackql/stackql/internal/stackql/parse"

	"github.com/stackql/stackql/internal/stackql/internaldto"
	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	_ Indirect = &view{}
)

type IndirectType int

const (
	ViewType IndirectType = iota
	SubqueryType
	CTEType
)

func NewViewIndirect(viewDTO internaldto.ViewDTO) (Indirect, error) {
	rv := &view{
		viewDTO: viewDTO,
	}
	return rv, nil
}

type Indirect interface {
	Parse() error
	GetSelectAST() sqlparser.SelectStatement
	GetType() IndirectType
}

type view struct {
	viewDTO    internaldto.ViewDTO
	selectStmt sqlparser.SelectStatement
}

func (v *view) GetType() IndirectType {
	return ViewType
}

func (v *view) getAST() (sqlparser.Statement, error) {
	return parse.ParseQuery(v.viewDTO.GetRawQuery())
}

func (v *view) GetSelectAST() sqlparser.SelectStatement {
	return v.selectStmt
}

func (v *view) Parse() error {
	parseResult, err := v.getAST()
	if err != nil {
		return err
	}
	switch pr := parseResult.(type) {
	case sqlparser.SelectStatement:
		v.selectStmt = pr
		return nil
	default:
		return fmt.Errorf("view of type '%T' not yet supported", pr)
	}
}
