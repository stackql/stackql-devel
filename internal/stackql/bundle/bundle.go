package bundle

import (
	"github.com/stackql/stackql/internal/stackql/garbagecollector"
	"github.com/stackql/stackql/internal/stackql/sqlcontrol"
	"github.com/stackql/stackql/internal/stackql/sqldialect"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/tablenamespace"
)

type Bundle interface {
	GetControlAttributes() sqlcontrol.ControlAttributes
	GetGC() garbagecollector.GarbageCollector
	GetNamespaceCollection() tablenamespace.TableNamespaceCollection
	GetSQLDialect() sqldialect.SQLDialect
	GetSQLEngine() sqlengine.SQLEngine
}

func NewBundle(
	garbageCollector garbagecollector.GarbageCollector,
	namespaces tablenamespace.TableNamespaceCollection,
	sqlEngine sqlengine.SQLEngine,
	sqlDialect sqldialect.SQLDialect,
	controlAttributes sqlcontrol.ControlAttributes,
) Bundle {
	return &simpleBundle{
		garbageCollector:  garbageCollector,
		namespaces:        namespaces,
		sqlEngine:         sqlEngine,
		sqlDialect:        sqlDialect,
		controlAttributes: controlAttributes,
	}
}

type simpleBundle struct {
	controlAttributes sqlcontrol.ControlAttributes
	garbageCollector  garbagecollector.GarbageCollector
	namespaces        tablenamespace.TableNamespaceCollection
	sqlEngine         sqlengine.SQLEngine
	sqlDialect        sqldialect.SQLDialect
}

func (sb *simpleBundle) GetControlAttributes() sqlcontrol.ControlAttributes {
	return sb.controlAttributes
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
