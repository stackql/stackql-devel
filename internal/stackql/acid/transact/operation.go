package transact

import (
	"fmt"

	"github.com/stackql/stackql/internal/stackql/primitive"
)

var (
	_ Operation = &reversibleOperation{}
	_ Operation = &irreversibleOperation{}
)

// The Operation is an abstract
// data type that represents
// a stackql action.
// The operation maps to each of:
//   - an executable action.
//   - a redo log entry.
//   - an undo log entry.
//
// One possible implementation is to
// store a nullable primitive (plan) graph
// node alongside log entries.
type Operation interface {
	// Execute the operation.
	Execute() error
	// Reverse the operation.
	Undo() error
	// Get the redo log entry.
	GetRedoLog() (LogEntry, bool)
	// Get the undo log entry.
	GetUndoLog() (LogEntry, bool)
}

//nolint:unused // under construction
type reversibleOperation struct {
	redoLog LogEntry
	undoLog LogEntry
	pr      primitive.IPrimitive
	pc      primitive.IPrimitiveCtx
}

func (op *reversibleOperation) Execute() error {
	return nil
}

func (op *reversibleOperation) Undo() error {
	return nil
}

func (op *reversibleOperation) Redo() error {
	return nil
}

func (op *reversibleOperation) GetRedoLog() (LogEntry, bool) {
	return op.redoLog, true
}

func (op *reversibleOperation) GetUndoLog() (LogEntry, bool) {
	return op.undoLog, true
}

type irreversibleOperation struct {
	redoLog LogEntry
	pc      primitive.IPrimitiveCtx
	pr      primitive.IPrimitive
}

func (op *irreversibleOperation) Execute() error {
	res := op.pr.Execute(op.pc)
	return res.GetError()
}

func (op *irreversibleOperation) Undo() error {
	return fmt.Errorf("irreversible operation cannot be undone")
}

func (op *irreversibleOperation) Redo() error {
	return nil
}

func (op *irreversibleOperation) GetRedoLog() (LogEntry, bool) {
	return op.redoLog, true
}

func (op *irreversibleOperation) GetUndoLog() (LogEntry, bool) {
	return nil, false
}
