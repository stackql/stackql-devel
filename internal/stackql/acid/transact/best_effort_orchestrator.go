package transact

import (
	"fmt"
	"strings"

	"github.com/stackql/stackql/internal/stackql/acid/binlog"
	"github.com/stackql/stackql/internal/stackql/acid/txn_context"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
)

var (
	_ Orchestrator = &bestEffortOrchestrator{}
)

// This orchestrator:
//   - Supports a simple reversibility semantic.
//   - In cases of network partitioning or other failures,
//     it will simply spew suggested undo logs.
type bestEffortOrchestrator struct {
	txnCoordinator Coordinator
	undoLogs       []binlog.LogEntry
	redoLogs       []binlog.LogEntry
}

func (orc *bestEffortOrchestrator) ProcessQueryOrQueries(
	handlerCtx handler.HandlerContext,
) ([]internaldto.ExecutorOutput, bool) {
	return orc.processQueryOrQueries(handlerCtx)
}

func (orc *bestEffortOrchestrator) processQueryOrQueries(
	handlerCtx handler.HandlerContext,
) ([]internaldto.ExecutorOutput, bool) {
	var retVal []internaldto.ExecutorOutput
	cmdString := handlerCtx.GetRawQuery()
	for _, s := range strings.Split(cmdString, ";") {
		response, hasResponse := orc.processQuery(handlerCtx, s)
		if hasResponse {
			retVal = append(retVal, response...)
		}
	}
	return retVal, len(retVal) > 0
}

//nolint:gocognit,dupl // TODO: review
func (orc *bestEffortOrchestrator) processQuery(
	handlerCtx handler.HandlerContext,
	query string,
) ([]internaldto.ExecutorOutput, bool) {
	if query == "" {
		return nil, false
	}
	clonedCtx := handlerCtx.Clone()
	clonedCtx.SetQuery(query)
	transactStatement := NewStatement(query, clonedCtx, txn_context.NewTransactionContext(orc.txnCoordinator.Depth()))
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
		txnCoordinator, beginErr := orc.txnCoordinator.Begin()
		if beginErr != nil {
			return []internaldto.ExecutorOutput{
				internaldto.NewErroneousExecutorOutput(beginErr),
			}, true
		}
		orc.txnCoordinator = txnCoordinator
		return []internaldto.ExecutorOutput{
			internaldto.NewNopEmptyExecutorOutput([]string{"OK"}),
		}, true
	} else if transactStatement.IsCommit() {
		commitCoDomain := orc.txnCoordinator.Commit()
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
		parent, hasParent := orc.txnCoordinator.GetParent()
		if hasParent {
			orc.txnCoordinator = parent
			retVal = append(retVal, internaldto.NewNopEmptyExecutorOutput([]string{"OK"}))
			return retVal, true
		}
		noParentErr := fmt.Errorf(noParentMessage)
		retVal = append(retVal, internaldto.NewErroneousExecutorOutput(noParentErr))
		return retVal, true
	} else if transactStatement.IsRollback() {
		var retVal []internaldto.ExecutorOutput
		rollbackErr := orc.txnCoordinator.Rollback()
		if rollbackErr != nil {
			retVal = append(retVal, internaldto.NewErroneousExecutorOutput(rollbackErr))
		}
		parent, hasParent := orc.txnCoordinator.GetParent()
		if hasParent {
			orc.txnCoordinator = parent
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
	if isReadOnly || orc.txnCoordinator.IsRoot() {
		stmtOutput := transactStatement.Execute()
		return []internaldto.ExecutorOutput{
			stmtOutput,
		}, true
	}

	// TODO: fix this crap
	//       remember, we do not have all undo log data until we get pk back
	undoLog, undoLogExists := transactStatement.GetUndoLog()
	if !undoLogExists {
		// TODO: bail
	}
	redoLog, redoLogExists := transactStatement.GetRedoLog()
	if redoLogExists {
		orc.redoLogs = append(orc.redoLogs, redoLog)
	}
	// logging.GetLogger().Debugf("undoLog.Size() = %d", undoLog.Size())
	orc.undoLogs = append(orc.undoLogs, undoLog)
	execErr := transactStatement.Execute()
	if execErr != nil {
		// TODO: bail
	}
	output := transactStatement.Execute()
	if output.GetError() != nil {
		// TODO: bail
	}
	return []internaldto.ExecutorOutput{output}, true
}
