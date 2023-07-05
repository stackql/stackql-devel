package transact

import (
	"github.com/stackql/stackql/internal/stackql/acid/acid_dto"
)

func NewCoordinator(maxTxnDepth int) Coordinator {
	return newBasicLazyTransactionCoordinator(nil, maxTxnDepth)
}

// The transaction coordinator ensures
// that undo and redo logs are kept
// and that 2PC is performed.
type Coordinator interface {
	Statement
	// Begin a new transaction.
	Begin() (Coordinator, error)
	// Commit the current transaction.
	Commit() acid_dto.CommitCoDomain
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
	GetParent() (Coordinator, bool)
	//
	IsRoot() bool
}
