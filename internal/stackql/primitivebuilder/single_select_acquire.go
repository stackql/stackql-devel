package primitivebuilder

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/stackql/any-sdk/anysdk"
	"github.com/stackql/any-sdk/pkg/dto"
	"github.com/stackql/any-sdk/pkg/httpelement"
	"github.com/stackql/any-sdk/pkg/logging"
	"github.com/stackql/any-sdk/pkg/response"
	"github.com/stackql/any-sdk/pkg/streaming"
	"github.com/stackql/stackql/internal/stackql/drm"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/primitive_context"
	"github.com/stackql/stackql/internal/stackql/primitive"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/tableinsertioncontainer"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
	"github.com/stackql/stackql/internal/stackql/util"

	sdk_internal_dto "github.com/stackql/any-sdk/pkg/internaldto"
)

// SingleSelectAcquire implements the Builder interface
// and represents the action of acquiring data from an endpoint
// and then persisting that data into a table.
// This data would then subsequently be queried by later execution phases.
type SingleSelectAcquire struct {
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

func NewSingleSelectAcquire(
	graphHolder primitivegraph.PrimitiveGraphHolder,
	handlerCtx handler.HandlerContext,
	insertionContainer tableinsertioncontainer.TableInsertionContainer,
	insertCtx drm.PreparedStatementCtx,
	rowSort func(map[string]map[string]interface{}) []string,
	stream streaming.MapStream,
) Builder {
	tableMeta := insertionContainer.GetTableMetadata()
	_, isGraphQL := tableMeta.GetGraphQL()
	if isGraphQL {
		return newGraphQLSingleSelectAcquire(
			graphHolder,
			handlerCtx,
			tableMeta,
			insertCtx,
			insertionContainer,
			rowSort,
			stream,
		)
	}
	return newSingleSelectAcquire(
		graphHolder,
		handlerCtx,
		tableMeta,
		insertCtx,
		insertionContainer,
		rowSort,
		stream,
	)
}

func newSingleSelectAcquire(
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
	return &SingleSelectAcquire{
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

func (ss *SingleSelectAcquire) GetRoot() primitivegraph.PrimitiveNode {
	return ss.root
}

func (ss *SingleSelectAcquire) GetTail() primitivegraph.PrimitiveNode {
	return ss.root
}

// type eliderPayload struct {
// 	currentTcc  internaldto.TxnControlCounters
// 	tableName   string
// 	reqEncoding string
// }

type standardMethodElider struct {
	elisionFunc func(...any) bool
}

func (sme *standardMethodElider) IsElide(argz ...any) bool {
	return sme.elisionFunc(argz...)
}

func newStandardMethodElider(elisionFunc func(...any) bool) methodElider {
	return &standardMethodElider{
		elisionFunc: elisionFunc,
	}
}

//nolint:lll // chaining
func (ss *SingleSelectAcquire) elideActionIfPossible(
	currentTcc internaldto.TxnControlCounters,
	tableName string,
	reqEncoding string,
) methodElider {
	elisionFunc := func(_ ...any) bool {
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

type methodElider interface {
	IsElide(...any) bool
}

// func (ss *SingleSelectAcquire) actionHTTP(
// 	_ methodElider,
// ) error {
// 	return nil
// }

type actionInsertResult struct {
	err                error
	isHousekeepingDone bool
}

type ActionInsertResult interface {
	GetError() (error, bool)
	IsHousekeepingDone() bool
}

//nolint:revive // no idea why this is a thing
func (air *actionInsertResult) GetError() (error, bool) {
	return air.err, air.err != nil
}

func (air *actionInsertResult) IsHousekeepingDone() bool {
	return air.isHousekeepingDone
}

func newActionInsertResult(isHousekeepingDone bool, err error) ActionInsertResult {
	return &actionInsertResult{
		err:                err,
		isHousekeepingDone: isHousekeepingDone,
	}
}

type itemsDTO struct {
	items        interface{}
	ok           bool
	isNilPayload bool
}

type ItemisationResult interface {
	GetItems() (interface{}, bool)
	IsOk() bool
	IsNilPayload() bool
}

func (id *itemsDTO) GetItems() (interface{}, bool) {
	return id.items, id.items != nil
}

func (id *itemsDTO) IsOk() bool {
	return id.ok
}

func (id *itemsDTO) IsNilPayload() bool {
	return id.isNilPayload
}

func newItemisationResult(
	items interface{},
	ok bool,
	isNilPayload bool,
) ItemisationResult {
	return &itemsDTO{
		items:        items,
		ok:           ok,
		isNilPayload: isNilPayload,
	}
}

//nolint:nestif // apathy
func itemise(
	target interface{},
	resErr error,
	selectItemsKey string,
) ItemisationResult {
	var items interface{}
	var ok bool
	logging.GetLogger().Infoln(fmt.Sprintf("SingleSelectAcquire.Execute() target = %v", target))
	switch pl := target.(type) {
	// add case for xml object,
	case map[string]interface{}:
		if selectItemsKey != "" && selectItemsKey != "/*" {
			items, ok = pl[selectItemsKey]
			if !ok {
				if resErr != nil {
					items = []interface{}{}
					ok = true
				} else {
					items = []interface{}{
						pl,
					}
					ok = true
				}
			}
		} else {
			items = []interface{}{
				pl,
			}
			ok = true
		}
	case []interface{}:
		items = pl
		ok = true
	case []map[string]interface{}:
		items = pl
		ok = true
	case nil:
		return newItemisationResult(nil, false, true)
	}
	return newItemisationResult(items, ok, false)
}

func inferNextPageResponseElement(provider anysdk.Provider, method anysdk.OperationStore) sdk_internal_dto.HTTPElement {
	st, ok := method.GetPaginationResponseTokenSemantic()
	if ok {
		if tp, err := sdk_internal_dto.ExtractHTTPElement(st.GetLocation()); err == nil {
			rv := sdk_internal_dto.NewHTTPElement(
				tp,
				st.GetKey(),
			)
			transformer, tErr := st.GetTransformer()
			if tErr == nil && transformer != nil {
				rv.SetTransformer(transformer)
			}
			return rv
		}
	}
	providerStr := provider.GetName()
	switch providerStr {
	case "github", "okta":
		rv := sdk_internal_dto.NewHTTPElement(
			sdk_internal_dto.Header,
			"Link",
		)
		rv.SetTransformer(anysdk.DefaultLinkHeaderTransformer)
		return rv
	default:
		return sdk_internal_dto.NewHTTPElement(
			sdk_internal_dto.BodyAttribute,
			"nextPageToken",
		)
	}
}

func inferNextPageRequestElement(provider anysdk.Provider, method anysdk.OperationStore) sdk_internal_dto.HTTPElement {
	st, ok := method.GetPaginationRequestTokenSemantic()
	if ok {
		if tp, err := sdk_internal_dto.ExtractHTTPElement(st.GetLocation()); err == nil {
			rv := sdk_internal_dto.NewHTTPElement(
				tp,
				st.GetKey(),
			)
			transformer, tErr := st.GetTransformer()
			if tErr == nil && transformer != nil {
				rv.SetTransformer(transformer)
			}
			return rv
		}
	}
	providerStr := provider.GetName()
	switch providerStr {
	case "github", "okta":
		return sdk_internal_dto.NewHTTPElement(
			sdk_internal_dto.RequestString,
			"",
		)
	default:
		return sdk_internal_dto.NewHTTPElement(
			sdk_internal_dto.QueryParam,
			"pageToken",
		)
	}
}

type PagingState interface {
	GetPageCount() int
	IsFinished() bool
	GetHTTPResponse() *http.Response
	GetAPIError() error
}

type httpPagingState struct {
	pageCount    int
	isFinished   bool
	httpResponse *http.Response
	apiErr       error
}

func (hps *httpPagingState) GetPageCount() int {
	return hps.pageCount
}

func (hps *httpPagingState) IsFinished() bool {
	return hps.isFinished
}

func (hps *httpPagingState) GetHTTPResponse() *http.Response {
	return hps.httpResponse
}

func (hps *httpPagingState) GetAPIError() error {
	return hps.apiErr
}

func newPagingState(
	pageCount int,
	isFinished bool,
	httpResponse *http.Response,
	apiErr error,
) PagingState {
	return &httpPagingState{
		pageCount:    pageCount,
		isFinished:   isFinished,
		httpResponse: httpResponse,
		apiErr:       apiErr,
	}
}

func page(
	res response.Response,
	method anysdk.OperationStore,
	provider anysdk.Provider,
	reqCtx anysdk.HTTPArmouryParameters,
	pageCount int,
	rtCtx dto.RuntimeCtx,
	authCtx *dto.AuthCtx,
	outErrFile io.Writer,
) PagingState {
	npt := inferNextPageResponseElement(provider, method)
	nptRequest := inferNextPageRequestElement(provider, method)
	if npt == nil || nptRequest == nil {
		return newPagingState(pageCount, true, nil, nil)
	}
	tk := extractNextPageToken(res, npt)
	//nolint:lll // long conditional
	if tk == "" || tk == "<nil>" || tk == "[]" || (rtCtx.HTTPPageLimit > 0 && pageCount >= rtCtx.HTTPPageLimit) {
		return newPagingState(pageCount, true, nil, nil)
	}
	pageCount++
	req, reqErr := reqCtx.SetNextPage(method, tk, nptRequest)
	if reqErr != nil {
		return newPagingState(pageCount, true, nil, reqErr)
	}
	cc := anysdk.NewAnySdkClientConfigurator(rtCtx, provider.GetName())
	response, apiErr := anysdk.HTTPApiCallFromRequest(
		cc, rtCtx, authCtx, authCtx.Type, false, outErrFile, provider, method, req)
	return newPagingState(pageCount, false, response, apiErr)
}

//nolint:nestif,gocognit // acceptable for now
func (ss *SingleSelectAcquire) actionInsertPreparation(
	itemisationResult ItemisationResult,
	housekeepingDone bool,
	tableName string,
	paramsUsed map[string]interface{},
	reqEncoding string,
) ActionInsertResult {
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
func (ss *SingleSelectAcquire) Build() error {
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
		logging.GetLogger().Infof("SingleSelectAcquire.Execute() beginning execution for table %s", tableName)
		currentTcc := ss.insertPreparedStatementCtx.GetGCCtrlCtrs().Clone()
		ss.graphHolder.AddTxnControlCounters(currentTcc)
		mr := prov.InferMaxResultsElement(m)
		// TODO: instrument for split source vertices !!!important!!!
		httpArmoury, armouryErr := ss.tableMeta.GetHTTPArmoury()
		if armouryErr != nil {
			//nolint:errcheck // TODO: fix
			ss.handlerCtx.GetOutErrFile().Write([]byte(
				fmt.Sprintf(
					"error assembling http aspects for resource '%s': %s\n",
					m.GetResource().GetID(),
					armouryErr.Error(),
				),
			),
			)
			return internaldto.NewErroneousExecutorOutput(armouryErr)
		}
		if mr != nil {
			// TODO: infer param position and act accordingly
			ok := true
			if ok && ss.handlerCtx.GetRuntimeContext().HTTPMaxResults > 0 {
				passOverParams := httpArmoury.GetRequestParams()
				for i, p := range passOverParams {
					param := p
					// param.Context.SetQueryParam("maxResults", strconv.Itoa(ss.handlerCtx.GetRuntimeContext().HTTPMaxResults))
					q := param.GetQuery()
					q.Set("maxResults", strconv.Itoa(ss.handlerCtx.GetRuntimeContext().HTTPMaxResults))
					param.SetRawQuery(q.Encode())
					passOverParams[i] = param
				}
				httpArmoury.SetRequestParams(passOverParams)
			}
		}
		reqParams := httpArmoury.GetRequestParams()
		logging.GetLogger().Infof("SingleSelectAcquire.Execute() req param count = %d", len(reqParams))
		for i, rc := range reqParams {
			var urlStringForLogging string
			if rc.GetRequest() != nil && rc.GetRequest().URL != nil {
				urlStringForLogging = rc.GetRequest().URL.String()
			}
			logging.GetLogger().Infof("SingleSelectAcquire.Execute() executing request %d: %s", i, urlStringForLogging)
			reqCtx := rc
			paramsUsed, paramErr := reqCtx.ToFlatMap()
			if paramErr != nil {
				return internaldto.NewErroneousExecutorOutput(paramErr)
			}
			reqEncoding := reqCtx.Encode()
			elider := ss.elideActionIfPossible(currentTcc, tableName, reqEncoding)
			elideOk := elider.IsElide(reqEncoding)
			if elideOk {
				return internaldto.NewEmptyExecutorOutput()
			}
			// TODO: fix cloning ops
			cc := anysdk.NewAnySdkClientConfigurator(ss.handlerCtx.GetRuntimeContext(), provider.GetName())
			response, apiErr := anysdk.HTTPApiCallFromRequest(
				cc,
				ss.handlerCtx.GetRuntimeContext(),
				authCtx,
				authCtx.Type,
				false,
				ss.handlerCtx.GetOutErrFile(),
				provider,
				m,
				reqCtx.GetRequest().Clone(
					reqCtx.GetRequest().Context(),
				),
			)
			// TODO: refactor into package !!TECH_DEBT!!
			if response != nil && response.StatusCode >= 400 {
				continue
			}
			housekeepingDone := false
			nptRequest := inferNextPageRequestElement(provider, m)
			pageCount := 1
			for {
				if apiErr != nil {
					return util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, nil, nil, ss.rowSort, apiErr, nil,
						ss.handlerCtx.GetTypingConfig(),
					))
				}
				processed, resErr := m.ProcessResponse(response)
				if resErr != nil {
					//nolint:errcheck // TODO: fix
					ss.handlerCtx.GetOutErrFile().Write(
						[]byte(fmt.Sprintf("error processing response: %s\n", resErr.Error())),
					)
					if processed == nil {
						return internaldto.NewErroneousExecutorOutput(resErr)
					}
				}
				res, respOk := processed.GetResponse()
				if !respOk {
					return internaldto.NewErroneousExecutorOutput(fmt.Errorf("response is not a valid response"))
				}
				if res.HasError() {
					return internaldto.NewNopEmptyExecutorOutput([]string{res.Error()})
				}
				ss.handlerCtx.LogHTTPResponseMap(res.GetProcessedBody())
				logging.GetLogger().Infoln(fmt.Sprintf("SingleSelectAcquire.Execute() response = %v", res))

				itemisationResult := itemise(res.GetProcessedBody(), resErr, ss.tableMeta.GetSelectItemsKey())

				if itemisationResult.IsNilPayload() {
					break
				}

				insertPrepResult := ss.actionInsertPreparation(
					itemisationResult,
					housekeepingDone,
					tableName,
					paramsUsed,
					reqEncoding,
				)
				housekeepingDone = insertPrepResult.IsHousekeepingDone()
				insertPrepErr, hasInsertPrepErr := insertPrepResult.GetError()
				if hasInsertPrepErr {
					return internaldto.NewErroneousExecutorOutput(insertPrepErr)
				}

				pageResult := page(
					res,
					m,
					provider,
					reqCtx,
					pageCount,
					ss.handlerCtx.GetRuntimeContext(),
					authCtx,
					ss.handlerCtx.GetOutErrFile(),
				)

				if pageResult.IsFinished() {
					break
				}

				pageCount = pageResult.GetPageCount()

				response = pageResult.GetHTTPResponse()
				apiErr = pageResult.GetAPIError()
			}
			if reqCtx.GetRequest() != nil {
				q := reqCtx.GetRequest().URL.Query()
				q.Del(nptRequest.GetName())
				reqCtx.SetRawQuery(q.Encode())
			}
		}
		logging.GetLogger().Infof("SingleSelectAcquire.Execute() returning empty for table %s", tableName)
		return internaldto.NewEmptyExecutorOutput()
	}

	prep := func() drm.PreparedStatementCtx {
		return ss.insertPreparedStatementCtx
	}
	primitiveCtx := primitive_context.NewPrimitiveContext()
	primitiveCtx.SetIsReadOnly(true)
	insertPrim := primitive.NewHTTPRestPrimitive(
		prov,
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

func extractNextPageToken(res response.Response, tokenKey sdk_internal_dto.HTTPElement) string {
	//nolint:exhaustive // TODO: review
	switch tokenKey.GetType() {
	case sdk_internal_dto.BodyAttribute:
		return extractNextPageTokenFromBody(res, tokenKey)
	case sdk_internal_dto.Header:
		return extractNextPageTokenFromHeader(res, tokenKey)
	}
	return ""
}

func extractNextPageTokenFromHeader(res response.Response, tokenKey sdk_internal_dto.HTTPElement) string {
	r := res.GetHttpResponse()
	if r == nil {
		return ""
	}
	header := r.Header
	if tokenKey.IsTransformerPresent() {
		tf, err := tokenKey.Transformer(header)
		if err != nil {
			return ""
		}
		rv, ok := tf.(string)
		if !ok {
			return ""
		}
		return rv
	}
	vals := header.Values(tokenKey.GetName())
	if len(vals) == 1 {
		return vals[0]
	}
	return ""
}

func extractNextPageTokenFromBody(res response.Response, tokenKey sdk_internal_dto.HTTPElement) string {
	elem, err := httpelement.NewHTTPElement(tokenKey.GetName(), "body")
	if err == nil {
		rawVal, rawErr := res.ExtractElement(elem)
		if rawErr == nil {
			switch v := rawVal.(type) {
			case []interface{}:
				if len(v) == 1 {
					return fmt.Sprintf("%v", v[0])
				}
			default:
				return fmt.Sprintf("%v", v)
			}
		}
	}
	body := res.GetProcessedBody()
	switch target := body.(type) { //nolint:gocritic // TODO: review
	case map[string]interface{}:
		tokenName := tokenKey.GetName()
		nextPageToken, ok := target[tokenName]
		if !ok || nextPageToken == "" {
			logging.GetLogger().Infoln("breaking out")
			return ""
		}
		tk, ok := nextPageToken.(string)
		if !ok {
			logging.GetLogger().Infoln("breaking out")
			return ""
		}
		return tk
	}
	return ""
}
