package primitivebuilder

import (
	"fmt"
	"strconv"

	"github.com/stackql/go-openapistackql/openapistackql"
	pkg_response "github.com/stackql/go-openapistackql/pkg/response"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"github.com/stackql/stackql/internal/stackql/acid/binlog"
	"github.com/stackql/stackql/internal/stackql/drm"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/httpmiddleware"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/builder_input"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/primitive_context"
	"github.com/stackql/stackql/internal/stackql/logging"
	"github.com/stackql/stackql/internal/stackql/primitive"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
)

type genericHTTP struct {
	graphHolder       primitivegraph.PrimitiveGraphHolder
	handlerCtx        handler.HandlerContext
	drmCfg            drm.Config
	root              primitivegraph.PrimitiveNode
	tbl               tablemetadata.ExtendedTableMetadata
	paramMap          map[int]map[string]interface{}
	commentDirectives sqlparser.CommentDirectives
	dependencyNode    primitivegraph.PrimitiveNode
	isAwait           bool
	verb              string // may be "insert" or "update"
	inputAlias        string
	isUndo            bool
}

func NewGenericHTTP(
	builderInput builder_input.BuilderInput,
) Builder {
	handlerCtx := builderInput.GetHandlerContext()
	return &genericHTTP{
		graphHolder:       builderInput.GetGraphHolder(),
		handlerCtx:        handlerCtx,
		drmCfg:            handlerCtx.GetDrmConfig(),
		tbl:               builderInput.GetTableMetadata(),
		paramMap:          builderInput.GetParamMap(),
		commentDirectives: builderInput.GetCommentDirectives(),
		dependencyNode:    builderInput.GetDependencyNode(),
		isAwait:           builderInput.IsAwait(),
		verb:              builderInput.GetVerb(),
		inputAlias:        builderInput.GetInputAlias(),
		isUndo:            builderInput.IsUndo(),
	}
}

func (gh *genericHTTP) GetRoot() primitivegraph.PrimitiveNode {
	return gh.root
}

func (gh *genericHTTP) GetTail() primitivegraph.PrimitiveNode {
	return gh.root
}

func (gh *genericHTTP) decorateOutput(op internaldto.ExecutorOutput, tableName string) internaldto.ExecutorOutput {
	op.SetUndoLog(
		binlog.NewSimpleLogEntry(
			nil,
			[]string{
				fmt.Sprintf("Undo the %s on %s", gh.verb, tableName),
			},
		),
	)
	return op
}

