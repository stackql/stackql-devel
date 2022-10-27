package sqldialect

import (
	"fmt"
	"strings"

	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/sqlcontrol"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/tablenamespace"
)

type SQLDialect interface {
	// GCAdd() will record a Txn as active
	GCAdd(string, dto.TxnControlCounters, dto.TxnControlCounters) error
	// GCCollectObsoleted() must be mutex-protected.
	GCCollectObsoleted(minTransactionID int) error
	// GCPurgeCache() will completely wipe the cache.
	GCPurgeCache() error
	// PurgeAll() drops all data tables, does **not** drop control tables.
	PurgeAll() error
}

func NewSQLDialect(sqlEngine sqlengine.SQLEngine, namespaces tablenamespace.TableNamespaceCollection, controlAttributes sqlcontrol.ControlAttributes, name string) (SQLDialect, error) {
	switch strings.ToLower(name) {
	case "sqlite":
		return newSQLiteDialct(sqlEngine, namespaces, controlAttributes)
	default:
		return nil, fmt.Errorf("cannot accomodate sql dialect '%s'", name)
	}
}

func newSQLiteDialct(sqlEngine sqlengine.SQLEngine, namespaces tablenamespace.TableNamespaceCollection, controlAttributes sqlcontrol.ControlAttributes) (SQLDialect, error) {
	rv := &sqLiteDialect{
		controlAttributes: controlAttributes,
		namespaces:        namespaces,
		sqlEngine:         sqlEngine,
	}
	err := rv.initSQLiteEngine()
	return rv, err
}

type sqLiteDialect struct {
	controlAttributes sqlcontrol.ControlAttributes
	namespaces        tablenamespace.TableNamespaceCollection
	sqlEngine         sqlengine.SQLEngine
}

func (eng *sqLiteDialect) initSQLiteEngine() error {
	_, err := eng.sqlEngine.Exec(sqlEngineSetupDDL)
	return err
}

func (sl *sqLiteDialect) GCAdd(tableName string, parentTcc, lockableTcc dto.TxnControlCounters) error {
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
			"%s" = ? 
			AND 
			"%s" = ? 
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

func (sl *sqLiteDialect) GCCollectObsoleted(minTransactionID int) error {
	return sl.gCCollectObsoleted(minTransactionID)
}

func (sl *sqLiteDialect) gCCollectObsoleted(minTransactionID int) error {
	maxTxnColName := sl.controlAttributes.GetControlMaxTxnColumnName()
	obtainQuery := fmt.Sprintf(
		`
		SELECT
			group_concat(
				'DELETE FROM "' || name | '" WHERE "%s" < %d ; ',
				' ' 
			)
		FROM
			sqlite_master 
		where 
			type = 'table'
		`,
		maxTxnColName,
		minTransactionID,
	)
	deleteQueryResultSet, err := sl.sqlEngine.Query(obtainQuery)
	if err != nil {
		return err
	}
	hasNext := deleteQueryResultSet.Next()
	if !hasNext {
		return fmt.Errorf("purgeAll() failed: query generation lacking result")
	}
	var deleteQueries string
	err = deleteQueryResultSet.Scan(&deleteQueries)
	if err != nil {
		return err
	}
	q := fmt.Sprintf(
		`BEGIN; %s COMMIT; `,
		deleteQueries,
	)
	_, err = sl.sqlEngine.Exec(q)
	return err
}

func (sl *sqLiteDialect) GCPurgeCache() error {
	return sl.gcPurgeCache()
}

func (sl *sqLiteDialect) gcPurgeCache() error {
	s, err := sl.getGCPurgeCacheTemplate()
	if err != nil {
		return err
	}
	err = sl.sqlEngine.ExecInTxn(s)
	return err
}

func (sl *sqLiteDialect) getGCPurgeCacheTemplate() ([]string, error) {
	query := `select distinct name from sqlite_schema where type = 'table' and name like ? ;`
	rows, err := sl.sqlEngine.Query(query, sl.namespaces.GetAnalyticsCacheTableNamespaceConfigurator().GetLikeString())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rv []string
	for {
		hasNext := rows.Next()
		if !hasNext {
			break
		}
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		rv = append(rv, fmt.Sprintf(`DROP TABLE IF EXISTS "%s" CASCADE; `, s))
	}
	return rv, nil
}

func (sl *sqLiteDialect) PurgeAll() error {
	return sl.purgeAll()
}

func (sl *sqLiteDialect) purgeAll() error {
	obtainQuery := `
		SELECT
			group_concat(
				'DROP TABLE IF EXISTS "' || name || '" ; ',
				' ' 
			)
		FROM
			sqlite_master 
		where 
			type = 'table'
		  AND
			name NOT LIKE '__iql__%'
		`
	deleteQueryResultSet, err := sl.sqlEngine.Query(obtainQuery)
	if err != nil {
		return err
	}
	hasNext := deleteQueryResultSet.Next()
	if !hasNext {
		return fmt.Errorf("purgeAll() failed: query generation lacking result")
	}
	var deleteQueries string
	err = deleteQueryResultSet.Scan(&deleteQueries)
	if err != nil {
		return err
	}
	q := fmt.Sprintf(
		`BEGIN; %s COMMIT; `,
		deleteQueries,
	)
	_, err = sl.sqlEngine.Exec(q)
	return err
}
