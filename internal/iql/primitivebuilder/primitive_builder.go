package primitivebuilder

import (
	"infraql/internal/iql/drm"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/iqlmodel"
	"infraql/internal/iql/metadata"
	"infraql/internal/iql/parserutil"
	"infraql/internal/iql/primitive"
	"infraql/internal/iql/primitivegraph"
	"infraql/internal/iql/provider"
	"infraql/internal/iql/symtab"
	"infraql/internal/iql/taxonomy"

	"infraql/internal/pkg/txncounter"

	"vitess.io/vitess/go/vt/sqlparser"
)

type PrimitiveBuilder struct {
	await bool

	ast sqlparser.Statement

	builder Builder

	graph *primitivegraph.PrimitiveGraph

	drmConfig drm.DRMConfig

	// needed globally for non-heirarchy queries, such as "SHOW SERVICES FROM google;"
	prov            provider.IProvider
	tableFilter     func(iqlmodel.ITable) (iqlmodel.ITable, error)
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

	// per query -- SHOW INSERT only
	insertSchemaMap map[string]metadata.Schema

	// TODO: universally retire in favour of builder, which returns primitive.IPrimitive
	primitive primitive.IPrimitive

	symTab symtab.HashMapTreeSymTab

	where *sqlparser.Where
}

func (pb *PrimitiveBuilder) SetTxnCtrlCtrs(tc *dto.TxnControlCounters) {
	pb.txnCtrlCtrs = tc
}

func (pb *PrimitiveBuilder) GetGraph() *primitivegraph.PrimitiveGraph {
	return pb.graph
}

func (pb *PrimitiveBuilder) GetSymbol(k interface{}) (symtab.SymTabEntry, error) {
	return pb.symTab.GetSymbol(k)
}

func (pb *PrimitiveBuilder) SetSymbol(k interface{}, v symtab.SymTabEntry) error {
	return pb.symTab.SetSymbol(k, v)
}

func (pb *PrimitiveBuilder) GetWhere() *sqlparser.Where {
	return pb.where
}

func (pb *PrimitiveBuilder) SetWhere(where *sqlparser.Where) {
	pb.where = where
}

func (pb *PrimitiveBuilder) SetLeaf(k interface{}, l symtab.SymTab) error {
	return pb.symTab.SetLeaf(k, l)
}

func (pb *PrimitiveBuilder) GetAst() sqlparser.Statement {
	return pb.ast
}

func (pb *PrimitiveBuilder) GetInsertSchemaMap() map[string]metadata.Schema {
	return pb.insertSchemaMap
}

func (pb *PrimitiveBuilder) GetTxnCounterManager() *txncounter.TxnCounterManager {
	return pb.txnCounterManager
}

func (pb *PrimitiveBuilder) GetQuery() string {
	if pb.builder != nil {
		return pb.builder.GetQuery()
	}
	return ""
}

func (pb *PrimitiveBuilder) SetInsertSchemaMap(m map[string]metadata.Schema) {
	pb.insertSchemaMap = m
}

func (pb *PrimitiveBuilder) GetInsertValOnlyRows() map[int]map[int]interface{} {
	return pb.insertValOnlyRows
}

func (pb *PrimitiveBuilder) SetInsertValOnlyRows(m map[int]map[int]interface{}) {
	pb.insertValOnlyRows = m
}

func (pb *PrimitiveBuilder) GetColumnOrder() []string {
	return pb.columnOrder
}

func (pb *PrimitiveBuilder) SetColumnOrder(co []parserutil.ColumnHandle) {
	var colOrd []string
	for _, v := range co {
		colOrd = append(colOrd, v.Name)
	}
	pb.columnOrder = colOrd
}

func (pb *PrimitiveBuilder) GetPrimitive() primitive.IPrimitive {
	return pb.primitive
}

func (pb *PrimitiveBuilder) SetPrimitive(primitive primitive.IPrimitive) {
	pb.primitive = primitive
}

func (pb *PrimitiveBuilder) GetCommentDirectives() sqlparser.CommentDirectives {
	return pb.commentDirectives
}