//nolint:funlen,errcheck,gocognit,cyclop,gocyclo // TODO: fix this
func (gh *genericHTTP) Build() error {
	paramMap := gh.paramMap
	tbl := gh.tbl
	handlerCtx := gh.handlerCtx
	commentDirectives := gh.commentDirectives
	isAwait := gh.isAwait
	prov, err := tbl.GetProvider()
	if err != nil {
		return err
	}
	svc, err := tbl.GetService()
	if err != nil {
		return err
	}
	m, err := tbl.GetMethod()
	if err != nil {
		return err
	}
	_, _, responseAnalysisErr := tbl.GetResponseSchemaAndMediaType()
	actionPrimitive := primitive.NewHTTPRestPrimitive(
		prov,
		nil,
		nil,
		nil,
		primitive_context.NewPrimitiveContext(),
	)
	target := make(map[string]interface{})
	ex := func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
		input, inputExists := actionPrimitive.GetInputFromAlias(gh.inputAlias)
		if !inputExists {
			return internaldto.NewErroneousExecutorOutput(fmt.Errorf("input does not exist"))
		}
		inputStream, inputErr := input.ResultToMap()
		if inputErr != nil {
			return internaldto.NewErroneousExecutorOutput(inputErr)
		}
		rr, rrErr := inputStream.Read()
		if rrErr != nil {
			return internaldto.NewErroneousExecutorOutput(rrErr)
		}
		inputMap, inputErr := rr.GetMap()
		if inputErr != nil {
			return internaldto.NewErroneousExecutorOutput(inputErr)
		}
		pr, prErr := prov.GetProvider()
		if prErr != nil {
			return internaldto.NewErroneousExecutorOutput(prErr)
		}
		httpPreparator := openapistackql.NewHTTPPreparator(
			pr,
			svc,
			m,
			inputMap,
			paramMap,
			nil,
			nil,
			logging.GetLogger(),
		)
		httpArmoury, httpErr := httpPreparator.BuildHTTPRequestCtx()
		if httpErr != nil {
			return internaldto.NewErroneousExecutorOutput(httpErr)
		}

		tableName, _ := tbl.GetTableName()

		// var reversalParameters []map[string]interface{}
		// TODO: Implement reversal operations:
		//           - depending upon reversal operation, collect sequence of HTTP operations.
		//           - for each HTTP operation, collect context and store in *some object*
		//                then attach to reversal graph for later retrieval and execution
		var nullaryExecutors []func() internaldto.ExecutorOutput
		for _, r := range httpArmoury.GetRequestParams() {
			req := r
			nullaryEx := func() internaldto.ExecutorOutput {
				response, apiErr := httpmiddleware.HTTPApiCallFromRequest(handlerCtx.Clone(), prov, m, req.GetRequest())
				if apiErr != nil {
					return internaldto.NewErroneousExecutorOutput(apiErr)
				}
				inverse := func() internaldto.ExecutorOutput {
					return gh.decorateOutput(nil, tableName)
				}
				logging.GetLogger().Infoln(fmt.Sprintf("inverse = %v", inverse()))

				if responseAnalysisErr == nil {
					var resp pkg_response.Response
					resp, err = m.ProcessResponse(response)
					if err != nil {
						return internaldto.NewErroneousExecutorOutput(err)
					}
					processedBody := resp.GetProcessedBody()
					switch processedBody := processedBody.(type) { //nolint:gocritic // TODO: fix this
					case map[string]interface{}:
						target = processedBody
					}
				}
				if err != nil {
					return internaldto.NewErroneousExecutorOutput(err)
				}
				logging.GetLogger().Infoln(fmt.Sprintf("target = %v", target))
				items, ok := target[tbl.LookupSelectItemsKey()]
				keys := make(map[string]map[string]interface{})
				if ok {
					iArr, iOk := items.([]interface{})
					if iOk && len(iArr) > 0 {
						for i := range iArr {
							item, itemOk := iArr[i].(map[string]interface{})
							if itemOk {
								keys[strconv.Itoa(i)] = item
							}
						}
					}
				}
				if err == nil {
					if response.StatusCode < 300 { //nolint:gomnd // TODO: fix this
						msgs := internaldto.NewBackendMessages(
							generateSuccessMessagesFromHeirarchy(tbl, isAwait),
						)
						return gh.decorateOutput(
							internaldto.NewExecutorOutput(
								nil,
								target,
								nil,
								msgs,
								nil,
							),
							tableName,
						)
					}
					generatedErr := fmt.Errorf("insert over HTTP error: %s", response.Status)
					return internaldto.NewExecutorOutput(
						nil,
						target,
						nil,
						nil,
						generatedErr,
					)
				}
				return internaldto.NewExecutorOutput(
					nil,
					target,
					nil,
					nil,
					err,
				)
			}

			nullaryExecutors = append(nullaryExecutors, nullaryEx)
		}
		resultSet := internaldto.NewErroneousExecutorOutput(fmt.Errorf("no executions detected"))
		if !isAwait {
			for _, ei := range nullaryExecutors {
				execInstance := ei
				aPrioriMessages := resultSet.GetMessages()
				resultSet = execInstance()
				resultSet.AppendMessages(aPrioriMessages)
				if resultSet.GetError() != nil {
					return resultSet
				}
			}
			return resultSet
		}
		for _, eI := range nullaryExecutors {
			execInstance := eI
			dependentInsertPrimitive := primitive.NewHTTPRestPrimitive(
				prov,
				nil,
				nil,
				nil,
				primitive_context.NewPrimitiveContext(),
			)
			err = dependentInsertPrimitive.SetExecutor(func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
				return execInstance()
			})
			if err != nil {
				return internaldto.NewErroneousExecutorOutput(err)
			}
			execPrim, execErr := composeAsyncMonitor(handlerCtx, dependentInsertPrimitive, tbl, commentDirectives)
			if execErr != nil {
				return internaldto.NewErroneousExecutorOutput(execErr)
			}
			resultSet = execPrim.Execute(pc)
			if resultSet.GetError() != nil {
				return resultSet
			}
		}
		return gh.decorateOutput(
			resultSet,
			tableName,
		)
	}
	err = actionPrimitive.SetExecutor(ex)
	if err != nil {
		return err
	}

	graphHolder := gh.graphHolder

	actionNode := graphHolder.CreatePrimitiveNode(actionPrimitive)
	if gh.dependencyNode != nil {
		actionPrimitive.SetInputAlias("", gh.dependencyNode.ID())
		graphHolder.NewDependency(gh.dependencyNode, actionNode, 1.0)
		gh.root = gh.dependencyNode
	} else {
		gh.root = actionNode
	}

	return nil
}
