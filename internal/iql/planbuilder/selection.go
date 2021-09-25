package planbuilder

import (
	"fmt"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/handler"
	"infraql/internal/iql/metadata"
	"infraql/internal/iql/parserutil"
	"infraql/internal/iql/taxonomy"
	"infraql/internal/iql/util"

	log "github.com/sirupsen/logrus"
	"vitess.io/vitess/go/vt/sqlparser"
)

func (p *primitiveGenerator) assembleUnarySelectionBuilder(
	handlerCtx *handler.HandlerContext,
	node sqlparser.SQLNode,
	rewrittenWhere *sqlparser.Where,
	hIds *dto.HeirarchyIdentifiers,
	schema *metadata.Schema,
	tbl *taxonomy.ExtendedTableMetadata,
	selectTabulation *metadata.Tabulation,
	insertTabulation *metadata.Tabulation,
	cols []parserutil.ColumnHandle,
) error {
	annotatedInsertTabulation := util.NewAnnotatedTabulation(insertTabulation, hIds)
	tableDTO, err := p.PrimitiveBuilder.GetDRMConfig().GetCurrentTable(hIds, handlerCtx.SQLEngine)
	if err != nil {
		return err
	}

	method, err := tbl.GetMethod()
	if err != nil {
		return err
	}

	insPsc, err := p.PrimitiveBuilder.GetDRMConfig().GenerateInsertDML(annotatedInsertTabulation, p.PrimitiveBuilder.GetTxnCounterManager(), tableDTO.GetDiscoveryID())
	if err != nil {
		return err
	}
	p.PrimitiveBuilder.SetTxnCtrlCtrs(insPsc.TxnCtrlCtrs)
	for _, col := range cols {
		foundSchema := schema.FindByPath(col.Name, nil)
		cc, ok := method.Parameters[col.Name]
		if ok && cc.ID == col.Name {
			continue
		}
		if foundSchema == nil && col.IsColumn {
			return fmt.Errorf("column = '%s' is NOT present in either:  - data returned from provider, - acceptable parameters, use the DESCRIBE command to view available fields for SELECT operations", col.Name)
		}
		selectTabulation.PushBackColumn(metadata.NewColumnDescriptor(col.Alias, col.Name, col.DecoratedColumn, foundSchema, col.Val))
		log.Infoln(fmt.Sprintf("rsc = %T", col))
		log.Infoln(fmt.Sprintf("schema type = %T", schema))
	}

	selPsc, err := p.PrimitiveBuilder.GetDRMConfig().GenerateSelectDML(util.NewAnnotatedTabulation(selectTabulation, hIds), insPsc.TxnCtrlCtrs, node, rewrittenWhere)
	if err != nil {
		return err
	}
	p.PrimitiveBuilder.SetInsertPreparedStatementCtx(&insPsc)
	p.PrimitiveBuilder.SetSelectPreparedStatementCtx(&selPsc)
	p.PrimitiveBuilder.SetColumnOrder(cols)
	return nil
}

func (p *primitiveGenerator) analyzeUnarySelection(
	handlerCtx *handler.HandlerContext,
	node sqlparser.SQLNode,
	rewrittenWhere *sqlparser.Where,
	tbl *taxonomy.ExtendedTableMetadata,
	cols []parserutil.ColumnHandle) error {
	_, err := tbl.GetProvider()
	if err != nil {
		return err
	}
	schema, err := tbl.GetResponseSchema()
	if err != nil {
		return err
	}
	provStr, _ := tbl.GetProviderStr()
	svcStr, _ := tbl.GetServiceStr()
	unsuitableSchemaMsg := "schema unsuitable for select query"
	log.Infoln(fmt.Sprintf("schema.ID = %v", schema.ID))
	log.Infoln(fmt.Sprintf("schema.Items = %v", schema.Items))
	log.Infoln(fmt.Sprintf("schema.Properties = %v", schema.Properties))
	var itemS *metadata.Schema
	itemS, tbl.SelectItemsKey = schema.GetSelectListItems(tbl.LookupSelectItemsKey())
	if itemS == nil {
		return fmt.Errorf(unsuitableSchemaMsg)
	}
	is := itemS.Items
	itemObjS, _ := is.GetSchema(schema.SchemaCentral)
	if itemObjS == nil {
		return fmt.Errorf(unsuitableSchemaMsg)
	}
	if len(cols) == 0 {
		colNames := itemObjS.GetAllColumns()
		for _, v := range colNames {
			cols = append(cols, parserutil.NewUnaliasedColumnHandle(v))
		}
	}
	insertTabulation := itemObjS.Tabulate(false)

	hIds := dto.NewHeirarchyIdentifiers(provStr, svcStr, insertTabulation.GetName(), "")
	selectTabulation := itemObjS.Tabulate(true)

	return p.assembleUnarySelectionBuilder(
		handlerCtx,
		node,
		rewrittenWhere,
		hIds,
		schema,
		tbl,
		selectTabulation,
		insertTabulation,
		cols,
	)
}
