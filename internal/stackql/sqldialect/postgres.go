package sqldialect

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/stackql/stackql/internal/stackql/constants"
	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/relationaldto"
	"github.com/stackql/stackql/internal/stackql/sqlcontrol"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/tablenamespace"
)

func newPostgresDialect(sqlEngine sqlengine.SQLEngine, namespaces tablenamespace.TableNamespaceCollection, controlAttributes sqlcontrol.ControlAttributes) (SQLDialect, error) {
	rv := &postgresDialect{
		controlAttributes: controlAttributes,
		namespaces:        namespaces,
		sqlEngine:         sqlEngine,
	}
	err := rv.initSQLiteEngine()
	return rv, err
}

type postgresDialect struct {
	controlAttributes sqlcontrol.ControlAttributes
	namespaces        tablenamespace.TableNamespaceCollection
	sqlEngine         sqlengine.SQLEngine
}

func (eng *postgresDialect) initSQLiteEngine() error {
	_, err := eng.sqlEngine.Exec(postgresEngineSetupDDL)
	return err
}

func (eng *postgresDialect) generateDropTableStatement(relationalTable relationaldto.RelationalTable) string {
	return fmt.Sprintf(`drop table if exists "%s"`, relationalTable.GetName())
}

func (eng *postgresDialect) GenerateDDL(relationalTable relationaldto.RelationalTable, dropTable bool) ([]string, error) {
	return eng.generateDDL(relationalTable, dropTable)
}

func (eng *postgresDialect) generateDDL(relationalTable relationaldto.RelationalTable, dropTable bool) ([]string, error) {
	var colDefs, retVal []string
	if dropTable {
		retVal = append(retVal, eng.generateDropTableStatement(relationalTable))
	}
	var rv strings.Builder
	tableName := relationalTable.GetName()
	rv.WriteString(fmt.Sprintf(`create table if not exists "%s" ( `, tableName))
	colDefs = append(colDefs, fmt.Sprintf(`"iql_%s_id" BIGSERIAL PRIMARY KEY`, tableName))
	genIdColName := eng.controlAttributes.GetControlGenIdColumnName()
	sessionIdColName := eng.controlAttributes.GetControlSsnIdColumnName()
	txnIdColName := eng.controlAttributes.GetControlTxnIdColumnName()
	maxTxnIdColName := eng.controlAttributes.GetControlMaxTxnColumnName()
	insIdColName := eng.controlAttributes.GetControlInsIdColumnName()
	lastUpdateColName := eng.controlAttributes.GetControlLatestUpdateColumnName()
	insertEncodedColName := eng.controlAttributes.GetControlInsertEncodedIdColumnName()
	gcStatusColName := eng.controlAttributes.GetControlGCStatusColumnName()
	colDefs = append(colDefs, fmt.Sprintf(`"%s" INTEGER `, genIdColName))
	colDefs = append(colDefs, fmt.Sprintf(`"%s" INTEGER `, sessionIdColName))
	colDefs = append(colDefs, fmt.Sprintf(`"%s" INTEGER `, txnIdColName))
	colDefs = append(colDefs, fmt.Sprintf(`"%s" INTEGER `, maxTxnIdColName))
	colDefs = append(colDefs, fmt.Sprintf(`"%s" INTEGER `, insIdColName))
	colDefs = append(colDefs, fmt.Sprintf(`"%s" TEXT `, insertEncodedColName))
	colDefs = append(colDefs, fmt.Sprintf(`"%s" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP `, lastUpdateColName))
	colDefs = append(colDefs, fmt.Sprintf(`"%s" SMALLINT NOT NULL DEFAULT %d `, gcStatusColName, constants.GCBlack))
	for _, col := range relationalTable.GetColumns() {
		var b strings.Builder
		colName := col.GetName()
		colType := col.GetType()
		b.WriteString(`"` + colName + `" `)
		b.WriteString(colType)
		colDefs = append(colDefs, b.String())
	}
	rv.WriteString(strings.Join(colDefs, " , "))
	rv.WriteString(" ) ")
	retVal = append(retVal, rv.String())
	retVal = append(retVal, fmt.Sprintf(`create index if not exists "idx_%s_%s" on "%s" ( "%s" ) `, strings.ReplaceAll(tableName, ".", "_"), genIdColName, tableName, genIdColName))
	retVal = append(retVal, fmt.Sprintf(`create index if not exists "idx_%s_%s" on "%s" ( "%s" ) `, strings.ReplaceAll(tableName, ".", "_"), sessionIdColName, tableName, sessionIdColName))
	retVal = append(retVal, fmt.Sprintf(`create index if not exists "idx_%s_%s" on "%s" ( "%s" ) `, strings.ReplaceAll(tableName, ".", "_"), txnIdColName, tableName, txnIdColName))
	retVal = append(retVal, fmt.Sprintf(`create index if not exists "idx_%s_%s" on "%s" ( "%s" ) `, strings.ReplaceAll(tableName, ".", "_"), insIdColName, tableName, insIdColName))
	return retVal, nil
}

