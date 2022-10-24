package gcexec

import (
	"sync"

	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/sqldialect"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/tablenamespace"
)

var (
	once                     sync.Once
	garbageCollectorExecutor GarbageCollectorExecutor
)

type TxnMap interface {
	Add(tcc *dto.TxnControlCounters) int
	Delete(tcc *dto.TxnControlCounters) int
	GetTxnIDs() []int
}

type basicTxnMap struct {
	mutex *sync.Mutex
	m     map[int]int
}

func newTxnMap() TxnMap {
	return basicTxnMap{
		mutex: &sync.Mutex{},
		m:     make(map[int]int),
	}
}

func (tm basicTxnMap) GetTxnIDs() []int {
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

func (tm basicTxnMap) Add(tcc *dto.TxnControlCounters) int {
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

func (tm basicTxnMap) Delete(tcc *dto.TxnControlCounters) int {
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

type BrutalGarbageCollectorExecutor interface {
	CollectAll() error
}

type AbstractFlatGarbageCollectorExecutor interface {
	Add(string, *dto.TxnControlCounters) error
	Condemn(string, *dto.TxnControlCounters) bool
	Collect() error
}

type GarbageCollectorExecutor interface {
	BrutalGarbageCollectorExecutor
	AbstractFlatGarbageCollectorExecutor
}

// Idiomatic golang singleton
// Credit to http://marcio.io/2015/07/singleton-pattern-in-go/
func GetGarbageCollectorExecutorInstance(sqlEngine sqlengine.SQLEngine, ns tablenamespace.TableNamespaceCollection, dialect sqldialect.SQLDialect) (GarbageCollectorExecutor, error) {
	var err error
	once.Do(func() {
		if err != nil {
			return
		}
		garbageCollectorExecutor, err = newBasicGarbageCollectorExecutor(dialect, ns)
	})
	return garbageCollectorExecutor, err
}

func newBasicGarbageCollectorExecutor(dialect sqldialect.SQLDialect, ns tablenamespace.TableNamespaceCollection) (GarbageCollectorExecutor, error) {
	return &basicGarbageCollectorExecutor{
		activeTxns:      newTxnMap(),
		activeTxnsCache: newTxnMap(),
		gcMutex:         &sync.Mutex{},
		ns:              ns,
		sqlDialect:      dialect,
	}, nil
}

// Algorithm summary:
//   - `Collect()` will reclaim resources from all txns **not** in supplied list of IDs.
//   - `CollectAll()` as assumed.
type basicGarbageCollectorExecutor struct {
	activeTxns      TxnMap
	activeTxnsCache TxnMap
	gcMutex         *sync.Mutex
	ns              tablenamespace.TableNamespaceCollection
	sqlDialect      sqldialect.SQLDialect
}

func (rc *basicGarbageCollectorExecutor) Add(tableName string, tcc *dto.TxnControlCounters) error {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	var err error
	if rc.ns.GetAnalyticsCacheTableNamespaceConfigurator().IsAllowed(tableName) {
		// err = rc.sqlDialect.GCAdd(tableName, tcc)
		if err != nil {
			return err
		}
		rc.activeTxnsCache.Add(tcc)
		return nil
	}
	// err = rc.sqlDialect.GCAdd(tableName, tcc)
	if err != nil {
		return err
	}
	rc.activeTxns.Add(tcc)
	return nil
}

func (rc *basicGarbageCollectorExecutor) Condemn(tableName string, tcc *dto.TxnControlCounters) bool {
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
func (rc *basicGarbageCollectorExecutor) Collect() error {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	activeTxnIDs := rc.activeTxns.GetTxnIDs()
	activeCacheTxnIDs := rc.activeTxnsCache.GetTxnIDs()
	return rc.sqlDialect.GCCollect(activeTxnIDs, activeCacheTxnIDs)
}

// Algorithm, **must be done during pause**:
//   - Execute **all possible** GC queries in a txn.
func (rc *basicGarbageCollectorExecutor) CollectAll() error {
	rc.gcMutex.Lock()
	defer rc.gcMutex.Unlock()
	return rc.sqlDialect.GCCollectAll()
}
