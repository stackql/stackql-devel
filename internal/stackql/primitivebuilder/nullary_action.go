package primitivebuilder

import (
	"github.com/stackql/go-openapistackql/openapistackql"
	"github.com/stackql/stackql/internal/stackql/drm"
	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/taxonomy"
)

type NullaryAction struct {
	query      string
	handlerCtx *handler.HandlerContext
	tableMeta  taxonomy.ExtendedTableMetadata
	tabulation openapistackql.Tabulation
	drmCfg     drm.DRMConfig
	txnCtrlCtr *dto.TxnControlCounters
	root       primitivegraph.PrimitiveNode
}
