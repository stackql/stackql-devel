package transact

import (
	"fmt"
	"sync"

	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"github.com/stackql/stackql/internal/stackql/acid/binlog"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
)

//nolint:gochecknoglobals // singleton pattern
var (
	coordinatorOnce      sync.Once
	coordinatorSingleton Coordinator
	_                    Coordinator = &standardCoordinator{}
	_                    Manager     = &basicTransactionManager{}
)

// The transaction coordinator is singleton
// that orchestrates transaction managers.
type Coordinator interface {
	// Create a new transaction manager.
	NewTxnManager() (Manager, error)
}

type standardCoordinator struct {
}

func (c *standardCoordinator) NewTxnManager() (Manager, error) {
	return NewManager(), nil
}

func GetCoordinatorInstance() (Coordinator, error) {
	var err error
	coordinatorOnce.Do(func() {
		if err != nil {
			return
		}
		coordinatorSingleton = &standardCoordinator{}
	})
	return coordinatorSingleton, err
}

// The transaction manager ensures
// that undo and redo logs are kept
// and that 2PC is performed.
type Manager interface {
	Statement
	// Begin a new transaction.
	Begin() (Manager, error)
	// Commit the current transaction.
	Commit() error
	// Rollback the current transaction.
	Rollback() error
	// Enqueue a transaction operation.
	// This method will return an error
	// in the case that the transaction
	// context disallows a particular
	// operation or type of operation.
	Enqueue(Statement) error
	// Get the depth of transaction nesting.
	Depth() int
	// Get the parent transaction manager.
	GetParent() (Manager, bool)
}

type basicTransactionManager struct {
	parent            Manager
	statementSequence []Statement
	undoLogs          []binlog.LogEntry
	redoLogs          []binlog.LogEntry
}

func newBasicTransactionManager(parent Manager) Manager {
	return &basicTransactionManager{
		parent: parent,
	}
}

func NewManager() Manager {
	return newBasicTransactionManager(nil)
}

func (m *basicTransactionManager) GetAST() (sqlparser.Statement, bool) {
	return nil, false
}

func (m *basicTransactionManager) GetParent() (Manager, bool) {
	return m.parent, m.parent != nil
}

func (m *basicTransactionManager) SetRedoLog(log binlog.LogEntry) {
	m.redoLogs = []binlog.LogEntry{log}
}

func (m *basicTransactionManager) SetUndoLog(log binlog.LogEntry) {
	m.undoLogs = []binlog.LogEntry{log}
}

func (m *basicTransactionManager) GetUndoLog() (binlog.LogEntry, bool) {
	if len(m.undoLogs) == 0 {
		return nil, false
	}
	initialUndoLog := m.undoLogs[len(m.undoLogs)-1]
	rv := initialUndoLog.Clone()
	for i := len(m.undoLogs) - 2; i >= 0; i-- { //nolint:gomnd // magic number second from last
		currentLog := m.undoLogs[i]
		if currentLog != nil {
			rv.AppendHumanReadable(currentLog.GetHumanReadable())
			rv.AppendRaw(currentLog.GetRaw())
		}
	}
	return rv, true
}

func (m *basicTransactionManager) GetRedoLog() (binlog.LogEntry, bool) {
	return nil, false
}

func (m *basicTransactionManager) Prepare() error {
	var err error
	for _, stmt := range m.statementSequence {
		err = stmt.Prepare()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *basicTransactionManager) Execute() internaldto.ExecutorOutput {
	err := m.execute()
	if err != nil {
		return internaldto.NewErroneousExecutorOutput(err)
	}
	return internaldto.NewExecutorOutput(
		nil,
		nil,
		nil,
		internaldto.NewBackendMessages([]string{"transaction committed"}),
		nil,
	)
}

func (m *basicTransactionManager) execute() error {
	for _, stmt := range m.statementSequence {
		coDomain := stmt.Execute()
		err := coDomain.GetError()
		undoLog, undoLogExists := stmt.GetUndoLog()
		redoLog, redoLogExists := stmt.GetRedoLog()
		if undoLogExists {
			m.undoLogs = append(m.undoLogs, undoLog)
		}
		if redoLogExists {
			m.redoLogs = append(m.redoLogs, redoLog)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *basicTransactionManager) Begin() (Manager, error) {
	if m.Depth() >= 1 {
		return nil, fmt.Errorf("cannot begin nested transaction")
	}
	return newBasicTransactionManager(m), nil
}

func (m *basicTransactionManager) Commit() error {
	return m.execute()
}

// Rollback is a no-op for now.
// The redo logs will simply be
// displayed to the user.
func (m *basicTransactionManager) Rollback() error {
	return nil
}

func (m *basicTransactionManager) Enqueue(stmt Statement) error {
	m.statementSequence = append(m.statementSequence, stmt)
	return nil
}

func (m *basicTransactionManager) Depth() int {
	if m.parent != nil {
		return m.parent.Depth() + 1
	}
	return 0
}
