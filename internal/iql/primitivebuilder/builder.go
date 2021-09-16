package primitivebuilder

import (
	"fmt"
	"sort"
	"strconv"

	"infraql/internal/iql/drm"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/handler"
	"infraql/internal/iql/httpexec"
	"infraql/internal/iql/httpmiddleware"
	"infraql/internal/iql/metadata"
	"infraql/internal/iql/primitive"
	"infraql/internal/iql/primitivegraph"
	"infraql/internal/iql/taxonomy"
	"infraql/internal/iql/util"

	"vitess.io/vitess/go/sqltypes"
	querypb "vitess.io/vitess/go/vt/proto/query"

	log "github.com/sirupsen/logrus"
)

type Builder interface {
	Build() error

	GetQuery() string

	GetRoot() primitivegraph.PrimitiveNode

	GetTail() primitivegraph.PrimitiveNode
}

type SingleSelectAcquire struct {
	primitiveBuilder           *PrimitiveBuilder
	query                      string
	handlerCtx                 *handler.HandlerContext
	tableMeta                  taxonomy.ExtendedTableMetadata
	tabulation                 metadata.Tabulation
	drmCfg                     drm.DRMConfig
	insertPreparedStatementCtx *drm.PreparedStatementCtx
	selectPreparedStatementCtx *drm.PreparedStatementCtx
	txnCtrlCtr                 *dto.TxnControlCounters
	rowSort                    func(map[string]map[string]interface{}) []string
	root                       primitivegraph.PrimitiveNode
}

type SingleSelect struct {
	primitiveBuilder           *PrimitiveBuilder
	query                      string
	handlerCtx                 *handler.HandlerContext
	tableMeta                  taxonomy.ExtendedTableMetadata
	tabulation                 metadata.Tabulation
	drmCfg                     drm.DRMConfig
	insertPreparedStatementCtx *drm.PreparedStatementCtx
	selectPreparedStatementCtx *drm.PreparedStatementCtx
	txnCtrlCtr                 *dto.TxnControlCounters
	rowSort                    func(map[string]map[string]interface{}) []string
	root                       primitivegraph.PrimitiveNode
}

type SingleAcquireAndSelect struct {
	primitiveBuilder *PrimitiveBuilder
	acquireBuilder   *SingleSelectAcquire
	selectBuilder    *SingleSelect
}

type Join struct {
	lhsPb, rhsPb *PrimitiveBuilder
	lhs, rhs     Builder
	handlerCtx   *handler.HandlerContext
	rowSort      func(map[string]map[string]interface{}) []string
}

func NewSingleSelectAcquire(pb *PrimitiveBuilder, handlerCtx *handler.HandlerContext, tableMeta taxonomy.ExtendedTableMetadata, insertCtx *drm.PreparedStatementCtx, selectCtx *drm.PreparedStatementCtx, rowSort func(map[string]map[string]interface{}) []string) *SingleSelectAcquire {
	return &SingleSelectAcquire{
		primitiveBuilder:           pb,
		handlerCtx:                 handlerCtx,
		tableMeta:                  tableMeta,
		rowSort:                    rowSort,
		drmCfg:                     handlerCtx.DrmConfig,
		insertPreparedStatementCtx: insertCtx,
		selectPreparedStatementCtx: selectCtx,
		txnCtrlCtr:                 selectCtx.TxnCtrlCtrs,
	}
}

func NewSingleSelect(pb *PrimitiveBuilder, handlerCtx *handler.HandlerContext, tableMeta taxonomy.ExtendedTableMetadata, insertCtx *drm.PreparedStatementCtx, selectCtx *drm.PreparedStatementCtx, rowSort func(map[string]map[string]interface{}) []string) *SingleSelect {
	return &SingleSelect{
		primitiveBuilder:           pb,
		handlerCtx:                 handlerCtx,
		tableMeta:                  tableMeta,
		rowSort:                    rowSort,
		drmCfg:                     handlerCtx.DrmConfig,
		insertPreparedStatementCtx: insertCtx,
		selectPreparedStatementCtx: selectCtx,
		txnCtrlCtr:                 selectCtx.TxnCtrlCtrs,
	}
}

func NewSingleAcquireAndSelect(pb *PrimitiveBuilder, handlerCtx *handler.HandlerContext, tableMeta taxonomy.ExtendedTableMetadata, insertCtx *drm.PreparedStatementCtx, selectCtx *drm.PreparedStatementCtx, rowSort func(map[string]map[string]interface{}) []string) *SingleAcquireAndSelect {
	return &SingleAcquireAndSelect{
		primitiveBuilder: pb,
		acquireBuilder:   NewSingleSelectAcquire(pb, handlerCtx, tableMeta, insertCtx, selectCtx, rowSort),
		selectBuilder:    NewSingleSelect(pb, handlerCtx, tableMeta, insertCtx, selectCtx, rowSort),
	}
}

