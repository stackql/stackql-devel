package sqldialect

import (
	"fmt"
	"strings"

	"github.com/stackql/stackql/internal/stackql/sqlengine"
)

type SQLDialect interface {
	GCCollect(transactionIDs []int) error
	GCCollectAll() error
}

func NewSQLDialect(sqlEngine sqlengine.SQLEngine, name string) (SQLDialect, error) {
	switch strings.ToLower(name) {
	case "sqlite":
		return newSQLiteDialct(sqlEngine)
	default:
		return nil, fmt.Errorf("cannot accomodate sql dialect '%s'", name)
	}
}

func newSQLiteDialct(sqlEngine sqlengine.SQLEngine) (SQLDialect, error) {
	return &SQLiteDialect{sqlEngine: sqlEngine}, nil
}

type SQLiteDialect struct {
	sqlEngine sqlengine.SQLEngine
}

func (sl *SQLiteDialect) GCCollectAll() error {
	return sl.gcCollectAll()
}

func (sl *SQLiteDialect) GCCollect(transactionIDs []int) error {
	return sl.gcCollect(transactionIDs)
}

func (sl *SQLiteDialect) gcCollectAll() error {
	s, err := sl.getGCCollectAllTemplate()
	if err != nil {
		return err
	}
	_, err = sl.sqlEngine.Exec(s)
	return err
}

func (sl *SQLiteDialect) gcCollect(transactionIDs []int) error {
	s, err := sl.getGCCollectTemplate(transactionIDs)
	if err != nil {
		return err
	}
	_, err = sl.sqlEngine.Exec(s)
	return err
}

func (sl *SQLiteDialect) getGCCollectAllTemplate() (string, error) {
	query := `SELECT DISTINCT table_name FROM "__iql__.control.gc.txn_table_x_ref" ;`
	rows, err := sl.sqlEngine.Query(query)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("begin transaction; ")
	for {
		hasNext := rows.Next()
		if !hasNext {
			break
		}
		var s string
		rows.Scan(&s)
		sb.WriteString(fmt.Sprintf("delete from %s; ", s))
	}
	sb.WriteString("commit; ")
	return sb.String(), nil
}

func (sl *SQLiteDialect) getGCCollectTemplate(transactionIDs []int) (string, error) {
	var transactionIDStrings []string
	for _, txn := range transactionIDs {
		transactionIDStrings = append(transactionIDStrings, fmt.Sprintf("%d", txn))
	}
	var inBuilder strings.Builder
	inBuilder.WriteString("( ")
	inBuilder.WriteString(strings.Join(transactionIDStrings, ", "))
	inBuilder.WriteString(" )")
	query := fmt.Sprintf(`SELECT DISTINCT table_name, iql_transaction_id FROM "__iql__.control.gc.txn_table_x_ref" WHERE iql_transaction_id IN %s ;`, inBuilder.String())
	rows, err := sl.sqlEngine.Query(query)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("begin transaction; ")
	for {
		hasNext := rows.Next()
		if !hasNext {
			break
		}
		var s string
		var i int
		rows.Scan(&s, &i)
		sb.WriteString(fmt.Sprintf(`delete from "%s" where iql_transaction_id  = %d ; `, s, i))
	}
	sb.WriteString("commit; ")
	return sb.String(), nil
}
