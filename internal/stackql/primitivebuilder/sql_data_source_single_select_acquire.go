package primitivebuilder

// import (
// 	"fmt"
// 	"strconv"

// 	"github.com/stackql/stackql/internal/stackql/drm"
// 	"github.com/stackql/stackql/internal/stackql/handler"
// 	"github.com/stackql/stackql/internal/stackql/httpmiddleware"
// 	"github.com/stackql/stackql/internal/stackql/internaldto"
// 	"github.com/stackql/stackql/internal/stackql/logging"
// 	"github.com/stackql/stackql/internal/stackql/primitive"
// 	"github.com/stackql/stackql/internal/stackql/primitivegraph"
// 	"github.com/stackql/stackql/internal/stackql/streaming"
// 	"github.com/stackql/stackql/internal/stackql/tableinsertioncontainer"
// 	"github.com/stackql/stackql/internal/stackql/tablemetadata"
// 	"github.com/stackql/stackql/internal/stackql/util"
// )

// // sqlDataSourceSingleSelectAcquire implements the Builder interface
// // and represents the action of acquiring data from an endpoint
// // and then persisting that data into a table.
// // This data would then subsequently be queried by later execution phases.
// type sqlDataSourceSingleSelectAcquire struct {
// 	query                      string
// 	graph                      primitivegraph.PrimitiveGraph
// 	handlerCtx                 handler.HandlerContext
// 	tableMeta                  tablemetadata.ExtendedTableMetadata
// 	drmCfg                     drm.DRMConfig
// 	insertPreparedStatementCtx drm.PreparedStatementCtx
// 	insertionContainer         tableinsertioncontainer.TableInsertionContainer
// 	txnCtrlCtr                 internaldto.TxnControlCounters
// 	rowSort                    func(map[string]map[string]interface{}) []string
// 	root                       primitivegraph.PrimitiveNode
// 	stream                     streaming.MapStream
// }

// func NewSQLDataSourceSingleSelectAcquire(
// 	graph primitivegraph.PrimitiveGraph,
// 	handlerCtx handler.HandlerContext,
// 	insertionContainer tableinsertioncontainer.TableInsertionContainer,
// 	query string,
// 	insertCtx drm.PreparedStatementCtx,
// 	rowSort func(map[string]map[string]interface{}) []string,
// 	stream streaming.MapStream,
// ) Builder {
// 	tableMeta := insertionContainer.GetTableMetadata()
// 	return newSQLDataSourceSingleSelectAcquire(
// 		graph,
// 		handlerCtx,
// 		tableMeta,
// 		insertCtx,
// 		insertionContainer,
// 		query,
// 		rowSort,
// 		stream,
// 	)
// }

// func newSQLDataSourceSingleSelectAcquire(
// 	graph primitivegraph.PrimitiveGraph,
// 	handlerCtx handler.HandlerContext,
// 	tableMeta tablemetadata.ExtendedTableMetadata,
// 	insertCtx drm.PreparedStatementCtx,
// 	insertionContainer tableinsertioncontainer.TableInsertionContainer,
// 	query string,
// 	rowSort func(map[string]map[string]interface{}) []string,
// 	stream streaming.MapStream,
// ) Builder {
// 	var tcc internaldto.TxnControlCounters
// 	if insertCtx != nil {
// 		tcc = insertCtx.GetGCCtrlCtrs()
// 	}
// 	if stream == nil {
// 		stream = streaming.NewNopMapStream()
// 	}
// 	return &sqlDataSourceSingleSelectAcquire{
// 		graph:                      graph,
// 		handlerCtx:                 handlerCtx,
// 		tableMeta:                  tableMeta,
// 		rowSort:                    rowSort,
// 		drmCfg:                     handlerCtx.GetDrmConfig(),
// 		insertPreparedStatementCtx: insertCtx,
// 		insertionContainer:         insertionContainer,
// 		txnCtrlCtr:                 tcc,
// 		stream:                     stream,
// 		query:                      query,
// 	}
// }

// func (ss *sqlDataSourceSingleSelectAcquire) GetRoot() primitivegraph.PrimitiveNode {
// 	return ss.root
// }

// func (ss *sqlDataSourceSingleSelectAcquire) GetTail() primitivegraph.PrimitiveNode {
// 	return ss.root
// }

// func (ss *sqlDataSourceSingleSelectAcquire) Build() error {
// 	sqlDB, ok := ss.tableMeta.GetSQLDataSource()
// 	if !ok {
// 		return fmt.Errorf("sql data source unavailable for sql data source query")
// 	}
// 	tableName, err := ss.tableMeta.GetTableName()
// 	if err != nil {
// 		return err
// 	}
// 	ex := func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
// 		rows, err := sqlDB.Query(ss.query)
// 		if err != nil {
// 			return internaldto.NewErroneousExecutorOutput(err)
// 		}
// 		currentTcc := ss.insertPreparedStatementCtx.GetGCCtrlCtrs().Clone()
// 		ss.graph.AddTxnControlCounters(currentTcc)

