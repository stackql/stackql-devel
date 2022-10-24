package dto

import (
	"github.com/stackql/stackql/pkg/txncounter"
)

type TxnControlCounters struct {
	GenId, SessionId, TxnId, InsertId, DiscoveryGenerationId int
	TableName                                                string
	RequestEncoding                                          []string
}

func NewTxnControlCounters(txnCtrMgr txncounter.TxnCounterManager, discoveryGenerationID int) (*TxnControlCounters, error) {
	genId, err := txnCtrMgr.GetCurrentGenerationId()
	if err != nil {
		return nil, err
	}
	ssnId, err := txnCtrMgr.GetCurrentSessionId()
	if err != nil {
		return nil, err
	}
	txnId, err := txnCtrMgr.GetNextTxnId()
	if err != nil {
		return nil, err
	}
	insertId, err := txnCtrMgr.GetNextInsertId()
	if err != nil {
		return nil, err
	}
	return &TxnControlCounters{
		GenId:                 genId,
		SessionId:             ssnId,
		TxnId:                 txnId,
		InsertId:              insertId,
		DiscoveryGenerationId: discoveryGenerationID,
	}, nil
}

func (tc *TxnControlCounters) CloneWithDiscoGenID(discoveryGenerationID int) *TxnControlCounters {
	return &TxnControlCounters{
		GenId:                 tc.GenId,
		SessionId:             tc.SessionId,
		TxnId:                 tc.TxnId,
		InsertId:              tc.InsertId,
		RequestEncoding:       tc.RequestEncoding,
		DiscoveryGenerationId: discoveryGenerationID,
	}
}

func (tc *TxnControlCounters) CloneAndIncrementInsertID() *TxnControlCounters {
	return &TxnControlCounters{
		GenId:                 tc.GenId,
		SessionId:             tc.SessionId,
		TxnId:                 tc.TxnId,
		InsertId:              tc.InsertId + 1,
		RequestEncoding:       tc.RequestEncoding,
		DiscoveryGenerationId: tc.DiscoveryGenerationId,
	}
}

func (tc *TxnControlCounters) SetTxnId(ti int) {
	tc.TxnId = ti
}