func NewJoin(lhsPb *PrimitiveBuilder, rhsPb *PrimitiveBuilder, handlerCtx *handler.HandlerContext, rowSort func(map[string]map[string]interface{}) []string) *Join {
	return &Join{
		lhsPb:      lhsPb,
		rhsPb:      rhsPb,
		handlerCtx: handlerCtx,
		rowSort:    rowSort,
	}
}

func (ss *SingleSelect) Build() error {

	selectEx := func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput {
		defer ss.handlerCtx.SQLEngine.GCCollectObsolete(ss.insertPreparedStatementCtx.TxnCtrlCtrs)
		altKeys := make(map[string]map[string]interface{})
		rawRows := make(map[int]map[int]interface{})

		// select phase
		log.Infoln(fmt.Sprintf("running select with control parameters: %v", ss.selectPreparedStatementCtx.TxnCtrlCtrs))
		r, sqlErr := ss.drmCfg.QueryDML(ss.handlerCtx.SQLEngine, ss.selectPreparedStatementCtx, nil)
		log.Infoln(fmt.Sprintf("select result = %v, error = %v", r, sqlErr))
		var ks []int
		i := 0
		var keyArr []string
		var ifArr []interface{}
		for i < len(ss.selectPreparedStatementCtx.NonControlColumns) {
			x := ss.selectPreparedStatementCtx.NonControlColumns[i]
			y := ss.drmCfg.GetGolangValue(x.GetType())
			ifArr = append(ifArr, y)
			keyArr = append(keyArr, x.Column.GetIdentifier())
			i++
		}
		if r != nil {
			i := 0
			for r.Next() {
				errScan := r.Scan(ifArr...)
				if errScan != nil {
					log.Infoln(fmt.Sprintf("%v", errScan))
				}
				for ord, val := range ifArr {
					log.Infoln(fmt.Sprintf("col #%d '%s':  %v  type: %T", ord, ss.selectPreparedStatementCtx.NonControlColumns[ord].GetName(), val, val))
				}
				im := make(map[string]interface{})
				imRaw := make(map[int]interface{})
				for ord, key := range keyArr {
					val := ifArr[ord]
					ev := ss.drmCfg.ExtractFromGolangValue(val)
					im[key] = ev
					imRaw[ord] = ev
				}
				altKeys[strconv.Itoa(i)] = im
				rawRows[i] = imRaw
				ks = append(ks, i)
				i++
			}

			for ord := range ks {
				val := altKeys[strconv.Itoa(ord)]
				log.Infoln(fmt.Sprintf("row #%d:  %v  type: %T", ord, val, val))
			}
		}
		var cNames []string
		for _, v := range ss.selectPreparedStatementCtx.NonControlColumns {
			cNames = append(cNames, v.Column.GetIdentifier())
		}
		rowSort := func(m map[string]map[string]interface{}) []string {
			var arr []int
			for k, _ := range m {
				ord, _ := strconv.Atoi(k)
				arr = append(arr, ord)
			}
			sort.Ints(arr)
			var rv []string
			for _, v := range arr {
				rv = append(rv, strconv.Itoa(v))
			}
			return rv
		}
		rv := util.PrepareResultSet(dto.NewPrepareResultSetPlusRawDTO(nil, altKeys, cNames, rowSort, nil, nil, rawRows))
		if rv.GetSQLResult() == nil {

			resVal := &sqltypes.Result{
				Fields: make([]*querypb.Field, len(ss.selectPreparedStatementCtx.NonControlColumns)),
			}

			for f := range resVal.Fields {
				resVal.Fields[f] = &querypb.Field{
					Name: cNames[f],
				}
			}
			rv.GetSQLResult = func() *sqltypes.Result { return resVal }
		}
		return rv
	}
	graph := ss.primitiveBuilder.GetGraph()
	selectNode := graph.CreatePrimitiveNode(NewLocalPrimitive(selectEx))
	ss.root = selectNode

	return nil
}

