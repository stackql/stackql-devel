package sqldialect

import (
	"fmt"
	"strings"

	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/tablenamespace"
)

type SQLDialect interface {
	GCCollect(transactionIDs []int) error
	GCCollectFromCache(transactionIDs []int) error
	GCCollectAll() error
}

func NewSQLDialect(sqlEngine sqlengine.SQLEngine, namespaces tablenamespace.TableNamespaceCollection, name string) (SQLDialect, error) {
	switch strings.ToLower(name) {
	case "sqlite":
		return newSQLiteDialct(sqlEngine, namespaces)
	default:
		return nil, fmt.Errorf("cannot accomodate sql dialect '%s'", name)
	}
}

func newSQLiteDialct(sqlEngine sqlengine.SQLEngine, namespaces tablenamespace.TableNamespaceCollection) (SQLDialect, error) {
	return &SQLiteDialect{
		namespaces: namespaces,
		sqlEngine:  sqlEngine,
	}, nil
}

type SQLiteDialect struct {
	namespaces tablenamespace.TableNamespaceCollection
	sqlEngine  sqlengine.SQLEngine
}

func (sl *SQLiteDialect) GCCollectAll() error {
	return sl.gcCollectAll()
}

func (sl *SQLiteDialect) GCCollect(transactionIDs []int) error {
	return sl.gcCollect(transactionIDs)
}

func (sl *SQLiteDialect) GCCollectFromCache(transactionIDs []int) error {
	return sl.gcCollectFromCache(transactionIDs)
}

func (sl *SQLiteDialect) gcCollectAll() error {
	s, err := sl.getGCCollectAllTemplate()
	if err != nil {
		return err
	}
	err = sl.sqlEngine.ExecInTxn(s)
	return err
}

func (sl *SQLiteDialect) gcCollect(transactionIDs []int) error {
	s, err := sl.getGCCollectTemplate(transactionIDs)
	if err != nil {
		return err
	}
	err = sl.sqlEngine.ExecInTxn(s)
	return err
}

func (sl *SQLiteDialect) gcCollectFromCache(transactionIDs []int) error {
	s, err := sl.getGCCollectCacheTemplate(transactionIDs)
	if err != nil {
		return err
	}
	err = sl.sqlEngine.ExecInTxn(s)
	return err
}

func (sl *SQLiteDialect) gcWipeCache() error {
	query := `drop table `
	_, err := sl.sqlEngine.Exec(query)
	return err
}

func (sl *SQLiteDialect) getGCCollectAllTemplate() ([]string, error) {
	query := `SELECT DISTINCT table_name FROM "__iql__.control.gc.txn_table_x_ref" ;`
	rows, err := sl.sqlEngine.Query(query)
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
		rv = append(rv, fmt.Sprintf("delete from %s; ", s))
	}
	return rv, nil
}

func (sl *SQLiteDialect) getGCCollectTemplate(transactionIDs []int) ([]string, error) {
	var transactionIDStrings []string
	for _, txn := range transactionIDs {
		transactionIDStrings = append(transactionIDStrings, fmt.Sprintf("%d", txn))
	}
	var inBuilder strings.Builder
	inBuilder.WriteString("( ")
	inBuilder.WriteString(strings.Join(transactionIDStrings, ", "))
	inBuilder.WriteString(" )")
	query := fmt.Sprintf(`SELECT DISTINCT table_name, iql_transaction_id FROM "__iql__.control.gc.txn_table_x_ref" WHERE iql_transaction_id NOT IN %s ;`, inBuilder.String())
	rows, err := sl.sqlEngine.Query(query)
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
		var i int
		err = rows.Scan(&s, &i)
		if err != nil {
			return nil, err
		}
		rv = append(rv, fmt.Sprintf(`delete from "%s" where iql_transaction_id  = %d ; `, s, i))
	}
	return rv, nil
}

func (sl *SQLiteDialect) getGCCollectCacheTemplate(transactionIDs []int) ([]string, error) {
	var transactionIDStrings []string
	for _, txn := range transactionIDs {
		transactionIDStrings = append(transactionIDStrings, fmt.Sprintf("%d", txn))
	}
	var inBuilder strings.Builder
	inBuilder.WriteString("( ")
	inBuilder.WriteString(strings.Join(transactionIDStrings, ", "))
	inBuilder.WriteString(" )")
	query := fmt.Sprintf(`SELECT DISTINCT table_name, iql_transaction_id FROM "__iql__.control.gc.txn_table_x_ref" WHERE iql_transaction_id NOT IN %s ;`, inBuilder.String())
	rows, err := sl.sqlEngine.Query(query)
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
		var i int
		err = rows.Scan(&s, &i)
		if err != nil {
			return nil, err
		}
		if sl.namespaces.GetAnalyticsCacheTableNamespaceConfigurator().IsAllowed(s) {
			rv = append(rv, fmt.Sprintf(`delete from "%s" where iql_transaction_id  = %d and iql_latest_update <= datetime('now','-%d second'); `, s, i, sl.namespaces.GetAnalyticsCacheTableNamespaceConfigurator().GetTTL()))
		}
	}
	return rv, nil
}
