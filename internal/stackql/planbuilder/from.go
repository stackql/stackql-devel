package planbuilder

import (
	"fmt"

	"github.com/stackql/stackql/internal/stackql/parserutil"
	"github.com/stackql/stackql/internal/stackql/taxonomy"
	"vitess.io/vitess/go/vt/sqlparser"
)

func analyzeFrom(from sqlparser.TableExprs, router *parserutil.ParameterRouter) ([]*taxonomy.ExtendedTableMetadata, error) {
	if len(from) > 1 {
		return nil, fmt.Errorf("cannot accomodate cartesian joins")
	}

	return nil, nil
}
