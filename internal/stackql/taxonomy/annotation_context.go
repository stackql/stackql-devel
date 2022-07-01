package taxonomy

import (
	"github.com/stackql/go-openapistackql/openapistackql"
	"github.com/stackql/stackql/internal/stackql/dto"
)

type AnnotationCtx struct {
	Schema     *openapistackql.Schema
	HIDs       *dto.HeirarchyIdentifiers
	TableMeta  *ExtendedTableMetadata
	Parameters map[string]interface{}
}
