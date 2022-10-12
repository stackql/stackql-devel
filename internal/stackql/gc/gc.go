package gc

import (
	"sync"

	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/sqldialect"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
)

type TxnMap map[int]struct{}

type BrutalGarbageCollector interface {
	CollectAll() error
}

type AbstractFlatGarbageCollector interface {
	Add(*dto.TxnControlCounters) bool
	Condemn(tcc *dto.TxnControlCounters) bool
	Collect() error
}

type GarbageCollector interface {
	BrutalGarbageCollector
	AbstractFlatGarbageCollector
}

func NewGarbageCollector(sqlEngine sqlengine.SQLEngine, dialectStr string) (GarbageCollector, error) {
	dialect, err := sqldialect.NewSQLDialect(sqlEngine, dialectStr)
	if err != nil {
		return nil, err
	}
	return newBasicGarbageCollector(dialect)
}

func newBasicGarbageCollector(dialect sqldialect.SQLDialect) (GarbageCollector, error) {
	return &BasicGarbageCollector{
		activeTxns: make(TxnMap),
		gcMutex:    &sync.Mutex{},
		sqlDialect: dialect,
	}, nil
}

// Algorithm summary:
//   - `Collect()` will reclaim resources from all txns **not** in supplied list of IDs.
//   - `CollectAll()` as assumed.
type BasicGarbageCollector struct {
	activeTxns TxnMap
	gcMutex    *sync.Mutex
	sqlDialect sqldialect.SQLDialect
}

func (rc *BasicGarbageCollector) Add(tcc *dto.TxnControlCounters) bool {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	rc.activeTxns[tcc.TxnId] = struct{}{}
	return true
}

func (rc *BasicGarbageCollector) Condemn(tcc *dto.TxnControlCounters) bool {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	delete(rc.activeTxns, tcc.TxnId)
	return true
}

// Algorithm, **must be done during pause**:
//   - Assemble active transactions.
//   - Retrieve GC queries from control table.
//   - Execute GC queries in a txn.
func (rc *BasicGarbageCollector) Collect() error {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	var activeTxnsToPreserve []int
	for k, _ := range rc.activeTxns {
		activeTxnsToPreserve = append(activeTxnsToPreserve, k)
	}
	return rc.sqlDialect.GCCollect(activeTxnsToPreserve)
}

// Algorithm, **must be done during pause**:
//   - Execute **all possible** GC queries in a txn.
func (rc *BasicGarbageCollector) CollectAll() error {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	return rc.sqlDialect.GCCollectAll()
}
