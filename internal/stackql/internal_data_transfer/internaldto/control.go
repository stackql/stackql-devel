package internaldto

import (
	"github.com/stackql/stackql/pkg/txncounter"
)

var (
	_ TxnControlCounters = &standardTxnControlCounters{}
)

type TxnControlCounters interface {
	GetGenID() int
	GetInsertID() int
	GetSessionID() int
	GetTxnID() int
	SetTableName(string)
	SetTxnID(int)
	Clone() TxnControlCounters
	Copy(TxnControlCounters) TxnControlCounters
	CloneAndIncrementInsertID() TxnControlCounters
}

type standardTxnControlCounters struct {
	genId, sessionID, txnID, insertId int
	tableName                         string
	requestEncoding                   []string
}

func NewTxnControlCounters(txnCtrMgr txncounter.Manager) (TxnControlCounters, error) {
	if txnCtrMgr == nil {
		return &standardTxnControlCounters{}, nil
	}
	genId, err := txnCtrMgr.GetCurrentGenerationID()
	if err != nil {
		return nil, err
	}
	ssnId, err := txnCtrMgr.GetCurrentSessionID()
	if err != nil {
		return nil, err
	}
	txnID, err := txnCtrMgr.GetNextTxnID()
	if err != nil {
		return nil, err
	}
	insertId, err := txnCtrMgr.GetNextInsertID()
	if err != nil {
		return nil, err
	}
	return &standardTxnControlCounters{
		genId:     genId,
		sessionID: ssnId,
		txnID:     txnID,
		insertId:  insertId,
	}, nil
}

func NewTxnControlCountersFromVals(genId, ssnId, txnID, insertId int) TxnControlCounters {
	return &standardTxnControlCounters{
		genId:     genId,
		sessionID: ssnId,
		txnID:     txnID,
		insertId:  insertId,
	}
}

func (tc *standardTxnControlCounters) SetTableName(tn string) {
	tc.tableName = tn
}

func (tc *standardTxnControlCounters) GetGenID() int {
	return tc.genId
}

func (tc *standardTxnControlCounters) GetSessionID() int {
	return tc.sessionID
}

func (tc *standardTxnControlCounters) GetTxnID() int {
	return tc.txnID
}

func (tc *standardTxnControlCounters) GetInsertID() int {
	return tc.insertId
}

func (tc *standardTxnControlCounters) Copy(input TxnControlCounters) TxnControlCounters {
	tc.genId = input.GetGenID()
	tc.insertId = input.GetInsertID()
	tc.sessionID = input.GetSessionID()
	tc.txnID = input.GetTxnID()
	return tc
}

func (tc *standardTxnControlCounters) Clone() TxnControlCounters {
	return &standardTxnControlCounters{
		genId:           tc.genId,
		sessionID:       tc.sessionID,
		txnID:           tc.txnID,
		insertId:        tc.insertId,
		requestEncoding: tc.requestEncoding,
	}
}

func (tc *standardTxnControlCounters) CloneAndIncrementInsertID() TxnControlCounters {
	return &standardTxnControlCounters{
		genId:           tc.genId,
		sessionID:       tc.sessionID,
		txnID:           tc.txnID,
		insertId:        tc.insertId + 1,
		requestEncoding: tc.requestEncoding,
	}
}

func (tc *standardTxnControlCounters) SetTxnID(ti int) {
	tc.txnID = ti
}
