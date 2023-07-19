package builder_input //nolint:revive,stylecheck // permissable deviation from norm

import (
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
)

var (
	_ BuilderInput = &builderInput{}
)

type BuilderInput interface {
	GetGraphHolder() primitivegraph.PrimitiveGraphHolder
	GetHandlerContext() handler.HandlerContext
	GetParamMap() map[int]map[string]interface{}
	GetTableMetadata() tablemetadata.ExtendedTableMetadata
	GetDependencyNode() primitivegraph.PrimitiveNode
	GetCommentDirectives() sqlparser.CommentDirectives
	IsAwait() bool
	GetVerb() string
	GetInputAlias() string
	IsUndo() bool
	SetInputAlias(inputAlias string)
	SetIsAwait(isAwait bool)
	SetCommentDirectives(commentDirectives sqlparser.CommentDirectives)
	SetIsUndo(isUndo bool)
}

type builderInput struct {
	graphHolder       primitivegraph.PrimitiveGraphHolder
	handlerCtx        handler.HandlerContext
	paramMap          map[int]map[string]interface{}
	tbl               tablemetadata.ExtendedTableMetadata
	dependencyNode    primitivegraph.PrimitiveNode
	commentDirectives sqlparser.CommentDirectives
	isAwait           bool
	verb              string
	inputAlias        string
	isUndo            bool
}

func NewBuilderInput(
	graphHolder primitivegraph.PrimitiveGraphHolder,
	handlerCtx handler.HandlerContext,
	paramMap map[int]map[string]interface{},
	tbl tablemetadata.ExtendedTableMetadata,
	verb string,
) BuilderInput {
	return &builderInput{
		graphHolder:       graphHolder,
		handlerCtx:        handlerCtx,
		paramMap:          paramMap,
		tbl:               tbl,
		verb:              verb,
		commentDirectives: sqlparser.CommentDirectives{},
		inputAlias:        "", // this default is explicit for emphasisis
	}
}

func (bi *builderInput) GetGraphHolder() primitivegraph.PrimitiveGraphHolder {
	return bi.graphHolder
}

func (bi *builderInput) GetHandlerContext() handler.HandlerContext {
	return bi.handlerCtx
}

func (bi *builderInput) GetParamMap() map[int]map[string]interface{} {
	return bi.paramMap
}

func (bi *builderInput) GetTableMetadata() tablemetadata.ExtendedTableMetadata {
	return bi.tbl
}

func (bi *builderInput) GetDependencyNode() primitivegraph.PrimitiveNode {
	return bi.dependencyNode
}

func (bi *builderInput) GetCommentDirectives() sqlparser.CommentDirectives {
	return bi.commentDirectives
}

func (bi *builderInput) IsAwait() bool {
	return bi.isAwait
}

func (bi *builderInput) GetVerb() string {
	return bi.verb
}

func (bi *builderInput) GetInputAlias() string {
	return bi.inputAlias
}

func (bi *builderInput) IsUndo() bool {
	return bi.isUndo
}

func (bi *builderInput) SetGraphHolder(graphHolder primitivegraph.PrimitiveGraphHolder) {
	bi.graphHolder = graphHolder
}

func (bi *builderInput) SetHandlerContext(handlerCtx handler.HandlerContext) {
	bi.handlerCtx = handlerCtx
}

func (bi *builderInput) SetParamMap(paramMap map[int]map[string]interface{}) {
	bi.paramMap = paramMap
}

func (bi *builderInput) SetTableMetadata(tbl tablemetadata.ExtendedTableMetadata) {
	bi.tbl = tbl
}

func (bi *builderInput) SetDependencyNode(dependencyNode primitivegraph.PrimitiveNode) {
	bi.dependencyNode = dependencyNode
}

func (bi *builderInput) SetCommentDirectives(commentDirectives sqlparser.CommentDirectives) {
	bi.commentDirectives = commentDirectives
}

func (bi *builderInput) SetIsAwait(isAwait bool) {
	bi.isAwait = isAwait
}

func (bi *builderInput) SetVerb(verb string) {
	bi.verb = verb
}

func (bi *builderInput) SetInputAlias(inputAlias string) {
	bi.inputAlias = inputAlias
}

func (bi *builderInput) SetIsUndo(isUndo bool) {
	bi.isUndo = isUndo
}

func (bi *builderInput) Copy() BuilderInput {
	return &builderInput{
		graphHolder:       bi.graphHolder,
		handlerCtx:        bi.handlerCtx,
		paramMap:          bi.paramMap,
		tbl:               bi.tbl,
		dependencyNode:    bi.dependencyNode,
		commentDirectives: bi.commentDirectives,
		isAwait:           bi.isAwait,
		verb:              bi.verb,
		inputAlias:        bi.inputAlias,
		isUndo:            bi.isUndo,
	}
}