func (sl *postgresDialect) GCAdd(tableName string, parentTcc, lockableTcc dto.TxnControlCounters) error {
	maxTxnColName := sl.controlAttributes.GetControlMaxTxnColumnName()
	q := fmt.Sprintf(
		`
		UPDATE "%s" 
		SET "%s" = r.current_value
		FROM (
			SELECT *
			FROM
				"__iql__.control.gc.rings"
		) AS r
		WHERE 
			"%s" = $1 
			AND 
			"%s" = $2 
			AND
			r.ring_name = 'transaction_id'
			AND
			"%s" < CASE 
			   WHEN ("%s" - r.current_offset) < 0
				 THEN CAST(pow(2, r.width_bits) + ("%s" - r.current_offset)  AS int)
				 ELSE "%s" - r.current_offset
				 END
		`,
		tableName,
		maxTxnColName,
		sl.controlAttributes.GetControlTxnIdColumnName(),
		sl.controlAttributes.GetControlInsIdColumnName(),
		maxTxnColName,
		maxTxnColName,
		maxTxnColName,
		maxTxnColName,
	)
	_, err := sl.sqlEngine.Exec(q, lockableTcc.TxnId, lockableTcc.InsertId)
	return err
}

func (sl *postgresDialect) GCCollectObsoleted(minTransactionID int) error {
	return sl.gCCollectObsoleted(minTransactionID)
}

func (sl *postgresDialect) gCCollectObsoleted(minTransactionID int) error {
	maxTxnColName := sl.controlAttributes.GetControlMaxTxnColumnName()
	obtainQuery := fmt.Sprintf(
		`
		SELECT
			'DELETE FROM "' || table_name || '" WHERE "%s" < %d ; '
		from 
			information_schema.tables 
		where 
			table_type = 'BASE TABLE' 
			and 
			table_catalog = $1
			and 
			table_schema = $2
		  and
			table_name not like '__iql__%%'
		`,
		maxTxnColName,
		minTransactionID,
	)
	deleteQueryResultSet, err := sl.sqlEngine.Query(obtainQuery, sl.sqlEngine.GetTableCatalog(), sl.sqlEngine.GetTableSchema())
	if err != nil {
		return err
	}
	return sl.readExecGeneratedQueries(deleteQueryResultSet)
}

func (sl *postgresDialect) GCCollectAll() error {
	return sl.gCCollectAll()
}

func (sl *postgresDialect) gCCollectAll() error {
	obtainQuery := `
		SELECT
			'DELETE FROM "' || table_name || '"  ; '
		from 
			information_schema.tables 
		where 
			table_type = 'BASE TABLE' 
			and 
			table_catalog = $1
			and 
			table_schema = $2
		  and
			table_name not like '__iql__%%'
		`
	deleteQueryResultSet, err := sl.sqlEngine.Query(obtainQuery, sl.sqlEngine.GetTableCatalog(), sl.sqlEngine.GetTableSchema())
	if err != nil {
		return err
	}
	return sl.readExecGeneratedQueries(deleteQueryResultSet)
}