// 		for {
// 			// TODO: fix cloning ops
// 			ok := rows.Next()
// 			if !ok {
// 				break
// 			}
// 			response, apiErr := httpmiddleware.HttpApiCallFromRequest(ss.handlerCtx.Clone(), prov, m, reqCtx.GetRequest().Clone(reqCtx.GetRequest().Context()))
// 			housekeepingDone := false
// 			for {
// 				if apiErr != nil {
// 					return util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, nil, nil, ss.rowSort, apiErr, nil))
// 				}
// 				res, err := m.ProcessResponse(response)
// 				if err != nil {
// 					return internaldto.NewErroneousExecutorOutput(err)
// 				}
// 				ss.handlerCtx.LogHTTPResponseMap(res.GetProcessedBody())
// 				if err != nil {
// 					return internaldto.NewErroneousExecutorOutput(err)
// 				}
// 				logging.GetLogger().Infoln(fmt.Sprintf("target = %v", res))
// 				var items interface{}
// 				var ok bool
// 				target := res.GetProcessedBody()
// 				switch pl := target.(type) {
// 				// add case for xml object,
// 				case map[string]interface{}:
// 					if ss.tableMeta.GetSelectItemsKey() != "" && ss.tableMeta.GetSelectItemsKey() != "/*" {
// 						items, ok = pl[ss.tableMeta.GetSelectItemsKey()]
// 						if !ok {
// 							items = []interface{}{
// 								pl,
// 							}
// 							ok = true
// 						}
// 					} else {
// 						items = []interface{}{
// 							pl,
// 						}
// 						ok = true
// 					}
// 				case []interface{}:
// 					items = pl
// 					ok = true
// 				case []map[string]interface{}:
// 					items = pl
// 					ok = true
// 				case nil:
// 					return internaldto.ExecutorOutput{}
// 				}
// 				keys := make(map[string]map[string]interface{})

// 				if ok {
// 					iArr, err := castItemsArray(items)
// 					if err != nil {
// 						return internaldto.NewErroneousExecutorOutput(err)
// 					}
// 					err = ss.stream.Write(iArr)
// 					if err != nil {
// 						return internaldto.NewErroneousExecutorOutput(err)
// 					}
// 					if ok && len(iArr) > 0 {
// 						if !housekeepingDone && ss.insertPreparedStatementCtx != nil {
// 							_, err = ss.handlerCtx.GetSQLEngine().Exec(ss.insertPreparedStatementCtx.GetGCHousekeepingQueries())
// 							tcc := ss.insertPreparedStatementCtx.GetGCCtrlCtrs()
// 							tcc.SetTableName(tableName)
// 							ss.insertionContainer.SetTableTxnCounters(tableName, tcc)
// 							housekeepingDone = true
// 						}
// 						if err != nil {
// 							return internaldto.NewErroneousExecutorOutput(err)
// 						}

// 						for i, item := range iArr {
// 							if item != nil {

// 								if err == nil {
// 									for k, v := range paramsUsed {
// 										if _, ok := item[k]; !ok {
// 											item[k] = v
// 										}
// 									}
// 								}

// 								logging.GetLogger().Infoln(fmt.Sprintf("running insert with control parameters: %v", ss.insertPreparedStatementCtx.GetGCCtrlCtrs()))
// 								r, err := ss.drmCfg.ExecuteInsertDML(ss.handlerCtx.GetSQLEngine(), ss.insertPreparedStatementCtx, item, reqEncoding)
// 								logging.GetLogger().Infoln(fmt.Sprintf("insert result = %v, error = %v", r, err))
// 								if err != nil {
// 									return internaldto.NewErroneousExecutorOutput(fmt.Errorf("sql insert error: '%s' from query: %s", err.Error(), ss.insertPreparedStatementCtx.GetQuery()))
// 								}
// 								keys[strconv.Itoa(i)] = item
// 							}
// 						}
// 					}
// 				}
// 				if npt == nil || nptRequest == nil {
// 					break
// 				}
// 				tk := extractNextPageToken(res, npt)
// 				if tk == "" || tk == "<nil>" || tk == "[]" || (ss.handlerCtx.GetRuntimeContext().HTTPPageLimit > 0 && pageCount >= ss.handlerCtx.GetRuntimeContext().HTTPPageLimit) {
// 					break
// 				}
// 				pageCount++
// 				req, err := reqCtx.SetNextPage(m, tk, nptRequest)
// 				if err != nil {
// 					return internaldto.NewErroneousExecutorOutput(err)
// 				}
// 				response, apiErr = httpmiddleware.HttpApiCallFromRequest(ss.handlerCtx.Clone(), prov, m, req)
// 			}
// 			if reqCtx.GetRequest() != nil {
// 				q := reqCtx.GetRequest().URL.Query()
// 				q.Del(nptRequest.GetName())
// 				reqCtx.SetRawQuery(q.Encode())
// 			}
// 		}
// 		return internaldto.ExecutorOutput{}
// 	}

// 	prep := func() drm.PreparedStatementCtx {
// 		return ss.insertPreparedStatementCtx
// 	}
// 	insertPrim := primitive.NewHTTPRestPrimitive(
// 		prov,
// 		ex,
// 		prep,
// 		ss.txnCtrlCtr,
// 	)
// 	graph := ss.graph
// 	insertNode := graph.CreatePrimitiveNode(insertPrim)
// 	ss.root = insertNode

// 	return nil
// }
