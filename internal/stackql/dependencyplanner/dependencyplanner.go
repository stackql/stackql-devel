package dependencyplanner

import (
	"github.com/stackql/go-openapistackql/pkg/media"
	"github.com/stackql/stackql/internal/stackql/astvisit"
	"github.com/stackql/stackql/internal/stackql/docparser"
	"github.com/stackql/stackql/internal/stackql/drm"
	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/httpbuild"
	"github.com/stackql/stackql/internal/stackql/parserutil"
	"github.com/stackql/stackql/internal/stackql/primitivebuilder"
	"github.com/stackql/stackql/internal/stackql/primitivecomposer"
	"github.com/stackql/stackql/internal/stackql/taxonomy"
	"github.com/stackql/stackql/internal/stackql/util"
	"vitess.io/vitess/go/vt/sqlparser"

	log "github.com/sirupsen/logrus"
)

type DependencyPlanner interface {
	Plan() error
	GetBldr() primitivebuilder.Builder
	GetSelectCtx() *drm.PreparedStatementCtx
}

type StandardDependencyPlanner struct {
	annotations       taxonomy.AnnotationCtxMap
	colRefs           parserutil.ColTableMap
	handlerCtx        *handler.HandlerContext
	execSlice         []primitivebuilder.Builder
	primaryTcc, tcc   *dto.TxnControlCounters
	primitiveComposer primitivecomposer.PrimitiveComposer
	rewrittenWhere    *sqlparser.Where
	secondaryTccs     []*dto.TxnControlCounters
	sqlStatement      sqlparser.SQLNode
	tableSlice        []*taxonomy.ExtendedTableMetadata
	tblz              taxonomy.TblMap
	discoGenIDs       map[sqlparser.SQLNode]int

	//
	bldr   primitivebuilder.Builder
	selCtx *drm.PreparedStatementCtx
}

func NewStandardDependencyPlanner(
	handlerCtx *handler.HandlerContext,
	annotations taxonomy.AnnotationCtxMap,
	colRefs parserutil.ColTableMap,
	rewrittenWhere *sqlparser.Where,
	sqlStatement sqlparser.SQLNode,
	tblz taxonomy.TblMap,
	primitiveComposer primitivecomposer.PrimitiveComposer,
) DependencyPlanner {
	return &StandardDependencyPlanner{
		handlerCtx:        handlerCtx,
		annotations:       annotations,
		colRefs:           colRefs,
		rewrittenWhere:    rewrittenWhere,
		sqlStatement:      sqlStatement,
		tblz:              tblz,
		primitiveComposer: primitiveComposer,
		discoGenIDs:       make(map[sqlparser.SQLNode]int),
	}
}

func (dp *StandardDependencyPlanner) GetBldr() primitivebuilder.Builder {
	return dp.bldr
}

func (dp *StandardDependencyPlanner) GetSelectCtx() *drm.PreparedStatementCtx {
	return dp.selCtx
}

func (dp *StandardDependencyPlanner) Plan() error {
	// BLOCK ANNOTATION_TRAVERSE
	// TODO: annotations need to be ordered
	//       and data dependencies need to be modelled.
	for k, va := range dp.annotations {
		pr, err := va.GetTableMeta().GetProvider()
		if err != nil {
			return err
		}
		prov, err := va.GetTableMeta().GetProviderObject()
		if err != nil {
			return err
		}
		svc, err := va.GetTableMeta().GetService()
		if err != nil {
			return err
		}
		m, err := va.GetTableMeta().GetMethod()
		if err != nil {
			return err
		}
		tab := va.GetSchema().Tabulate(false)
		_, mediaType, err := m.GetResponseBodySchemaAndMediaType()
		if err != nil {
			return err
		}
		switch mediaType {
		case media.MediaTypeTextXML, media.MediaTypeXML:
			tab = tab.RenameColumnsToXml()
		}
		anTab := util.NewAnnotatedTabulation(tab, va.GetHIDs(), va.GetTableMeta().Alias)

		discoGenId, err := docparser.OpenapiStackQLTabulationsPersistor(prov, svc, []util.AnnotatedTabulation{anTab}, dp.primitiveComposer.GetSQLEngine(), prov.Name)
		if err != nil {
			return err
		}
		dp.discoGenIDs[k] = discoGenId
		parametersCleaned, err := util.TransformSQLRawParameters(va.GetParameters())
		if err != nil {
			return err
		}
		httpArmoury, err := httpbuild.BuildHTTPRequestCtxFromAnnotation(dp.handlerCtx, parametersCleaned, pr, m, svc, nil, nil)
		if err != nil {
			return err
		}
		va.GetTableMeta().HttpArmoury = httpArmoury
		tableDTO, err := dp.primitiveComposer.GetDRMConfig().GetCurrentTable(va.GetHIDs(), dp.handlerCtx.SQLEngine)
		if err != nil {
			return err
		}
		if dp.tcc == nil {
			dp.tcc = dto.NewTxnControlCounters(dp.primitiveComposer.GetTxnCounterManager(), tableDTO.GetDiscoveryID())
			dp.primaryTcc = dp.tcc
		} else {
			dp.tcc = dp.tcc.CloneAndIncrementInsertID()
			dp.tcc.DiscoveryGenerationId = discoGenId
			dp.secondaryTccs = append(dp.secondaryTccs, dp.tcc)
		}
		insPsc, err := dp.primitiveComposer.GetDRMConfig().GenerateInsertDML(anTab, dp.tcc)
		if err != nil {
			return err
		}
		builder := primitivebuilder.NewSingleSelectAcquire(dp.primitiveComposer.GetGraph(), dp.handlerCtx, va.GetTableMeta(), insPsc, nil)
		dp.execSlice = append(dp.execSlice, builder)
		dp.tableSlice = append(dp.tableSlice, va.GetTableMeta())
		// END_BLOCK ANNOTATION_TRAVERSE
	}
	rewrittenWhereStr := astvisit.GenerateModifiedWhereClause(dp.rewrittenWhere)
	log.Debugf("rewrittenWhereStr = '%s'", rewrittenWhereStr)
	v := astvisit.NewQueryRewriteAstVisitor(
		dp.handlerCtx,
		dp.tblz,
		dp.tableSlice,
		dp.annotations,
		dp.discoGenIDs,
		dp.colRefs,
		drm.GetGoogleV1SQLiteConfig(),
		dp.primaryTcc,
		dp.secondaryTccs,
		rewrittenWhereStr,
	)
	err := v.Visit(dp.sqlStatement)
	if err != nil {
		return err
	}
	selCtx, err := v.GenerateSelectDML()
	if err != nil {
		return err
	}
	selBld := primitivebuilder.NewSingleSelect(dp.primitiveComposer.GetGraph(), dp.handlerCtx, selCtx, nil)
	dp.bldr = primitivebuilder.NewMultipleAcquireAndSelect(dp.primitiveComposer.GetGraph(), dp.execSlice, selBld)
	dp.selCtx = selCtx
	return nil
}
