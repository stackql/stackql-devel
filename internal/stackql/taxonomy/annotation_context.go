package taxonomy

import (
	"fmt"

	"github.com/stackql/go-openapistackql/openapistackql"
	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/httpbuild"
	"github.com/stackql/stackql/internal/stackql/provider"
	"github.com/stackql/stackql/internal/stackql/util"
)

type AnnotationCtx interface {
	GetHIDs() *dto.HeirarchyIdentifiers
	IsDynamic() bool
	GetParameters() map[string]interface{}
	GetSchema() *openapistackql.Schema
	GetTableMeta() *ExtendedTableMetadata
	Prepare(handlerCtx *handler.HandlerContext, pr provider.IProvider, opStore *openapistackql.OperationStore, svc *openapistackql.Service) error
}

type StandardAnnotationCtx struct {
	isDynamic  bool
	Schema     *openapistackql.Schema
	HIDs       *dto.HeirarchyIdentifiers
	TableMeta  *ExtendedTableMetadata
	Parameters map[string]interface{}
}

func NewStaticStandardAnnotationCtx(
	schema *openapistackql.Schema,
	hIds *dto.HeirarchyIdentifiers,
	tableMeta *ExtendedTableMetadata,
	parameters map[string]interface{},
) AnnotationCtx {
	return &StandardAnnotationCtx{
		isDynamic:  false,
		Schema:     schema,
		HIDs:       hIds,
		TableMeta:  tableMeta,
		Parameters: parameters,
	}
}

func (ac *StandardAnnotationCtx) IsDynamic() bool {
	return ac.isDynamic
}

func (ac *StandardAnnotationCtx) Prepare(
	handlerCtx *handler.HandlerContext,
	pr provider.IProvider,
	opStore *openapistackql.OperationStore,
	svc *openapistackql.Service,
) error {
	if ac.isDynamic {
		return fmt.Errorf("dynamic parameterinference not yet supported")
	}
	parametersCleaned, err := util.TransformSQLRawParameters(ac.GetParameters())
	if err != nil {
		return err
	}
	httpArmoury, err := httpbuild.BuildHTTPRequestCtxFromAnnotation(handlerCtx, parametersCleaned, pr, opStore, svc, nil, nil)
	if err != nil {
		return err
	}
	ac.TableMeta.HttpArmoury = httpArmoury
	return nil
}

func (ac *StandardAnnotationCtx) GetHIDs() *dto.HeirarchyIdentifiers {
	return ac.HIDs
}

func (ac *StandardAnnotationCtx) GetParameters() map[string]interface{} {
	return ac.Parameters
}

func (ac *StandardAnnotationCtx) GetSchema() *openapistackql.Schema {
	return ac.Schema
}

func (ac *StandardAnnotationCtx) GetTableMeta() *ExtendedTableMetadata {
	return ac.TableMeta
}
