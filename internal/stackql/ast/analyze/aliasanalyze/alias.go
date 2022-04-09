package aliasanalyze

import (
	"fmt"

	"vitess.io/vitess/go/vt/sqlparser"

	log "github.com/sirupsen/logrus"
)

type ResolvedAliases = map[string]ResolvedAlias

type ResolvedAlias struct {
	Alias string
	Table interface{}
}

func ResolveAliases(stmt sqlparser.Statement) (ResolvedAliases, error) {
	switch s := stmt.(type) {
	case *sqlparser.Select:
		log.Infof("s = %v", s)
	default:
		return nil, fmt.Errorf("")
	}

	return nil, fmt.Errorf("")
}

// func annotateSelectCols(sel *sqlparser.Select) error {
// 	for _,  := range
// }
