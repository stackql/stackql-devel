package typing

import (
	"database/sql"
	"reflect"

	"github.com/lib/pq/oid"
	"github.com/stackql/psql-wire/pkg/sqldata"
)

type Config interface {
	GetGolangKind(discoType string) reflect.Kind
	GetGolangValue(discoType string) interface{}
	GetRelationalType(discoType string) string
	GetOidForSQLType(colType *sql.ColumnType) oid.Oid
	GetPlaceholderColumn(
		table sqldata.ISQLTable, colName string, colOID oid.Oid) sqldata.ISQLColumn
	GetPlaceholderColumnForNativeResult(
		table sqldata.ISQLTable,
		colName string, colSchema *sql.ColumnType) sqldata.ISQLColumn
	GetDefaultOID() oid.Oid
}

func NewTypingConfig(sqlDialect string) (Config, error) {
	return newTypingConfig(sqlDialect)
}
