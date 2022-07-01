package primitivebuilder

import (
	"fmt"

	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/primitive"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/util"
)

type Join struct {
	lhsPb, rhsPb *PrimitiveBuilder
	lhs, rhs     Builder
	handlerCtx   *handler.HandlerContext
	rowSort      func(map[string]map[string]interface{}) []string
}

func NewJoin(lhsPb *PrimitiveBuilder, rhsPb *PrimitiveBuilder, handlerCtx *handler.HandlerContext, rowSort func(map[string]map[string]interface{}) []string) *Join {
	return &Join{
		lhsPb:      lhsPb,
		rhsPb:      rhsPb,
		handlerCtx: handlerCtx,
		rowSort:    rowSort,
	}
}

func (j *Join) Build() error {
	return nil
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
