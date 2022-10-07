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
	SetTxnCounters(*dto.TxnControlCounters)
	GetTxnCounters() *dto.TxnControlCounters
}

type StandardTableInsertionContainer struct {
	tm  *taxonomy.ExtendedTableMetadata
	tcc *dto.TxnControlCounters
}

func (ic *StandardTableInsertionContainer) GetTableMetadata() *taxonomy.ExtendedTableMetadata {
	return ic.tm
}

func (ic *StandardTableInsertionContainer) SetTxnCounters(tcc *dto.TxnControlCounters) {
	ic.tcc = tcc
}

func (ic *StandardTableInsertionContainer) GetTxnCounters() *dto.TxnControlCounters {
	return ic.tcc
}

func NewTableInsertionContainer(tm *taxonomy.ExtendedTableMetadata) TableInsertionContainer {
	return &StandardTableInsertionContainer{tm: tm}
}

func NewTableInsertionContainers(tms []*taxonomy.ExtendedTableMetadata) []TableInsertionContainer {
	var rv []TableInsertionContainer
	for _, tm := range tms {
		rv = append(rv, NewTableInsertionContainer(tm))
	}
	return rv
}