func (pb *PrimitiveBuilder) SetCommentDirectives(dirs sqlparser.CommentDirectives) {
	pb.commentDirectives = dirs
}

func (pb *PrimitiveBuilder) GetLikeAbleColumns() []string {
	return pb.likeAbleColumns
}

func (pb *PrimitiveBuilder) SetLikeAbleColumns(cols []string) {
	pb.likeAbleColumns = cols
}

func (pb *PrimitiveBuilder) GetValOnlyColKeys() []int {
	keys := make([]int, 0, len(pb.valOnlyCols))
	for k := range pb.valOnlyCols {
		keys = append(keys, k)
	}
	return keys
}

func (pb *PrimitiveBuilder) GetValOnlyCol(key int) map[string]interface{} {
	return pb.valOnlyCols[key]
}

func (pb *PrimitiveBuilder) SetValOnlyCols(m map[int]map[string]interface{}) {
	pb.valOnlyCols = m
}

func (pb *PrimitiveBuilder) SetColVisited(colname string, isVisited bool) {
	pb.colsVisited[colname] = isVisited
}

func (pb *PrimitiveBuilder) GetTableFilter() func(iqlmodel.ITable) (iqlmodel.ITable, error) {
	return pb.tableFilter
}

func (pb *PrimitiveBuilder) SetTableFilter(tableFilter func(iqlmodel.ITable) (iqlmodel.ITable, error)) {
	pb.tableFilter = tableFilter
}

func (pb *PrimitiveBuilder) SetInsertPreparedStatementCtx(ctx *drm.PreparedStatementCtx) {
	pb.insertPreparedStatementCtx = ctx
}

func (pb *PrimitiveBuilder) GetInsertPreparedStatementCtx() *drm.PreparedStatementCtx {
	return pb.insertPreparedStatementCtx
}

func (pb *PrimitiveBuilder) SetSelectPreparedStatementCtx(ctx *drm.PreparedStatementCtx) {
	pb.selectPreparedStatementCtx = ctx
}

func (pb *PrimitiveBuilder) GetSelectPreparedStatementCtx() *drm.PreparedStatementCtx {
	return pb.selectPreparedStatementCtx
}

func (pb *PrimitiveBuilder) GetProvider() provider.IProvider {
	return pb.prov
}

func (pb *PrimitiveBuilder) SetProvider(prov provider.IProvider) {
	pb.prov = prov
}

func (pb *PrimitiveBuilder) GetBuilder() Builder {
	return pb.builder
}

func (pb *PrimitiveBuilder) SetBuilder(builder Builder) {
	pb.builder = builder
}

func (pb *PrimitiveBuilder) IsAwait() bool {
	return pb.await
}

func (pb *PrimitiveBuilder) SetAwait(await bool) {
	pb.await = await
}

func (pb PrimitiveBuilder) GetTable(node sqlparser.SQLNode) (taxonomy.ExtendedTableMetadata, error) {
	return pb.tables.GetTable(node)
}

func (pb PrimitiveBuilder) SetTable(node sqlparser.SQLNode, table taxonomy.ExtendedTableMetadata) {
	pb.tables.SetTable(node, table)
}

func (pb PrimitiveBuilder) GetTables() taxonomy.TblMap {
	return pb.tables
}

func (pb PrimitiveBuilder) GetDRMConfig() drm.DRMConfig {
	return pb.drmConfig
}

type HTTPRestPrimitive struct {
	Provider             provider.IProvider
	Executor             func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput
	Preparator           func() *drm.PreparedStatementCtx
	TxnControlCtr        *dto.TxnControlCounters
	Inputs               map[int64]dto.ExecutorOutput
	InputAliases         map[string]int64
	id                   int64
	cachedExecutorOutout *dto.ExecutorOutput
}

type MetaDataPrimitive struct {
	Provider   provider.IProvider
	Executor   func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput
	Preparator func() *drm.PreparedStatementCtx
	id         int64
}

type LocalPrimitive struct {
	Executor   func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput
	Preparator func() *drm.PreparedStatementCtx
	id         int64
}

func (pr *HTTPRestPrimitive) SetTxnId(id int) {
	if pr.TxnControlCtr != nil {
		pr.TxnControlCtr.TxnId = id
	}
}

