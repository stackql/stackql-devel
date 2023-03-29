package driver

import (
	"context"
	"fmt"
	"strings"

	"github.com/stackql/psql-wire/pkg/sqldata"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/logging"
	"github.com/stackql/stackql/internal/stackql/querysubmit"
	"github.com/stackql/stackql/internal/stackql/responsehandler"
	"github.com/stackql/stackql/internal/stackql/util"

	sqlbackend "github.com/stackql/psql-wire/pkg/sqlbackend"
)

var (
	_ StackQLDriver = &basicStackQLDriver{}
)

type StackQLDriver interface {
	sqlbackend.ISQLBackend
	ProcessDryRun(string)
	ProcessQuery(string)
}

func (dr *basicStackQLDriver) ProcessDryRun(query string) {
	resultMap := map[string]map[string]interface{}{
		"1": {
			"query": query,
		},
	}
	logging.GetLogger().Debugln("dryrun query underway...")
	response := util.PrepareResultSet(internaldto.NewPrepareResultSetDTO(nil, resultMap, nil, nil, nil, nil))
	responsehandler.HandleResponse(dr.handlerCtx, response) //nolint:errcheck // TODO: investigate
}

func (dr *basicStackQLDriver) ProcessQuery(query string) {
	clonedCtx := dr.handlerCtx.Clone()
	clonedCtx.SetRawQuery(query)
	responses, ok := dr.processQueryOrQueries(clonedCtx)
	if ok {
		for _, r := range responses {
			responsehandler.HandleResponse(clonedCtx, r) //nolint:errcheck // TODO: investigate
		}
	}
}

type basicStackQLDriver struct {
	handlerCtx handler.HandlerContext
}

//nolint:revive // TODO: review
func (dr *basicStackQLDriver) HandleSimpleQuery(ctx context.Context, query string) (sqldata.ISQLResultStream, error) {
	dr.handlerCtx.SetRawQuery(query)
	// if strings.Count(query, ";") > 1 {
	// 	return nil, fmt.Errorf("only support single queries in server mode at this time")
	// }
	res, ok := dr.processQueryOrQueries(dr.handlerCtx)
	if !ok {
		return nil, fmt.Errorf("no SQLresults available")
	}
	r := res[0]
	if r.GetError() != nil {
		return nil, r.GetError()
	}
	return r.GetSQLResult(), nil
}

func (dr *basicStackQLDriver) SplitCompoundQuery(s string) ([]string, error) {
	res := []string{}
	var beg int
	var inDoubleQuotes bool

	for i := 0; i < len(s); i++ {
		if s[i] == ';' && !inDoubleQuotes {
			res = append(res, s[beg:i])
			beg = i + 1
		} else if s[i] == '"' {
			if !inDoubleQuotes {
				inDoubleQuotes = true
			} else if i > 0 && s[i-1] != '\\' {
				inDoubleQuotes = false
			}
		}
	}
	return append(res, s[beg:]), nil
}

func NewStackQLDriver(handlerCtx handler.HandlerContext) (StackQLDriver, error) {
	return &basicStackQLDriver{
		handlerCtx: handlerCtx,
	}, nil
}

func (dr *basicStackQLDriver) processQueryOrQueries(
	handlerCtx handler.HandlerContext,
) ([]internaldto.ExecutorOutput, bool) {
	var retVal []internaldto.ExecutorOutput
	cmdString := handlerCtx.GetRawQuery()
	querySubmitter := querysubmit.NewQuerySubmitter()
	for _, s := range strings.Split(cmdString, ";") {
		if s == "" {
			continue
		}
		clonedCtx := handlerCtx.Clone()
		clonedCtx.SetQuery(s)
		err := querySubmitter.PrepareQuery(clonedCtx)
		if err != nil {
			retVal = append(retVal, internaldto.NewErroneousExecutorOutput(err))
			continue
		}
		retVal = append(retVal, querySubmitter.SubmitQuery())
	}
	return retVal, true
}