func (ss *SingleSelectAcquire) Build() error {
	prov, err := ss.tableMeta.GetProvider()
	if err != nil {
		return err
	}
	ex := func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput {
		mr := prov.InferMaxResultsElement(ss.tableMeta.HeirarchyObjects.Method)
		if mr != nil {
			_, ok := ss.tableMeta.HeirarchyObjects.Method.Parameters[mr.Name]
			if ok && ss.handlerCtx.RuntimeContext.HTTPMaxResults > 0 {
				for i, param := range ss.tableMeta.HttpArmoury.RequestParams {
					param.Context.SetQueryParam("maxResults", strconv.Itoa(ss.handlerCtx.RuntimeContext.HTTPMaxResults))
					ss.tableMeta.HttpArmoury.RequestParams[i] = param
				}
			}
		}
		for _, reqCtx := range ss.tableMeta.HttpArmoury.RequestParams {
			response, apiErr := httpmiddleware.HttpApiCall(*(ss.handlerCtx), prov, reqCtx.Context)
			housekeepingDone := false
			for {
				if apiErr != nil {
					return util.PrepareResultSet(dto.NewPrepareResultSetDTO(nil, nil, nil, ss.rowSort, apiErr, nil))
				}
				target, err := httpexec.ProcessHttpResponse(response)
				if err != nil {
					return dto.NewErroneousExecutorOutput(err)
				}
				log.Infoln(fmt.Sprintf("target = %v", target))
				items, ok := target[ss.tableMeta.SelectItemsKey]
				keys := make(map[string]map[string]interface{})

				if ok {
					iArr, ok := items.([]interface{})
					if ok && len(iArr) > 0 {
						if !housekeepingDone {
							_, err = ss.handlerCtx.SQLEngine.Exec(ss.insertPreparedStatementCtx.GetGCHousekeepingQueries())
							housekeepingDone = true
						}
						if err != nil {
							return dto.NewErroneousExecutorOutput(err)
						}

						for i := range iArr {
							item, ok := iArr[i].(map[string]interface{})
							if ok {

								log.Infoln(fmt.Sprintf("running insert with control parameters: %v", ss.insertPreparedStatementCtx.TxnCtrlCtrs))
								r, err := ss.drmCfg.ExecuteInsertDML(ss.handlerCtx.SQLEngine, ss.insertPreparedStatementCtx, item)
								log.Infoln(fmt.Sprintf("insert result = %v, error = %v", r, err))
								keys[strconv.Itoa(i)] = item
							}
						}
					}
				}
				npt := prov.InferNextPageResponseElement(ss.tableMeta.HeirarchyObjects.Method)
				nptKey := prov.InferNextPageRequestElement(ss.tableMeta.HeirarchyObjects.Method)
				if npt == nil || nptKey == nil {
					break
				}
				nextPageToken, ok := target[npt.Name]
				if !ok || nextPageToken == "" {
					log.Infoln("breaking out")
					break
				}
				tk, ok := nextPageToken.(string)
				if !ok {
					log.Infoln("breaking out")
					break
				}
				reqCtx.Context.SetQueryParam(nptKey.Name, tk)
				response, apiErr = httpmiddleware.HttpApiCall(*(ss.handlerCtx), prov, reqCtx.Context)
			}
		}
		return dto.ExecutorOutput{}
	}

	prep := func() *drm.PreparedStatementCtx {
		return ss.selectPreparedStatementCtx
	}
	insertPrim := NewHTTPRestPrimitive(
		prov,
		ex,
		prep,
		ss.txnCtrlCtr,
	)
	graph := ss.primitiveBuilder.GetGraph()
	insertNode := graph.CreatePrimitiveNode(insertPrim)
	ss.root = insertNode

	return nil
}

func (ss *SingleAcquireAndSelect) GetRoot() primitivegraph.PrimitiveNode {
	return ss.acquireBuilder.GetRoot()
}

func (ss *SingleAcquireAndSelect) GetTail() primitivegraph.PrimitiveNode {
	return ss.selectBuilder.GetTail()
}

func (ss *SingleAcquireAndSelect) GetQuery() string {
	return ss.acquireBuilder.query
}

func (ss *SingleAcquireAndSelect) Build() error {
	err := ss.acquireBuilder.Build()
	if err != nil {
		return err
	}
	err = ss.selectBuilder.Build()
	if err != nil {
		return err
	}
	graph := ss.primitiveBuilder.GetGraph()
	graph.NewDependency(ss.acquireBuilder.root, ss.selectBuilder.root, 1.0)
	return nil
}

func (ss *SingleSelect) GetRoot() primitivegraph.PrimitiveNode {
	return ss.root
}

func (ss *SingleSelect) GetTail() primitivegraph.PrimitiveNode {
	return ss.root
}

func (ss *SingleSelect) GetQuery() string {
	return ss.query
}

func (ss *SingleSelectAcquire) GetRoot() primitivegraph.PrimitiveNode {
	return ss.root
}

func (ss *SingleSelectAcquire) GetTail() primitivegraph.PrimitiveNode {
	return ss.root
}

func (ss *SingleSelectAcquire) GetQuery() string {
	return ss.query
}

func (j *Join) Build() error {
	return nil
}

func (j *Join) GetQuery() string {
	return ""
}

func (j *Join) getErrNode() primitivegraph.PrimitiveNode {
	graph := j.lhsPb.GetGraph()
	return graph.CreatePrimitiveNode(
		NewLocalPrimitive(
			func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput {
				return util.GenerateSimpleErroneousOutput(fmt.Errorf("joins not yet supported"))
			},
		))
}

func (j *Join) GetRoot() primitivegraph.PrimitiveNode {
	return j.getErrNode()
}

func (j *Join) GetTail() primitivegraph.PrimitiveNode {
	return j.getErrNode()
}
