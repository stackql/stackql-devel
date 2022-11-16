package pginternal

import (
	"regexp"

	"github.com/stackql/stackql/internal/stackql/constants"
	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/logging"
	"github.com/stackql/stackql/internal/stackql/sqldialect"
	"vitess.io/vitess/go/vt/sqlparser"
)

var (
	_                        PGInternalRouter = &standardPGInternalRouter{}
	multipleWhitespaceRegexp *regexp.Regexp   = regexp.MustCompile(`\s+`)
	internalTableRegexp      *regexp.Regexp   = regexp.MustCompile(`(?i)^(?:public\.)?(?:pg_type|pg_catalog.*|current_schema)`)
	showHousekeepingRegexp   *regexp.Regexp   = regexp.MustCompile(`(?i)(?:\s+transaction\s+isolation\s+level|standard_conforming_strings)`)
)

type PGInternalRouter interface {
	CanRoute(node sqlparser.SQLNode) (constants.BackendQueryType, bool)
}

func GetPGInternalRouter(cfg dto.PGInternalCfg, sqlDialect sqldialect.SQLDialect) (PGInternalRouter, error) {
	return &standardPGInternalRouter{
		cfg:        cfg,
		sqlDialect: sqlDialect,
	}, nil
}

type standardPGInternalRouter struct {
	cfg        dto.PGInternalCfg
	sqlDialect sqldialect.SQLDialect
}

func (pgr *standardPGInternalRouter) CanRoute(node sqlparser.SQLNode) (constants.BackendQueryType, bool) {
	if pgr.sqlDialect.GetName() != constants.SQLDialectPostgres {
		return pgr.negative()
	}
	switch node := node.(type) {
	case *sqlparser.Select:
		logging.GetLogger().Debugf("node = %v\n", node)
		return pgr.analyzeSelect(node)
	case *sqlparser.Set:
		return constants.BackendExec, true
	case *sqlparser.Show:
		return pgr.analyzeShow(node)
	}
	return pgr.negative()
}

func (pgr *standardPGInternalRouter) negative() (constants.BackendQueryType, bool) {
	return constants.BackendNop, false
}

func (pgr *standardPGInternalRouter) analyzeSelect(node *sqlparser.Select) (constants.BackendQueryType, bool) {
	if len(node.From) < 1 {
		return pgr.negative()
	}
	if pgr.analyzeTableExpr(node.From[0]) {
		return constants.BackendQuery, true
	}
	return pgr.negative()
}

func (pgr *standardPGInternalRouter) analyzeShow(node *sqlparser.Show) (constants.BackendQueryType, bool) {
	if node.Type != "" && showHousekeepingRegexp.MatchString(node.Type) {
		return constants.BackendQuery, true
	}
	if pgr.analyzeTableName(node.OnTable) {
		if pgr.analyzeTableName(node.OnTable) {
			return constants.BackendQuery, true
		}
	}
	return pgr.negative()
}

func (pgr *standardPGInternalRouter) analyzeTableExpr(node sqlparser.TableExpr) bool {
	switch node := node.(type) {
	case *sqlparser.AliasedTableExpr:
		switch expr := node.Expr.(type) {
		case sqlparser.TableName:
			return pgr.analyzeTableName(expr)
		}
	}
	return false
}

func (pgr *standardPGInternalRouter) analyzeTableName(node sqlparser.TableName) bool {
	rawName := node.GetRawVal()
	return internalTableRegexp.MatchString(rawName)
}
