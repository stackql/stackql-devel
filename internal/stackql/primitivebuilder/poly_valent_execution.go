package primitivebuilder

// import (
// 	"fmt"
// 	"strconv"

// 	"github.com/stackql/any-sdk/anysdk"
// 	"github.com/stackql/any-sdk/pkg/logging"
// 	pkg_response "github.com/stackql/any-sdk/pkg/response"
// 	"github.com/stackql/any-sdk/pkg/streaming"
// 	"github.com/stackql/stackql-parser/go/vt/sqlparser"
// 	"github.com/stackql/stackql/internal/stackql/drm"
// 	"github.com/stackql/stackql/internal/stackql/handler"
// 	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
// 	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/primitive_context"
// 	"github.com/stackql/stackql/internal/stackql/primitive"
// 	"github.com/stackql/stackql/internal/stackql/primitivegraph"
// 	"github.com/stackql/stackql/internal/stackql/tableinsertioncontainer"
// 	"github.com/stackql/stackql/internal/stackql/tablemetadata"
// )

// var (
// 	_ PolyValentExecutorFactory = (*polyValentExecution)(nil)
// )

// type PolyValentExecutorFactory interface {
// 	GetExecutor() (func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput, error)
// }

// type polyValentExecution struct {
// 	graphHolder                primitivegraph.PrimitiveGraphHolder
// 	handlerCtx                 handler.HandlerContext
// 	tableMeta                  tablemetadata.ExtendedTableMetadata
// 	drmCfg                     drm.Config
// 	insertPreparedStatementCtx drm.PreparedStatementCtx
// 	insertionContainer         tableinsertioncontainer.TableInsertionContainer
// 	txnCtrlCtr                 internaldto.TxnControlCounters
// 	rowSort                    func(map[string]map[string]interface{}) []string
// 	root                       primitivegraph.PrimitiveNode
// 	stream                     streaming.MapStream
// 	isReadOnly                 bool //nolint:unused // TODO: build out
// 	isAwait                    bool
// 	commentDirectives          sqlparser.CommentDirectives
// }

// func newPolyValentExecutorFactory(
// 	graphHolder primitivegraph.PrimitiveGraphHolder,
// 	handlerCtx handler.HandlerContext,
// 	tableMeta tablemetadata.ExtendedTableMetadata,
// 	insertCtx drm.PreparedStatementCtx,
// 	insertionContainer tableinsertioncontainer.TableInsertionContainer,
// 	rowSort func(map[string]map[string]interface{}) []string,
// 	stream streaming.MapStream,
// ) PolyValentExecutorFactory {
// 	var tcc internaldto.TxnControlCounters
// 	if insertCtx != nil {
// 		tcc = insertCtx.GetGCCtrlCtrs()
// 	}
// 	if stream == nil {
// 		stream = streaming.NewNopMapStream()
// 	}
// 	return &polyValentExecution{
// 		graphHolder:                graphHolder,
// 		handlerCtx:                 handlerCtx,
// 		tableMeta:                  tableMeta,
// 		rowSort:                    rowSort,
// 		drmCfg:                     handlerCtx.GetDrmConfig(),
// 		insertPreparedStatementCtx: insertCtx,
// 		insertionContainer:         insertionContainer,
// 		txnCtrlCtr:                 tcc,
// 		stream:                     stream,
// 	}
// }

// //nolint:funlen,gocognit,gocyclo,cyclop,revive // TODO: investigate
// func (pv *polyValentExecution) GetExecutor() (func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput, error) {
// 	prov, err := pv.tableMeta.GetProvider()
// 	if err != nil {
// 		return nil, err
// 	}
// 	provider, providerErr := prov.GetProvider()
// 	if providerErr != nil {
// 		return nil, providerErr
// 	}
// 	m, err := pv.tableMeta.GetMethod()
// 	if err != nil {
// 		return nil, err
// 	}
// 	tableName, err := pv.tableMeta.GetTableName()
// 	if err != nil {
// 		return nil, err
// 	}
// 	authCtx, authCtxErr := pv.handlerCtx.GetAuthContext(prov.GetProviderString())
// 	if authCtxErr != nil {
// 		return nil, authCtxErr
// 	}
// 	handlerCtx := pv.handlerCtx
// 	rtCtx := handlerCtx.GetRuntimeContext()
// 	outErrFile := handlerCtx.GetOutErrFile()
// 	commentDirectives := pv.commentDirectives
// 	isAwait := pv.isAwait
// 	_, _, responseAnalysisErr := m.GetResponseBodySchemaAndMediaType()
// 	target := make(map[string]interface{})
// 	ex := func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
// 		httpPreparator, httpPreparatorExists := gh.reversalStream.Next()
// 		resultSet := internaldto.NewErroneousExecutorOutput(fmt.Errorf("no executions detected"))
// 		var err error
// 		for {
// 			if !httpPreparatorExists {
// 				break
// 			}
// 			httpArmoury, httpErr := httpPreparator.BuildHTTPRequestCtx()
// 			if httpErr != nil {
// 				return internaldto.NewErroneousExecutorOutput(httpErr)
// 			}

