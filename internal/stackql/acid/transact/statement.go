package transact

import (
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/querysubmit"
)

type Statement interface {
	Prepare() error
	Execute() internaldto.ExecutorOutput
}

type basicStatement struct {
	query          string
	handlerCtx     handler.HandlerContext
	querySubmitter querysubmit.QuerySubmitter
}

func NewStatement(query string, handlerCtx handler.HandlerContext) Statement {
	return &basicStatement{
		query:          query,
		handlerCtx:     handlerCtx,
		querySubmitter: querysubmit.NewQuerySubmitter(),
	}
}

func (st *basicStatement) Prepare() error {
	cmdString := st.query
	clonedCtx := st.handlerCtx.Clone()
	clonedCtx.SetQuery(cmdString)
	return st.querySubmitter.PrepareQuery(clonedCtx)
}

func (st *basicStatement) Execute() internaldto.ExecutorOutput {
	return st.querySubmitter.SubmitQuery()
}