func (sl *postgresDialect) GCControlTablesPurge() error {
	return sl.gcControlTablesPurge()
}

func (sl *postgresDialect) gcControlTablesPurge() error {
	obtainQuery := `
		SELECT
		  'DELETE FROM "' || table_name || '" ; '
			from 
			information_schema.tables 
		where 
			table_type = 'BASE TABLE' 
			and 
			table_catalog = $1
			and 
			table_schema = $2
		  and
			table_name like '__iql__%%'
		`
	deleteQueryResultSet, err := sl.sqlEngine.Query(obtainQuery, sl.sqlEngine.GetTableCatalog(), sl.sqlEngine.GetTableSchema())
	if err != nil {
		return err
	}
	return sl.readExecGeneratedQueries(deleteQueryResultSet)
}

func (sl *postgresDialect) GCPurgeEphemeral() error {
	return sl.gcPurgeEphemeral()
}

func (sl *postgresDialect) GCPurgeCache() error {
	return sl.gcPurgeCache()
}

func (sl *postgresDialect) gcPurgeCache() error {
	query := `
	select distinct 
		'DROP TABLE IF EXISTS "' || table_name || '" ; ' 
	from 
		information_schema.tables 
	where 
		table_type = 'BASE TABLE' 
		and 
		table_catalog = $1
		and 
		table_schema = $2
		and 
		table_name like $3
	`
	rows, err := sl.sqlEngine.Query(query, sl.sqlEngine.GetTableCatalog(), sl.sqlEngine.GetTableSchema(), sl.namespaces.GetAnalyticsCacheTableNamespaceConfigurator().GetLikeString())
	if err != nil {
		return err
	}
	return sl.readExecGeneratedQueries(rows)
}

func (sl *postgresDialect) gcPurgeEphemeral() error {
	query := `
	select distinct 
		'DROP TABLE IF EXISTS "' || table_name || '" ; ' 
	from 
		information_schema.tables 
	where 
		table_type = 'BASE TABLE' 
		and 
		table_catalog = $1
		and 
		table_schema = $2
		and 
		table_name NOT like $3
		and 
		table_name not like '__iql__%' 
	`
	rows, err := sl.sqlEngine.Query(query, sl.sqlEngine.GetTableCatalog(), sl.sqlEngine.GetTableSchema(), sl.namespaces.GetAnalyticsCacheTableNamespaceConfigurator().GetLikeString())
	if err != nil {
		return err
	}
	return sl.readExecGeneratedQueries(rows)
}

func (sl *postgresDialect) PurgeAll() error {
	return sl.purgeAll()
}

func (sl *postgresDialect) GetSQLEngine() sqlengine.SQLEngine {
	return sl.sqlEngine
}

func (sl *postgresDialect) purgeAll() error {
	obtainQuery := `
		SELECT
			'DROP TABLE IF EXISTS "' || table_name || '" ; '
		from 
			information_schema.tables 
		where 
			table_type = 'BASE TABLE' 
			and 
			table_catalog = $1 
			and 
			table_schema = $2
		  AND
			table_name NOT LIKE '__iql__%'
		`
	deleteQueryResultSet, err := sl.sqlEngine.Query(obtainQuery, sl.sqlEngine.GetTableCatalog(), sl.sqlEngine.GetTableSchema())
	if err != nil {
		return err
	}
	return sl.readExecGeneratedQueries(deleteQueryResultSet)
}

func (sl *postgresDialect) readExecGeneratedQueries(queryResultSet *sql.Rows) error {
	defer queryResultSet.Close()
	var queries []string
	for {
		hasNext := queryResultSet.Next()
		if !hasNext {
			break
		}
		var s string
		err := queryResultSet.Scan(&s)
		if err != nil {
			return err
		}
		queries = append(queries, s)
	}
	err := sl.sqlEngine.ExecInTxn(queries)
	return err
}