func (pr *MetaDataPrimitive) SetTxnId(id int) {
}

func (pr *LocalPrimitive) SetTxnId(id int) {
}

func (pr *HTTPRestPrimitive) IncidentData(fromId int64, input dto.ExecutorOutput) error {
	pr.Inputs[fromId] = input
	return nil
}

func (pr *MetaDataPrimitive) IncidentData(fromId int64, input dto.ExecutorOutput) error {
	return nil
}

func (pr *LocalPrimitive) IncidentData(fromId int64, input dto.ExecutorOutput) error {
	return nil
}

func (pr *HTTPRestPrimitive) SetInputAlias(alias string, id int64) error {
	pr.InputAliases[alias] = id
	return nil
}

func (pr *MetaDataPrimitive) SetInputAlias(alias string, id int64) error {
	return nil
}

func (pr *LocalPrimitive) SetInputAlias(alias string, id int64) error {
	return nil
}

func (pr *HTTPRestPrimitive) Optimise() error {
	return nil
}

func (pr *MetaDataPrimitive) Optimise() error {
	return nil
}

func (pr *LocalPrimitive) Optimise() error {
	return nil
}

func (pr *HTTPRestPrimitive) Execute(pc primitive.IPrimitiveCtx) dto.ExecutorOutput {
	if pr.cachedExecutorOutout != nil {
		return *(pr.cachedExecutorOutout)
	}
	if pr.Executor != nil {
		op := pr.Executor(pc)
		pr.cachedExecutorOutout = &op
		return op
	}
	return dto.NewExecutorOutput(nil, nil, nil, nil, nil)
}

func (pr *HTTPRestPrimitive) ID() int64 {
	return pr.id
}

func (pr *MetaDataPrimitive) ID() int64 {
	return pr.id
}

func (pr *LocalPrimitive) ID() int64 {
	return pr.id
}

func (pr *MetaDataPrimitive) Execute(pc primitive.IPrimitiveCtx) dto.ExecutorOutput {
	if pr.Executor != nil {
		return pr.Executor(pc)
	}
	return dto.NewExecutorOutput(nil, nil, nil, nil, nil)
}

func (pr *LocalPrimitive) Execute(pc primitive.IPrimitiveCtx) dto.ExecutorOutput {
	if pr.Executor != nil {
		return pr.Executor(pc)
	}
	return dto.NewExecutorOutput(nil, nil, nil, nil, nil)
}

func NewMetaDataPrimitive(provider provider.IProvider, executor func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput) *MetaDataPrimitive {
	return &MetaDataPrimitive{
		Provider: provider,
		Executor: executor,
	}
}

func NewHTTPRestPrimitive(provider provider.IProvider, executor func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput, preparator func() *drm.PreparedStatementCtx, txnCtrlCtr *dto.TxnControlCounters) *HTTPRestPrimitive {
	return &HTTPRestPrimitive{
		Provider:      provider,
		Executor:      executor,
		Preparator:    preparator,
		TxnControlCtr: txnCtrlCtr,
		Inputs:        make(map[int64]dto.ExecutorOutput),
		InputAliases:  make(map[string]int64),
	}
}

func NewLocalPrimitive(executor func(pc primitive.IPrimitiveCtx) dto.ExecutorOutput) *LocalPrimitive {
	return &LocalPrimitive{
		Executor: executor,
	}
}

func NewPrimitiveBuilder(ast sqlparser.Statement, drmConfig drm.DRMConfig, txnCtrMgr *txncounter.TxnCounterManager, graph *primitivegraph.PrimitiveGraph) *PrimitiveBuilder {
	return &PrimitiveBuilder{
		ast:               ast,
		drmConfig:         drmConfig,
		tables:            make(map[sqlparser.SQLNode]taxonomy.ExtendedTableMetadata),
		valOnlyCols:       make(map[int]map[string]interface{}),
		insertValOnlyRows: make(map[int]map[int]interface{}),
		colsVisited:       make(map[string]bool),
		txnCounterManager: txnCtrMgr,
		symTab:            symtab.NewHashMapTreeSymTab(),
		graph:             graph,
	}
}
