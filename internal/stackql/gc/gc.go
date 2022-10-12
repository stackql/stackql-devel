package gc

import (
	"sync"

	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/sqldialect"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/tablenamespace"
)

type TxnMap struct {
	mutex *sync.Mutex
	m     map[int]int
}

func NewTxnMap() TxnMap {
	return TxnMap{
		mutex: &sync.Mutex{},
		m:     make(map[int]int),
	}
}

func (tm TxnMap) GetTxnIDs() []int {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	var rv []int
	for k, v := range tm.m {
		if v > 0 {
			rv = append(rv, k)
		}
	}
	return rv
}

func (tm TxnMap) Add(tcc *dto.TxnControlCounters) int {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	key := tcc.TxnId
	existingVal, ok := tm.m[key]
	if ok {
		tm.m[key] = existingVal + 1
		return existingVal + 1
	}
	tm.m[key] = 1
	return 1
}

func (tm TxnMap) Delete(tcc *dto.TxnControlCounters) int {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	key := tcc.TxnId
	existingVal, ok := tm.m[key]
	if ok {
		newVal := existingVal - 1
		if newVal == 0 {
			delete(tm.m, key)
			return 0
		}
		tm.m[key] = newVal
		return newVal
	}
	return 0
}

type BrutalGarbageCollector interface {
	CollectAll() error
}

type AbstractFlatGarbageCollector interface {
	Add(string, *dto.TxnControlCounters) bool
	Condemn(string, *dto.TxnControlCounters) bool
	Collect() error
}

type GarbageCollector interface {
	BrutalGarbageCollector
	AbstractFlatGarbageCollector
}

func NewGarbageCollector(sqlEngine sqlengine.SQLEngine, ns tablenamespace.TableNamespaceCollection, dialectStr string) (GarbageCollector, error) {
	dialect, err := sqldialect.NewSQLDialect(sqlEngine, ns, dialectStr)
	if err != nil {
		return nil, err
	}
	return newBasicGarbageCollector(dialect, ns)
}

func newBasicGarbageCollector(dialect sqldialect.SQLDialect, ns tablenamespace.TableNamespaceCollection) (GarbageCollector, error) {
	return &BasicGarbageCollector{
		activeTxns:      NewTxnMap(),
		activeTxnsCache: NewTxnMap(),
		gcMutex:         &sync.Mutex{},
		ns:              ns,
		sqlDialect:      dialect,
	}, nil
}

// Algorithm summary:
//   - `Collect()` will reclaim resources from all txns **not** in supplied list of IDs.
//   - `CollectAll()` as assumed.
type BasicGarbageCollector struct {
	activeTxns      TxnMap
	activeTxnsCache TxnMap
	gcMutex         *sync.Mutex
	ns              tablenamespace.TableNamespaceCollection
	sqlDialect      sqldialect.SQLDialect
}

func (rc *BasicGarbageCollector) Add(tableName string, tcc *dto.TxnControlCounters) bool {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	if rc.ns.GetAnalyticsCacheTableNamespaceConfigurator().IsAllowed(tableName) {
		rc.activeTxnsCache.Add(tcc)
	}
	rc.activeTxns.Add(tcc)
	return true
}

func (rc *BasicGarbageCollector) Condemn(tableName string, tcc *dto.TxnControlCounters) bool {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	if rc.ns.GetAnalyticsCacheTableNamespaceConfigurator().IsAllowed(tableName) {
		rc.activeTxnsCache.Delete(tcc)
		return true
	}
	rc.activeTxns.Delete(tcc)
	return true
}

// Algorithm, **must be done during pause**:
//   - Assemble active transactions.
//   - Retrieve GC queries from control table.
//   - Execute GC queries in a txn.
func (rc *BasicGarbageCollector) Collect() error {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	activeTxnIDs := rc.activeTxns.GetTxnIDs()
	return rc.sqlDialect.GCCollect(activeTxnIDs)
}

// Algorithm, **must be done during pause**:
//   - Execute **all possible** GC queries in a txn.
func (rc *BasicGarbageCollector) CollectAll() error {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	return rc.sqlDialect.GCCollectAll()
}
