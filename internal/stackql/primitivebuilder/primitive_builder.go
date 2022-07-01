package primitivebuilder

import (
	"github.com/stackql/stackql/internal/stackql/drm"
	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/parserutil"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/provider"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
	"github.com/stackql/stackql/internal/stackql/symtab"
	"github.com/stackql/stackql/internal/stackql/taxonomy"

	"github.com/stackql/go-openapistackql/openapistackql"

	"github.com/stackql/stackql/pkg/txncounter"

	"vitess.io/vitess/go/vt/sqlparser"
)

type PrimitiveBuilder interface {
	AddChild(val PrimitiveBuilder)
	GetAst() sqlparser.SQLNode
	GetBuilder() Builder
	GetChildren() []PrimitiveBuilder
	GetColumnOrder() []string
	GetCommentDirectives() sqlparser.CommentDirectives
	GetDRMConfig() drm.DRMConfig
	GetGraph() *primitivegraph.PrimitiveGraph
	GetInsertPreparedStatementCtx() *drm.PreparedStatementCtx
	GetInsertValOnlyRows() map[int]map[int]interface{}
	GetLikeAbleColumns() []string
	GetParent() PrimitiveBuilder
	GetProvider() provider.IProvider
	GetRoot() primitivegraph.PrimitiveNode
	GetSelectPreparedStatementCtx() *drm.PreparedStatementCtx
	GetSQLEngine() sqlengine.SQLEngine
	GetSymbol(k interface{}) (symtab.SymTabEntry, error)
	GetSymTab() symtab.SymTab
	GetTable(node sqlparser.SQLNode) (*taxonomy.ExtendedTableMetadata, error)
	GetTableFilter() func(openapistackql.ITable) (openapistackql.ITable, error)
	GetTables() taxonomy.TblMap
	GetTxnCounterManager() *txncounter.TxnCounterManager
	GetTxnCtrlCtrs() *dto.TxnControlCounters
	GetValOnlyCol(key int) map[string]interface{}
	GetValOnlyColKeys() []int
	GetWhere() *sqlparser.Where
	IsAwait() bool
	NewChildPrimitiveBuilder(ast sqlparser.SQLNode) PrimitiveBuilder
	SetAwait(await bool)
	SetBuilder(builder Builder)
	SetColumnOrder(co []parserutil.ColumnHandle)
	SetColVisited(colname string, isVisited bool)
	SetCommentDirectives(dirs sqlparser.CommentDirectives)
	SetInsertPreparedStatementCtx(ctx *drm.PreparedStatementCtx)
	SetInsertValOnlyRows(m map[int]map[int]interface{})
	SetLikeAbleColumns(cols []string)
	SetProvider(prov provider.IProvider)
	SetRoot(root primitivegraph.PrimitiveNode)
	SetSelectPreparedStatementCtx(ctx *drm.PreparedStatementCtx)
	SetSymbol(k interface{}, v symtab.SymTabEntry) error
	SetTable(node sqlparser.SQLNode, table *taxonomy.ExtendedTableMetadata)
	SetTableFilter(tableFilter func(openapistackql.ITable) (openapistackql.ITable, error))
	SetTxnCtrlCtrs(tc *dto.TxnControlCounters)
	SetValOnlyCols(m map[int]map[string]interface{})
	SetWhere(where *sqlparser.Where)
	ShouldCollectGarbage() bool
}

type StandardPrimitiveBuilder struct {
	parent PrimitiveBuilder

	children []PrimitiveBuilder

	await bool

	ast sqlparser.SQLNode

	builder Builder

	graph *primitivegraph.PrimitiveGraph

	drmConfig drm.DRMConfig

	// needed globally for non-heirarchy queries, such as "SHOW SERVICES FROM google;"
	prov            provider.IProvider
	tableFilter     func(openapistackql.ITable) (openapistackql.ITable, error)
	colsVisited     map[string]bool
	likeAbleColumns []string

	// per table
	tables taxonomy.TblMap

	// per query
	columnOrder       []string
	commentDirectives sqlparser.CommentDirectives
	txnCounterManager *txncounter.TxnCounterManager
	txnCtrlCtrs       *dto.TxnControlCounters

	// per query -- SELECT only
	insertValOnlyRows          map[int]map[int]interface{}
	valOnlyCols                map[int]map[string]interface{}
	insertPreparedStatementCtx *drm.PreparedStatementCtx
	selectPreparedStatementCtx *drm.PreparedStatementCtx

	// TODO: universally retire in favour of builder, which returns primitive.IPrimitive
	root primitivegraph.PrimitiveNode

	symTab symtab.SymTab

	where *sqlparser.Where

	sqlEngine sqlengine.SQLEngine
}

