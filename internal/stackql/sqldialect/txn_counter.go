package sqldialect

// import (
// 	"sync"

// 	"github.com/stackql/stackql/pkg/txncounter"
// )

// func GetTxnCounter() txncounter.TxnCounterManager {
// 	return nil
// }

// type standardSQLTxnCounterManager struct {
// 	perTxnMutex     *sync.Mutex
// 	generationId    int
// 	sessionId       int
// 	currentInsertId int
// }

// func NewTxnCounterManager(generationId, sessionId int) txncounter.TxnCounterManager {
// 	return &standardSQLTxnCounterManager{
// 		generationId: generationId,
// 		sessionId:    sessionId,
// 		perTxnMutex:  &sync.Mutex{},
// 	}
// }

// func (tc *standardSQLTxnCounterManager) GetCurrentGenerationId() int {
// 	return tc.generationId
// }

// func (tc *standardSQLTxnCounterManager) GetCurrentSessionId() int {
// 	return tc.sessionId
// }

// func (tc *standardSQLTxnCounterManager) GetNextTxnId() int {
// 	q := `
// 	SELECT
// 		r.current_value
// 	FROM
// 		"__iql__.control.gc.rings"
// 	WHERE
// 		r.ring_name = 'transaction_id'
// 	`
// 	txnCtrlMutex.Lock()
// 	defer txnCtrlMutex.Unlock()
// 	*currentTxnId++
// 	return *currentTxnId
// }

// func (tc *standardSQLTxnCounterManager) GetNextInsertId() int {
// 	tc.perTxnMutex.Lock()
// 	defer tc.perTxnMutex.Unlock()
// 	tc.currentInsertId++
// 	return tc.currentInsertId
// }
