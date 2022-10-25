package garbagecollector

import (
	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/gcexec"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/tableinsertioncontainer"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
)

type GarbageCollector interface {
	Update(string, dto.TxnControlCounters, dto.TxnControlCounters) error
	AddInsertContainer(tm *tablemetadata.ExtendedTableMetadata) tableinsertioncontainer.TableInsertionContainer
	Close() error
	Collect() error
}

func NewGarbageCollector(gcExecutor gcexec.GarbageCollectorExecutor, gcCfg dto.GCCfg, sqlEngine sqlengine.SQLEngine) GarbageCollector {
	return newStandardGarbageCollector(gcExecutor, gcCfg, sqlEngine)
}

func newStandardGarbageCollector(gcExecutor gcexec.GarbageCollectorExecutor, policy dto.GCCfg, sqlEngine sqlengine.SQLEngine) GarbageCollector {
	return &standardGarbageCollector{
		gcExecutor: gcExecutor,
		isEager:    policy.IsEager,
		sqlEngine:  sqlEngine,
	}
}

type standardGarbageCollector struct {
	gcExecutor       gcexec.GarbageCollectorExecutor
	insertContainers []tableinsertioncontainer.TableInsertionContainer
	isEager          bool
	sqlEngine        sqlengine.SQLEngine
}

func (gc *standardGarbageCollector) Update(tableName string, parentTcc, tcc dto.TxnControlCounters) error {
	return gc.gcExecutor.Update(tableName, parentTcc, tcc)
}

func (gc *standardGarbageCollector) Close() error {
	for _, ic := range gc.insertContainers {
		a, b := ic.GetTableTxnCounters()
		gc.gcExecutor.Condemn(a, *b)
	}
	if gc.isEager {
		return gc.gcExecutor.Collect()
	}
	return nil
}

func (gc *standardGarbageCollector) Collect() error {
	return gc.gcExecutor.Collect()
}

func (gc *standardGarbageCollector) AddInsertContainer(tm *tablemetadata.ExtendedTableMetadata) tableinsertioncontainer.TableInsertionContainer {
	rv := tableinsertioncontainer.NewTableInsertionContainer(tm, gc.sqlEngine)
	gc.insertContainers = append(gc.insertContainers, rv)
	return rv
}

func (gc *standardGarbageCollector) GetGarbageCollectorExecutor() gcexec.GarbageCollectorExecutor {
	return gc.gcExecutor
}

func (gc *standardGarbageCollector) GetInsertContainers() []tableinsertioncontainer.TableInsertionContainer {
	return gc.insertContainers
}