func (pb *StandardPrimitiveBuilder) ShouldCollectGarbage() bool {
	return pb.parent == nil
}

func (pb *StandardPrimitiveBuilder) SetTxnCtrlCtrs(tc *dto.TxnControlCounters) {
	pb.txnCtrlCtrs = tc
}

func (pb *StandardPrimitiveBuilder) GetTxnCtrlCtrs() *dto.TxnControlCounters {
	return pb.txnCtrlCtrs
}

func (pb *StandardPrimitiveBuilder) GetGraph() *primitivegraph.PrimitiveGraph {
	return pb.graph
}

func (pb *StandardPrimitiveBuilder) GetParent() PrimitiveBuilder {
	return pb.parent
}

func (pb *StandardPrimitiveBuilder) GetChildren() []PrimitiveBuilder {
	return pb.children
}

func (pb *StandardPrimitiveBuilder) AddChild(val PrimitiveBuilder) {
	pb.children = append(pb.children, val)
}

func (pb *StandardPrimitiveBuilder) GetSymbol(k interface{}) (symtab.SymTabEntry, error) {
	return pb.symTab.GetSymbol(k)
}

func (pb *StandardPrimitiveBuilder) GetSymTab() symtab.SymTab {
	return pb.symTab
}

func (pb *StandardPrimitiveBuilder) SetSymbol(k interface{}, v symtab.SymTabEntry) error {
	return pb.symTab.SetSymbol(k, v)
}

func (pb *StandardPrimitiveBuilder) GetWhere() *sqlparser.Where {
	return pb.where
}

func (pb *StandardPrimitiveBuilder) SetWhere(where *sqlparser.Where) {
	pb.where = where
}

func (pb *StandardPrimitiveBuilder) GetAst() sqlparser.SQLNode {
	return pb.ast
}

func (pb *StandardPrimitiveBuilder) GetTxnCounterManager() *txncounter.TxnCounterManager {
	return pb.txnCounterManager
}

func (pb *StandardPrimitiveBuilder) NewChildPrimitiveBuilder(ast sqlparser.SQLNode) PrimitiveBuilder {
	child := NewPrimitiveBuilder(pb, ast, pb.drmConfig, pb.txnCounterManager, pb.graph, pb.tables, pb.symTab, pb.sqlEngine)
	pb.children = append(pb.children, child)
	return child
}

func (pb *StandardPrimitiveBuilder) GetInsertValOnlyRows() map[int]map[int]interface{} {
	return pb.insertValOnlyRows
}

func (pb *StandardPrimitiveBuilder) SetInsertValOnlyRows(m map[int]map[int]interface{}) {
	pb.insertValOnlyRows = m
}

func (pb *StandardPrimitiveBuilder) GetColumnOrder() []string {
	return pb.columnOrder
}

func (pb *StandardPrimitiveBuilder) SetColumnOrder(co []parserutil.ColumnHandle) {
	var colOrd []string
	for _, v := range co {
		colOrd = append(colOrd, v.Name)
	}
	pb.columnOrder = colOrd
}

func (pb *StandardPrimitiveBuilder) GetRoot() primitivegraph.PrimitiveNode {
	return pb.root
}

func (pb *StandardPrimitiveBuilder) SetRoot(root primitivegraph.PrimitiveNode) {
	pb.root = root
}

func (pb *StandardPrimitiveBuilder) GetCommentDirectives() sqlparser.CommentDirectives {
	return pb.commentDirectives
}

func (pb *StandardPrimitiveBuilder) SetCommentDirectives(dirs sqlparser.CommentDirectives) {
	pb.commentDirectives = dirs
}

func (pb *StandardPrimitiveBuilder) GetLikeAbleColumns() []string {
	return pb.likeAbleColumns
}

func (pb *StandardPrimitiveBuilder) SetLikeAbleColumns(cols []string) {
	pb.likeAbleColumns = cols
}

func (pb *StandardPrimitiveBuilder) GetValOnlyColKeys() []int {
	keys := make([]int, 0, len(pb.valOnlyCols))
	for k := range pb.valOnlyCols {
		keys = append(keys, k)
	}
	return keys
}

