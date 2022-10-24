package txncounter

import (
	"sync"
)

var (
	genCtrlMutex        *sync.Mutex = &sync.Mutex{}
	txnCtrlMutex        *sync.Mutex = &sync.Mutex{}
	currentTxnId        *int        = new(int)
	currentGenerationId *int        = new(int)
)

func GetNextGenerationId() int {
	genCtrlMutex.Lock()
	defer genCtrlMutex.Unlock()
	*currentGenerationId++
	return *currentGenerationId
}

type TxnCounterManager interface {
	GetCurrentGenerationId() int
	GetCurrentSessionId() int
	GetNextInsertId() int
	GetNextTxnId() int
}

type standardTxnCounterManager struct {
	perTxnMutex     *sync.Mutex
	generationId    int
	sessionId       int
	currentInsertId int
}

func NewTxnCounterManager(generationId, sessionId int) TxnCounterManager {
	return &standardTxnCounterManager{
		generationId: generationId,
		sessionId:    sessionId,
		perTxnMutex:  &sync.Mutex{},
	}
}

func (tc *standardTxnCounterManager) GetCurrentGenerationId() int {
	return tc.generationId
}

func (tc *standardTxnCounterManager) GetCurrentSessionId() int {
	return tc.sessionId
}

func (tc *standardTxnCounterManager) GetNextTxnId() int {
	txnCtrlMutex.Lock()
	defer txnCtrlMutex.Unlock()
	*currentTxnId++
	return *currentTxnId
}

func (tc *standardTxnCounterManager) GetNextInsertId() int {
	tc.perTxnMutex.Lock()
	defer tc.perTxnMutex.Unlock()
	tc.currentInsertId++
	return tc.currentInsertId
}
