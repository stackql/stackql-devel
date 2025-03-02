package primitivebuilder

import (
	"fmt"
	"strconv"

	"github.com/stackql/any-sdk/pkg/logging"
	"github.com/stackql/any-sdk/pkg/streaming"
	"github.com/stackql/stackql/internal/stackql/drm"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/primitive_context"
	"github.com/stackql/stackql/internal/stackql/primitive"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/tableinsertioncontainer"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
)

// polyValentExecution implements the Builder interface
// and represents the action of acquiring data from an endpoint
// and then persisting that data into a table.
// This data would then subsequently be queried by later execution phases.
type polyValentExecution struct {
	graphHolder                primitivegraph.PrimitiveGraphHolder
	handlerCtx                 handler.HandlerContext
	tableMeta                  tablemetadata.ExtendedTableMetadata
	drmCfg                     drm.Config
	insertPreparedStatementCtx drm.PreparedStatementCtx
	insertionContainer         tableinsertioncontainer.TableInsertionContainer
	txnCtrlCtr                 internaldto.TxnControlCounters
	rowSort                    func(map[string]map[string]interface{}) []string
	root                       primitivegraph.PrimitiveNode
	stream                     streaming.MapStream
	isReadOnly                 bool //nolint:unused // TODO: build out
}

func newPolyValentExecution(
	graphHolder primitivegraph.PrimitiveGraphHolder,
	handlerCtx handler.HandlerContext,
	tableMeta tablemetadata.ExtendedTableMetadata,
	insertCtx drm.PreparedStatementCtx,
	insertionContainer tableinsertioncontainer.TableInsertionContainer,
	rowSort func(map[string]map[string]interface{}) []string,
	stream streaming.MapStream,
) Builder {
	var tcc internaldto.TxnControlCounters
	if insertCtx != nil {
		tcc = insertCtx.GetGCCtrlCtrs()
	}
	if stream == nil {
		stream = streaming.NewNopMapStream()
	}
	return &polyValentExecution{
		graphHolder:                graphHolder,
		handlerCtx:                 handlerCtx,
		tableMeta:                  tableMeta,
		rowSort:                    rowSort,
		drmCfg:                     handlerCtx.GetDrmConfig(),
		insertPreparedStatementCtx: insertCtx,
		insertionContainer:         insertionContainer,
		txnCtrlCtr:                 tcc,
		stream:                     stream,
	}
}

func (ss *polyValentExecution) GetRoot() primitivegraph.PrimitiveNode {
	return ss.root
}

func (ss *polyValentExecution) GetTail() primitivegraph.PrimitiveNode {
	return ss.root
}

// type eliderPayload struct {
// 	currentTcc  internaldto.TxnControlCounters
// 	tableName   string
// 	reqEncoding string
// }

//nolint:lll // chaining
func (ss *polyValentExecution) elideActionIfPossible(
	currentTcc internaldto.TxnControlCounters,
	tableName string,
	reqEncoding string,
) methodElider {
	elisionFunc := func(reqEncoding string, _ ...any) bool {
		olderTcc, isMatch := ss.handlerCtx.GetNamespaceCollection().GetAnalyticsCacheTableNamespaceConfigurator().Match(
			tableName,
			reqEncoding,
			ss.drmCfg.GetControlAttributes().GetControlLatestUpdateColumnName(), ss.drmCfg.GetControlAttributes().GetControlInsertEncodedIDColumnName())
		if isMatch {
			nonControlColumns := ss.insertPreparedStatementCtx.GetNonControlColumns()
			var nonControlColumnNames []string
			for _, c := range nonControlColumns {
				nonControlColumnNames = append(nonControlColumnNames, c.GetName())
			}
			//nolint:errcheck // TODO: fix
			ss.handlerCtx.GetGarbageCollector().Update(
				tableName,
				olderTcc.Clone(),
				currentTcc,
			)
			//nolint:errcheck // TODO: fix
			ss.insertionContainer.SetTableTxnCounters(tableName, olderTcc)
			ss.insertPreparedStatementCtx.SetGCCtrlCtrs(olderTcc)
			r, sqlErr := ss.handlerCtx.GetNamespaceCollection().GetAnalyticsCacheTableNamespaceConfigurator().Read(
				tableName, reqEncoding,
				ss.drmCfg.GetControlAttributes().GetControlInsertEncodedIDColumnName(),
				nonControlColumnNames)
			if sqlErr != nil {
				internaldto.NewErroneousExecutorOutput(sqlErr)
			}
			ss.drmCfg.ExtractObjectFromSQLRows(r, nonControlColumns, ss.stream)
			return true
		}
		return false
	}
	return newStandardMethodElider(elisionFunc)
}

