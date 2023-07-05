package transact

import (
	"fmt"
	"strings"
	"sync"

	"github.com/stackql/stackql/internal/stackql/acid/txn_context"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
)

//nolint:gochecknoglobals // singleton pattern
var (
	providerOnce      sync.Once
	providerSingleton Provider
	_                 Provider = &standardProvider{}
	noParentMessage   string   = "no parent transaction manager available" //nolint:gochecknoglobals,revive,lll // permissable
)

const (
	defaultMaxStackDepth = 1
)

// The transaction provider is singleton
// that orchestrates transaction managers.
type Provider interface {
	// Create a new transaction manager.
	ProcessQueryOrQueries(handler.HandlerContext) ([]internaldto.ExecutorOutput, bool)
}

type standardProvider struct {
	ctx            txn_context.ITransactionCoordinatorContext
	txnCoordinator Coordinator
}

func newTxnCoordinator(ctx txn_context.ITransactionCoordinatorContext) Coordinator {
	maxTxnDepth := defaultMaxStackDepth
	if ctx != nil {
		maxTxnDepth = ctx.GetMaxStackDepth()
	}
	return NewCoordinator(maxTxnDepth)
}

func GetProviderInstance(ctx txn_context.ITransactionCoordinatorContext) (Provider, error) {
	var err error
	providerOnce.Do(func() {
		if err != nil {
			return
		}
		providerSingleton = &standardProvider{
			ctx:            ctx,
			txnCoordinator: newTxnCoordinator(ctx),
		}
	})
	return providerSingleton, err
}

func (c *standardProvider) ProcessQueryOrQueries(
	handlerCtx handler.HandlerContext,
) ([]internaldto.ExecutorOutput, bool) {
	return c.processQueryOrQueries(handlerCtx)
}

func (c *standardProvider) processQueryOrQueries(
	handlerCtx handler.HandlerContext,
) ([]internaldto.ExecutorOutput, bool) {
	var retVal []internaldto.ExecutorOutput
	cmdString := handlerCtx.GetRawQuery()
	for _, s := range strings.Split(cmdString, ";") {
		response, hasResponse := c.processQuery(handlerCtx, s)
		if hasResponse {
			retVal = append(retVal, response...)
		}
	}
	return retVal, len(retVal) > 0
}

//nolint:gocognit // TODO: review
func (c *standardProvider) processQuery(
	handlerCtx handler.HandlerContext,
	query string,
) ([]internaldto.ExecutorOutput, bool) {
	if query == "" {
		return nil, false
	}
	clonedCtx := handlerCtx.Clone()
	clonedCtx.SetQuery(query)
	transactStatement := NewStatement(query, clonedCtx, txn_context.NewTransactionContext(c.txnCoordinator.Depth()))
	prepareErr := transactStatement.Prepare()
	if prepareErr != nil {
		return []internaldto.ExecutorOutput{
			internaldto.NewErroneousExecutorOutput(prepareErr),
		}, true
	}
	isReadOnly := transactStatement.IsReadOnly()
	// TODO: implement eager execution for non-mutating statements
	//       and lazy execution for mutating statements.
	// TODO: implement transaction stack.
	if transactStatement.IsBegin() { //nolint:gocritic,nestif // TODO: review
		txnCoordinator, beginErr := c.txnCoordinator.Begin()
		if beginErr != nil {
			return []internaldto.ExecutorOutput{
				internaldto.NewErroneousExecutorOutput(beginErr),
			}, true
		}
		c.txnCoordinator = txnCoordinator
		return []internaldto.ExecutorOutput{
			internaldto.NewNopEmptyExecutorOutput([]string{"OK"}),
		}, true
	} else if transactStatement.IsCommit() {
		commitCoDomain := c.txnCoordinator.Commit()
		commitErr, commitErrExists := commitCoDomain.GetError()
		if commitErrExists {
			retVal := []internaldto.ExecutorOutput{
				internaldto.NewErroneousExecutorOutput(commitErr),
			}
			undoLog, undoLogExists := commitCoDomain.GetUndoLog()
			if undoLogExists && undoLog != nil {
				humanReadable := undoLog.GetHumanReadable()
				if len(humanReadable) > 0 {
					displayUndoLogs := make([]string, len(humanReadable))
					for i, h := range humanReadable {
						displayUndoLogs[i] = fmt.Sprintf("UNDO required: %s", h)
					}
					retVal = append(retVal, internaldto.NewNopEmptyExecutorOutput(displayUndoLogs))
				}
			}
			return retVal, true
		}
		retVal := commitCoDomain.GetExecutorOutput()
		parent, hasParent := c.txnCoordinator.GetParent()
		if hasParent {
			c.txnCoordinator = parent
			retVal = append(retVal, internaldto.NewNopEmptyExecutorOutput([]string{"OK"}))
			return retVal, true
		}
		noParentErr := fmt.Errorf(noParentMessage)
		retVal = append(retVal, internaldto.NewErroneousExecutorOutput(noParentErr))
		return retVal, true
	} else if transactStatement.IsRollback() {
		var retVal []internaldto.ExecutorOutput
		rollbackErr := c.txnCoordinator.Rollback()
		if rollbackErr != nil {
			retVal = append(retVal, internaldto.NewErroneousExecutorOutput(rollbackErr))
		}
		parent, hasParent := c.txnCoordinator.GetParent()
		if hasParent {
			c.txnCoordinator = parent
			retVal = append(retVal, internaldto.NewNopEmptyExecutorOutput([]string{"Rollback OK"}))
			return retVal, true
		}
		retVal = append(
			retVal,
			internaldto.NewErroneousExecutorOutput(
				fmt.Errorf(noParentMessage)),
		)
		return retVal, true
	}
	if isReadOnly || c.txnCoordinator.IsRoot() {
		stmtOutput := transactStatement.Execute()
		return []internaldto.ExecutorOutput{
			stmtOutput,
		}, true
	}
	c.txnCoordinator.Enqueue(transactStatement) //nolint:errcheck // TODO: investigate
	return []internaldto.ExecutorOutput{
		internaldto.NewNopEmptyExecutorOutput([]string{"mutating statement queued"}),
	}, true
}