func (pb *StandardPrimitiveBuilder) GetValOnlyCol(key int) map[string]interface{} {
	return pb.valOnlyCols[key]
}

func (pb *StandardPrimitiveBuilder) SetValOnlyCols(m map[int]map[string]interface{}) {
	pb.valOnlyCols = m
}

func (pb *StandardPrimitiveBuilder) SetColVisited(colname string, isVisited bool) {
	pb.colsVisited[colname] = isVisited
}

func (pb *StandardPrimitiveBuilder) GetTableFilter() func(openapistackql.ITable) (openapistackql.ITable, error) {
	return pb.tableFilter
}

func (pb *StandardPrimitiveBuilder) SetTableFilter(tableFilter func(openapistackql.ITable) (openapistackql.ITable, error)) {
	pb.tableFilter = tableFilter
}

func (pb *StandardPrimitiveBuilder) SetInsertPreparedStatementCtx(ctx *drm.PreparedStatementCtx) {
	pb.insertPreparedStatementCtx = ctx
}

func (pb *StandardPrimitiveBuilder) GetInsertPreparedStatementCtx() *drm.PreparedStatementCtx {
	return pb.insertPreparedStatementCtx
}

func (pb *StandardPrimitiveBuilder) SetSelectPreparedStatementCtx(ctx *drm.PreparedStatementCtx) {
	pb.selectPreparedStatementCtx = ctx
}

func (pb *StandardPrimitiveBuilder) GetSelectPreparedStatementCtx() *drm.PreparedStatementCtx {
	return pb.selectPreparedStatementCtx
}

func (pb *StandardPrimitiveBuilder) GetProvider() provider.IProvider {
	return pb.prov
}

func (pb *StandardPrimitiveBuilder) SetProvider(prov provider.IProvider) {
	pb.prov = prov
}

func (pb *StandardPrimitiveBuilder) GetBuilder() Builder {
	if pb.children == nil || len(pb.children) == 0 {
		return pb.builder
	}
	var builders []Builder
	for _, child := range pb.children {
		if bldr := child.GetBuilder(); bldr != nil {
			builders = append(builders, bldr)
		}
	}
	if true {
		return NewDiamondBuilder(pb.builder, builders, pb.graph, pb.sqlEngine, pb.ShouldCollectGarbage())
	}
	return NewSubTreeBuilder(builders)
}

func (pb *StandardPrimitiveBuilder) SetBuilder(builder Builder) {
	pb.builder = builder
}

func (pb *StandardPrimitiveBuilder) IsAwait() bool {
	return pb.await
}

func (pb *StandardPrimitiveBuilder) SetAwait(await bool) {
	pb.await = await
}

func (pb *StandardPrimitiveBuilder) GetTable(node sqlparser.SQLNode) (*taxonomy.ExtendedTableMetadata, error) {
	return pb.tables.GetTable(node)
}

func (pb *StandardPrimitiveBuilder) SetTable(node sqlparser.SQLNode, table *taxonomy.ExtendedTableMetadata) {
	pb.tables.SetTable(node, table)
}

func (pb *StandardPrimitiveBuilder) GetTables() taxonomy.TblMap {
	return pb.tables
}

func (pb *StandardPrimitiveBuilder) GetDRMConfig() drm.DRMConfig {
	return pb.drmConfig
}

func (pb *StandardPrimitiveBuilder) GetSQLEngine() sqlengine.SQLEngine {
	return pb.sqlEngine
}

func NewPrimitiveBuilder(parent PrimitiveBuilder, ast sqlparser.SQLNode, drmConfig drm.DRMConfig, txnCtrMgr *txncounter.TxnCounterManager, graph *primitivegraph.PrimitiveGraph, tblMap taxonomy.TblMap, symTab symtab.SymTab, sqlEngine sqlengine.SQLEngine) PrimitiveBuilder {
	return &StandardPrimitiveBuilder{
		parent:            parent,
		ast:               ast,
		drmConfig:         drmConfig,
		tables:            tblMap,
		valOnlyCols:       make(map[int]map[string]interface{}),
		insertValOnlyRows: make(map[int]map[int]interface{}),
		colsVisited:       make(map[string]bool),
		txnCounterManager: txnCtrMgr,
		symTab:            symTab,
		graph:             graph,
		sqlEngine:         sqlEngine,
	}
}