// 			var nullaryExecutors []func() internaldto.ExecutorOutput
// 			for _, r := range httpArmoury.GetRequestParams() {
// 				req := r
// 				newMonoValentExecutorFactory(
// 					pv.graphHolder,
// 					pv.handlerCtx,
// 					pv.tableMeta,
// 					pv.insertCtx,
// 					insertionContainer,
// 					rowSort,
// 					stream,
// 				)
// 				nullaryEx := func() internaldto.ExecutorOutput {
// 					cc := anysdk.NewAnySdkClientConfigurator(rtCtx, provider.GetName())
// 					response, apiErr := anysdk.CallFromSignature(
// 						cc, rtCtx, authCtx, authCtx.Type, false, outErrFile, provider,
// 						anysdk.NewAnySdkOpStoreDesignation(m), req.GetArgList())
// 					if apiErr != nil {
// 						return internaldto.NewErroneousExecutorOutput(apiErr)
// 					}
// 					httpResponse, httpResponseErr := response.GetHttpResponse()
// 					if httpResponse != nil && httpResponse.Body != nil {
// 						defer httpResponse.Body.Close()
// 					}
// 					if httpResponseErr != nil {
// 						return internaldto.NewErroneousExecutorOutput(httpResponseErr)
// 					}

// 					if responseAnalysisErr == nil {
// 						var resp pkg_response.Response
// 						processed, processErr := m.ProcessResponse(httpResponse)
// 						if processErr != nil {
// 							return internaldto.NewErroneousExecutorOutput(processErr)
// 						}
// 						resp, respOk := processed.GetResponse()
// 						if !respOk {
// 							return internaldto.NewErroneousExecutorOutput(fmt.Errorf("response is not a valid response"))
// 						}
// 						processedBody := resp.GetProcessedBody()
// 						switch processedBody := processedBody.(type) { //nolint:gocritic // TODO: fix this
// 						case map[string]interface{}:
// 							target = processedBody
// 						}
// 					}
// 					if err != nil {
// 						return internaldto.NewErroneousExecutorOutput(err)
// 					}
// 					logging.GetLogger().Infoln(fmt.Sprintf("target = %v", target))
// 					items, ok := target[tablemetadata.LookupSelectItemsKey(m)]
// 					keys := make(map[string]map[string]interface{})
// 					if ok {
// 						iArr, iOk := items.([]interface{})
// 						if iOk && len(iArr) > 0 {
// 							for i := range iArr {
// 								item, itemOk := iArr[i].(map[string]interface{})
// 								if itemOk {
// 									keys[strconv.Itoa(i)] = item
// 								}
// 							}
// 						}
// 					}
// 					if err == nil {
// 						if httpResponse.StatusCode < 300 { //nolint:mnd // TODO: fix this
// 							msgs := internaldto.NewBackendMessages(
// 								[]string{"undo over HTTP successful"},
// 							)
// 							return gh.decorateOutput(
// 								internaldto.NewExecutorOutput(
// 									nil,
// 									target,
// 									nil,
// 									msgs,
// 									nil,
// 								),
// 								tableName,
// 							)
// 						}
// 						generatedErr := fmt.Errorf("undo over HTTP error: %s", httpResponse.Status)
// 						return internaldto.NewExecutorOutput(
// 							nil,
// 							target,
// 							nil,
// 							nil,
// 							generatedErr,
// 						)
// 					}
// 					return internaldto.NewExecutorOutput(
// 						nil,
// 						target,
// 						nil,
// 						nil,
// 						err,
// 					)
// 				}

// 				nullaryExecutors = append(nullaryExecutors, nullaryEx)
// 			}
// 			if !isAwait {
// 				for _, ei := range nullaryExecutors {
// 					execInstance := ei
// 					aPrioriMessages := resultSet.GetMessages()
// 					resultSet = execInstance()
// 					resultSet.AppendMessages(aPrioriMessages)
// 					if resultSet.GetError() != nil {
// 						return resultSet
// 					}
// 				}
// 				return resultSet
// 			}
// 			for _, eI := range nullaryExecutors {
// 				execInstance := eI
// 				dependentInsertPrimitive := primitive.NewGenericPrimitive(
// 					nil,
// 					nil,
// 					nil,
// 					primitive_context.NewPrimitiveContext(),
// 				)
// 				//nolint:revive // no big deal
// 				err = dependentInsertPrimitive.SetExecutor(func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
// 					return execInstance()
// 				})
// 				if err != nil {
// 					return internaldto.NewErroneousExecutorOutput(err)
// 				}
// 				execPrim, execErr := composeAsyncMonitor(handlerCtx, dependentInsertPrimitive, prov, m, commentDirectives)
// 				if execErr != nil {
// 					return internaldto.NewErroneousExecutorOutput(execErr)
// 				}
// 				resultSet = execPrim.Execute(pc)
// 				if resultSet.GetError() != nil {
// 					return resultSet
// 				}
// 			}
// 			httpPreparator, httpPreparatorExists = gh.reversalStream.Next()
// 		}
// 		return gh.decorateOutput(
// 			resultSet,
// 			tableName,
// 		)
// 	}
// 	return ex, nil
// }
