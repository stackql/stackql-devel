package primitivebuilder

import (
	"fmt"
	"strings"

	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"github.com/stackql/stackql/internal/stackql/astformat"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/parserutil"
	"github.com/stackql/stackql/internal/stackql/primitive"
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/util"
)

type ddl struct {
	graph      primitivegraph.PrimitiveGraphHolder
	ddlObject  *sqlparser.DDL
	handlerCtx handler.HandlerContext
	root, tail primitivegraph.PrimitiveNode
}

func (ddo *ddl) Build() error {
	sqlSystem := ddo.handlerCtx.GetSQLSystem()
	if sqlSystem == nil {
		return fmt.Errorf("cannot proceed DDL execution with nil sql system object")
	}
	parserDDLObj := ddo.ddlObject
	if sqlSystem == nil {
		return fmt.Errorf("cannot proceed DDL execution with nil ddl object")
	}
	unionEx := func(pc primitive.IPrimitiveCtx) internaldto.ExecutorOutput {
		actionLowered := strings.ToLower(parserDDLObj.Action)
		switch actionLowered {
		case "create":
			tableName := strings.Trim(astformat.String(parserDDLObj.Table, sqlSystem.GetASTFormatter()), `"`)
			isTable := parserutil.IsCreatePhysicalTable(parserDDLObj)
			isTempTable := parserutil.IsCreateTemporaryPhysicalTable(parserDDLObj)
			isMaterializedView := parserutil.IsCreateMaterializedView(parserDDLObj)
			if isTable || isTempTable { // TODO: support for create tables
				return internaldto.NewErroneousExecutorOutput(fmt.Errorf("create table is not supported"))
			}
			if isMaterializedView { // TODO: support for create materialized views
				return internaldto.NewErroneousExecutorOutput(fmt.Errorf("create materialized view is not supported"))
			}
			relationDDL := strings.ReplaceAll(
				astformat.String(parserDDLObj.SelectStatement, astformat.DefaultSelectExprsFormatter), `"`, "")
			err := sqlSystem.CreateView(tableName, relationDDL, parserDDLObj.OrReplace)
			if err != nil {
				return internaldto.NewErroneousExecutorOutput(err)
			}
		case "drop":
			if tl := len(parserDDLObj.FromTables); tl != 1 {
				return internaldto.NewErroneousExecutorOutput(fmt.Errorf("cannot drop table with supplied table count = %d", tl))
			}
			tableName := strings.Trim(astformat.String(parserDDLObj.FromTables[0], sqlSystem.GetASTFormatter()), `"`)
			err := sqlSystem.DropView(tableName)
			if err != nil {
				return internaldto.NewErroneousExecutorOutput(err)
			}
		default:
		}
		return util.PrepareResultSet(
			internaldto.NewPrepareResultSetPlusRawDTO(
				nil,
				map[string]map[string]interface{}{"0": {"message": "DDL execution completed"}},
				[]string{"message"},
				nil,
				nil,
				nil,
				nil,
				ddo.handlerCtx.GetTypingConfig(),
			),
		)
	}
	graph := ddo.graph
	unionNode := graph.CreatePrimitiveNode(primitive.NewLocalPrimitive(unionEx))
	ddo.root = unionNode
	ddo.tail = unionNode
	return nil
}

func NewDDL(
	graph primitivegraph.PrimitiveGraphHolder,
	handlerCtx handler.HandlerContext,
	ddlObject *sqlparser.DDL,
) Builder {
	return &ddl{
		graph:      graph,
		handlerCtx: handlerCtx,
		ddlObject:  ddlObject,
	}
}

func (ddo *ddl) GetRoot() primitivegraph.PrimitiveNode {
	return ddo.root
}

func (ddo *ddl) GetTail() primitivegraph.PrimitiveNode {
	return ddo.tail
}
