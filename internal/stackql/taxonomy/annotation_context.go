package taxonomy

import (
	"github.com/stackql/go-openapistackql/openapistackql"
	"github.com/stackql/stackql/internal/stackql/dto"
)

type AnnotationCtx interface {
	GetHIDs() *dto.HeirarchyIdentifiers
	GetParameters() map[string]interface{}
	GetSchema() *openapistackql.Schema
	GetTableMeta() *ExtendedTableMetadata
}

type StandardAnnotationCtx struct {
	Schema     *openapistackql.Schema
	HIDs       *dto.HeirarchyIdentifiers
	TableMeta  *ExtendedTableMetadata
	Parameters map[string]interface{}
}

func NewStandardAnnotationCtx(
	schema *openapistackql.Schema,
	hIds *dto.HeirarchyIdentifiers,
	tableMeta *ExtendedTableMetadata,
	parameters map[string]interface{},
) AnnotationCtx {
	return &StandardAnnotationCtx{
		Schema:     schema,
		HIDs:       hIds,
		TableMeta:  tableMeta,
		Parameters: parameters,
	}
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
