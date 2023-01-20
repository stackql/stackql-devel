package sql_dialect

import "fmt"

type SQLDialect interface {
	GetTableMetadataQuery(tableName string) string
}

type postgresSQLDialect struct {
	//
}

func (sd *postgresSQLDialect) GetTableMetadataQuery(schemaName, tableName string) string {
	return fmt.Sprintf(`SELECT `)
}
