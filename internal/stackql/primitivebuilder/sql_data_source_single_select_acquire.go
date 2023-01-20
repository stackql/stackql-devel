package primitivebuilder

import (
	"fmt"

	"github.com/stackql/stackql/internal/stackql/datasource/sql_datasource"
	"github.com/stackql/stackql/internal/stackql/drm"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/primitive"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/streaming"
	"github.com/stackql/stackql/internal/stackql/tableinsertioncontainer"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
	"github.com/stackql/stackql/internal/stackql/util"
)

// sqlDataSourceSingleSelectAcquire implements the Builder interface
// and represents the action of acquiring data from an endpoint
// and then persisting that data into a table.
// This data would then subsequently be queried by later execution phases.
type sqlDataSourceSingleSelectAcquire struct {
	query                      string
	queryArgs                  []interface{}
	graph                      primitivegraph.PrimitiveGraph
	handlerCtx                 handler.HandlerContext
	tableMeta                  tablemetadata.ExtendedTableMetadata
	drmCfg                     drm.DRMConfig
	insertPreparedStatementCtx drm.PreparedStatementCtx
	insertionContainer         tableinsertioncontainer.TableInsertionContainer
	txnCtrlCtr                 internaldto.TxnControlCounters
	rowSort                    func(map[string]map[string]interface{}) []string
	root                       primitivegraph.PrimitiveNode
	stream                     streaming.MapStream
	sqlDataSource              sql_datasource.SQLDataSource
}

func NewSQLDataSourceSingleSelectAcquire(
	graph primitivegraph.PrimitiveGraph,
	handlerCtx handler.HandlerContext,
	insertionContainer tableinsertioncontainer.TableInsertionContainer,
	query string,
	queryArgs []interface{},
	insertCtx drm.PreparedStatementCtx,
	rowSort func(map[string]map[string]interface{}) []string,
	stream streaming.MapStream,
) Builder {
	tableMeta := insertionContainer.GetTableMetadata()
	return newSQLDataSourceSingleSelectAcquire(
		graph,
		handlerCtx,
		tableMeta,
		insertCtx,
		insertionContainer,
		query,
		queryArgs,
		rowSort,
		stream,
	)
}

func newSQLDataSourceSingleSelectAcquire(
	graph primitivegraph.PrimitiveGraph,
	handlerCtx handler.HandlerContext,
	tableMeta tablemetadata.ExtendedTableMetadata,
	insertCtx drm.PreparedStatementCtx,
	insertionContainer tableinsertioncontainer.TableInsertionContainer,
	query string,
	queryArgs []interface{},
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
	return &sqlDataSourceSingleSelectAcquire{
		graph:                      graph,
		handlerCtx:                 handlerCtx,
		tableMeta:                  tableMeta,
		rowSort:                    rowSort,
		drmCfg:                     handlerCtx.GetDrmConfig(),
		insertPreparedStatementCtx: insertCtx,
		insertionContainer:         insertionContainer,
		txnCtrlCtr:                 tcc,
		stream:                     stream,
		query:                      query,
		queryArgs:                  queryArgs,
	}
}

func (ss *sqlDataSourceSingleSelectAcquire) GetRoot() primitivegraph.PrimitiveNode {
	return ss.root
}

func (ss *sqlDataSourceSingleSelectAcquire) GetTail() primitivegraph.PrimitiveNode {
	return ss.root
}

func (ss *sqlDataSourceSingleSelectAcquire) Build() error {
	sqlDB, ok := ss.tableMeta.GetSQLDataSource()
	if !ok {
		return fmt.Errorf("sql data source unavailable for sql data source query")
	}
	_, err := ss.tableMeta.GetTableName()
	if err != nil {
		return err
	}
	ex := func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
		rows, err := sqlDB.Query(ss.query, ss.queryArgs...)
		if err != nil {
			return internaldto.NewErroneousExecutorOutput(err)
		}
		currentTcc := ss.insertPreparedStatementCtx.GetGCCtrlCtrs().Clone()
		ss.graph.AddTxnControlCounters(currentTcc)

		for {
			// TODO: fix cloning ops
			ok := rows.Next()
			if !ok {
				break
			}
			reqEncoding := "dummy_encoding"
			var item map[string]interface{}
			_, err := ss.drmCfg.ExecuteInsertDML(ss.handlerCtx.GetSQLEngine(), ss.insertPreparedStatementCtx, item, reqEncoding)
			if err != nil {
				return util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, nil, nil, ss.rowSort, err, nil))
			}

		}
		return internaldto.ExecutorOutput{}
	}

	prep := func() drm.PreparedStatementCtx {
		return ss.insertPreparedStatementCtx
	}
	insertPrim := primitive.NewHTTPRestPrimitive(
		nil,
		ex,
		prep,
		ss.txnCtrlCtr,
	)
	graph := ss.graph
	insertNode := graph.CreatePrimitiveNode(insertPrim)
	ss.root = insertNode

	return nil
}
