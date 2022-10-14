package garbagecollector

import (
	"github.com/stackql/stackql/internal/stackql/gcexec"
	"github.com/stackql/stackql/internal/stackql/tableinsertioncontainer"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
)

type GarbageCollector interface {
	AddInsertContainer(tm *tablemetadata.ExtendedTableMetadata) tableinsertioncontainer.TableInsertionContainer
	Close() error
}

func NewGarbageCollector(gcExecutor gcexec.GarbageCollectorExecutor) GarbageCollector {
	return newStandardGarbageCollector(gcExecutor)
}

func newStandardGarbageCollector(gcExecutor gcexec.GarbageCollectorExecutor) GarbageCollector {
	return &StandardGarbageCollector{
		gcExecutor: gcExecutor,
	}
}

type StandardGarbageCollector struct {
	gcExecutor       gcexec.GarbageCollectorExecutor
	insertContainers []tableinsertioncontainer.TableInsertionContainer
}

func (gc *StandardGarbageCollector) Close() error {
	for _, ic := range gc.insertContainers {
		gc.gcExecutor.Condemn(ic.GetTableTxnCounters())
	}
	return nil
}

func (gc *StandardGarbageCollector) AddInsertContainer(tm *tablemetadata.ExtendedTableMetadata) tableinsertioncontainer.TableInsertionContainer {
	rv := tableinsertioncontainer.NewTableInsertionContainer(tm)
	gc.insertContainers = append(gc.insertContainers, rv)
	return rv
}

func (gc *StandardGarbageCollector) GetGarbageCollectorExecutor() gcexec.GarbageCollectorExecutor {
	return gc.gcExecutor
}

func (gc *StandardGarbageCollector) GetInsertContainers() []tableinsertioncontainer.TableInsertionContainer {
	return gc.insertContainers
}
