package bundle

import (
	"github.com/stackql/stackql/internal/stackql/garbagecollector"
	"github.com/stackql/stackql/internal/stackql/kstore"
	"github.com/stackql/stackql/internal/stackql/pginternal"
	"github.com/stackql/stackql/internal/stackql/sqlcontrol"
	"github.com/stackql/stackql/internal/stackql/sqldialect"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/tablenamespace"
	"github.com/stackql/stackql/pkg/txncounter"
	"vitess.io/vitess/go/vt/sqlparser"
)

type Bundle interface {
	GetControlAttributes() sqlcontrol.ControlAttributes
	GetGC() garbagecollector.GarbageCollector
	GetNamespaceCollection() tablenamespace.TableNamespaceCollection
	GetPGInternalRouter() pginternal.PGInternalRouter
	GetSQLDialect() sqldialect.SQLDialect
	GetSQLEngine() sqlengine.SQLEngine
	GetTxnCounterManager() txncounter.TxnCounterManager
	GetTxnStore() kstore.KStore
}

func NewBundle(
	garbageCollector garbagecollector.GarbageCollector,
	namespaces tablenamespace.TableNamespaceCollection,
	sqlEngine sqlengine.SQLEngine,
	sqlDialect sqldialect.SQLDialect,
	pgInternalRouter pginternal.PGInternalRouter,
	controlAttributes sqlcontrol.ControlAttributes,
	txnStore kstore.KStore,
	txnCtrMgr txncounter.TxnCounterManager,
) Bundle {
	return &simpleBundle{
		garbageCollector:  garbageCollector,
		namespaces:        namespaces,
		sqlEngine:         sqlEngine,
		sqlDialect:        sqlDialect,
		controlAttributes: controlAttributes,
		txnStore:          txnStore,
		txnCtrMgr:         txnCtrMgr,
		formatter:         sqlDialect.GetASTFormatter(),
		pgInternalRouter:  pgInternalRouter,
	}
}

type simpleBundle struct {
	controlAttributes sqlcontrol.ControlAttributes
	garbageCollector  garbagecollector.GarbageCollector
	namespaces        tablenamespace.TableNamespaceCollection
	sqlEngine         sqlengine.SQLEngine
	sqlDialect        sqldialect.SQLDialect
	txnStore          kstore.KStore
	txnCtrMgr         txncounter.TxnCounterManager
	formatter         sqlparser.NodeFormatter
	pgInternalRouter  pginternal.PGInternalRouter
}

func (sb *simpleBundle) GetControlAttributes() sqlcontrol.ControlAttributes {
	return sb.controlAttributes
}

func (sb *simpleBundle) GetPGInternalRouter() pginternal.PGInternalRouter {
	return sb.pgInternalRouter
}

func (sb *simpleBundle) GetASTFormatter() sqlparser.NodeFormatter {
	return sb.formatter
}

func (sb *simpleBundle) GetTxnStore() kstore.KStore {
	return sb.txnStore
}

func (sb *simpleBundle) GetTxnCounterManager() txncounter.TxnCounterManager {
	return sb.txnCtrMgr
}

func (sb *simpleBundle) GetGC() garbagecollector.GarbageCollector {
	return sb.garbageCollector
}

func (sb *simpleBundle) GetSQLEngine() sqlengine.SQLEngine {
	return sb.sqlEngine
}

func (sb *simpleBundle) GetSQLDialect() sqldialect.SQLDialect {
	return sb.sqlDialect
}

func (sb *simpleBundle) GetNamespaceCollection() tablenamespace.TableNamespaceCollection {
	return sb.namespaces
}
