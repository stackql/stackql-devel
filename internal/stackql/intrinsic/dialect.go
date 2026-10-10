package intrinsic

import (
	"fmt"

	"github.com/stackql-labs/omnisdk/pkg/sqlfn"
	"github.com/stackql/any-sdk/pkg/constants"
	"github.com/stackql/any-sdk/pkg/dto"
)

// backendDialect is the SQL dialect of the session's backend, which decides what each function a
// query calls means: the embedded SQLite's, or Postgres's. It is read from the backend's
// configuration; a backend omnisdk has no catalogue for is refused.
func backendDialect(ctx queryContext) (sqlfn.Dialect, error) {
	cfg, err := dto.GetSQLBackendCfg(ctx.GetRuntimeContext().SQLBackendCfgRaw)
	if err != nil {
		return "", err
	}
	switch cfg.GetSQLDialect() {
	case constants.SQLDialectSQLite3:
		return sqlfn.SQLite, nil
	case constants.SQLDialectPostgres:
		return sqlfn.Postgres, nil
	}
	return "", fmt.Errorf("SQL backend %q has no omnisdk function catalogue (want %q or %q)",
		cfg.GetSQLDialect(), constants.SQLDialectSQLite3, constants.SQLDialectPostgres)
}