//nolint:nestif,gocognit // acceptable for now
func (ss *polyValentExecution) ActionInsertPreparation(
	payload ActionInsertPayload,
) ActionInsertResult {
	itemisationResult := payload.GetItemisationResult()
	housekeepingDone := payload.IsHousekeepingDone()
	tableName := payload.GetTableName()
	paramsUsed := payload.GetParamsUsed()
	reqEncoding := payload.GetReqEncoding()

	items, _ := itemisationResult.GetItems()
	keys := make(map[string]map[string]interface{})
	iArr, iErr := castItemsArray(items)
	if iErr != nil {
		return newActionInsertResult(housekeepingDone, iErr)
	}
	streamErr := ss.stream.Write(iArr)
	if streamErr != nil {
		return newActionInsertResult(housekeepingDone, streamErr)
	}
	if len(iArr) > 0 {
		if !housekeepingDone && ss.insertPreparedStatementCtx != nil {
			_, execErr := ss.handlerCtx.GetSQLEngine().Exec(ss.insertPreparedStatementCtx.GetGCHousekeepingQueries())
			tcc := ss.insertPreparedStatementCtx.GetGCCtrlCtrs()
			tcc.SetTableName(tableName)
			//nolint:errcheck // TODO: fix
			ss.insertionContainer.SetTableTxnCounters(tableName, tcc)
			housekeepingDone = true
			if execErr != nil {
				return newActionInsertResult(housekeepingDone, execErr)
			}
		}

		for i, item := range iArr {
			if item != nil {
				if len(paramsUsed) > 0 {
					for k, v := range paramsUsed {
						if _, itemOk := item[k]; !itemOk {
							item[k] = v
						}
					}
				}

				logging.GetLogger().Infoln(
					fmt.Sprintf(
						"running insert with query = '''%s''', control parameters: %v",
						ss.insertPreparedStatementCtx.GetQuery(),
						ss.insertPreparedStatementCtx.GetGCCtrlCtrs(),
					),
				)
				r, rErr := ss.drmCfg.ExecuteInsertDML(
					ss.handlerCtx.GetSQLEngine(),
					ss.insertPreparedStatementCtx,
					item,
					reqEncoding,
				)
				logging.GetLogger().Infoln(
					fmt.Sprintf(
						"insert result = %v, error = %v",
						r,
						rErr,
					),
				)
				if rErr != nil {
					expandedErr := fmt.Errorf(
						"sql insert error: '%w' from query: %s",
						rErr,
						ss.insertPreparedStatementCtx.GetQuery(),
					)
					return newActionInsertResult(housekeepingDone, expandedErr)
				}
				keys[strconv.Itoa(i)] = item
			}
		}
	}

	return newActionInsertResult(housekeepingDone, nil)
}

//nolint:funlen,gocognit,gocyclo,cyclop,revive // TODO: investigate
func (ss *polyValentExecution) Build() error {
	prov, err := ss.tableMeta.GetProvider()
	if err != nil {
		return err
	}
	provider, providerErr := prov.GetProvider()
	if providerErr != nil {
		return providerErr
	}
	m, err := ss.tableMeta.GetMethod()
	if err != nil {
		return err
	}
	tableName, err := ss.tableMeta.GetTableName()
	if err != nil {
		return err
	}
	authCtx, authCtxErr := ss.handlerCtx.GetAuthContext(prov.GetProviderString())
	if authCtxErr != nil {
		return authCtxErr
	}
	ex := func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
		logging.GetLogger().Infof("polyValentExecution.Execute() beginning execution for table %s", tableName)
		currentTcc := ss.insertPreparedStatementCtx.GetGCCtrlCtrs().Clone()
		ss.graphHolder.AddTxnControlCounters(currentTcc)
		mr := prov.InferMaxResultsElement(m)
		polyHandler := newStandardPolyHandler(
			ss.handlerCtx,
		)
		agnosticatePayload := newHTTPAgnosticatePayload(
			ss.tableMeta,
			provider,
			m,
			tableName,
			authCtx,
			ss.handlerCtx.GetRuntimeContext(),
			ss.handlerCtx.GetOutErrFile(),
			mr,
			ss.elideActionIfPossible(
				currentTcc,
				tableName,
				"", // late binding, should remove AOT reference
			),
			true,
			polyHandler,
			ss.tableMeta.GetSelectItemsKey(),
			ss,
		)
		agnosticErr := agnosticate(agnosticatePayload)
		if agnosticErr != nil {
			return internaldto.NewErroneousExecutorOutput(agnosticErr)
		}
		messages := polyHandler.GetMessages()
		if len(messages) > 0 {
			return internaldto.NewNopEmptyExecutorOutput(messages)
		}
		return internaldto.NewEmptyExecutorOutput()
	}

	prep := func() drm.PreparedStatementCtx {
		return ss.insertPreparedStatementCtx
	}
	primitiveCtx := primitive_context.NewPrimitiveContext()
	primitiveCtx.SetIsReadOnly(true)
	insertPrim := primitive.NewGenericPrimitive(
		ex,
		prep,
		ss.txnCtrlCtr,
		primitiveCtx,
	).WithDebugName(fmt.Sprintf("insert_%s_%s", tableName, ss.tableMeta.GetAlias()))
	graphHolder := ss.graphHolder
	insertNode := graphHolder.CreatePrimitiveNode(insertPrim)
	ss.root = insertNode

	return nil
}
