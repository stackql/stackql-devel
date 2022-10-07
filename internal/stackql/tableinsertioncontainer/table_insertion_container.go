package tableinsertioncontainer

import (
	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/taxonomy"
)

var (
	_ TableInsertionContainer = &StandardTableInsertionContainer{}
)

type TableInsertionContainer interface {
	GetTableMetadata() *taxonomy.ExtendedTableMetadata
	IsCountersSet() bool
	SetTxnCounters(*dto.TxnControlCounters)
	GetTxnCounters() *dto.TxnControlCounters
}

type StandardTableInsertionContainer struct {
	tm            *taxonomy.ExtendedTableMetadata
	tcc           *dto.TxnControlCounters
	isCountersSet bool
}

func (ic *StandardTableInsertionContainer) GetTableMetadata() *taxonomy.ExtendedTableMetadata {
	return ic.tm
}

func (ic *StandardTableInsertionContainer) SetTxnCounters(tcc *dto.TxnControlCounters) {
	ic.tcc.GenId = tcc.GenId
	ic.tcc.SessionId = tcc.SessionId
	ic.tcc.InsertId = tcc.InsertId
	ic.tcc.TxnId = tcc.TxnId
	ic.isCountersSet = true
}

func (ic *StandardTableInsertionContainer) GetTxnCounters() *dto.TxnControlCounters {
	return ic.tcc
}

func (ic *StandardTableInsertionContainer) IsCountersSet() bool {
	return ic.isCountersSet
}

func NewTableInsertionContainer(tm *taxonomy.ExtendedTableMetadata) TableInsertionContainer {
	return &StandardTableInsertionContainer{
		tm:  tm,
		tcc: &dto.TxnControlCounters{},
	}
}

func NewTableInsertionContainers(tms []*taxonomy.ExtendedTableMetadata) []TableInsertionContainer {
	var rv []TableInsertionContainer
	for _, tm := range tms {
		rv = append(rv, NewTableInsertionContainer(tm))
	}
	return rv
}
