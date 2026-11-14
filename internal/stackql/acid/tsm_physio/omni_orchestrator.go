package tsm_physio

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stackql-labs/omnisdk/pkg/omnisdk"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"github.com/stackql/stackql/internal/stackql/acid/txn_context"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/intrinsic"
)

type providerTransactionHost interface {
	handler.HandlerContext
	ProviderTransactionConfigured() bool
	ProviderTransaction() (*omnisdk.TransactionSession, omnisdk.TransactionHandle, bool)
	BeginProviderTransaction(context.Context) error
	ClearProviderTransaction()
}

// SQL syntax/session integration only. Omni owns durability, ordering, locks,
// abort state, commit and compensation. Legacy coordinators remain isolated for
// sessions that have not opted into the new transaction contract.
type omniOrchestrator struct{ legacy Orchestrator }
type queuedSQL struct {
	SQL string `json:"sql"`
}

func (o *omniOrchestrator) legacyQuery(hc handler.HandlerContext, sql string) ([]internaldto.ExecutorOutput, bool) {
	switch legacy := o.legacy.(type) {
	case *standardOrchestrator:
		return legacy.processQuery(hc, sql)
	case *bestEffortOrchestrator:
		return legacy.processQuery(hc, sql)
	default:
		return []internaldto.ExecutorOutput{internaldto.NewErroneousExecutorOutput(fmt.Errorf("unsupported legacy orchestrator"))}, true
	}
}
func (o *omniOrchestrator) ProcessQueryOrQueries(hc handler.HandlerContext) ([]internaldto.ExecutorOutput, bool) {
	host, ok := hc.(providerTransactionHost)
	if !ok {
		return o.legacy.ProcessQueryOrQueries(hc)
	}
	_, _, active := host.ProviderTransaction()
	if !active && !host.ProviderTransactionConfigured() {
		lower := strings.ToLower(hc.GetRawQuery())
		if !strings.Contains(lower, "transaction") && !strings.Contains(lower, "begin") {
			return o.legacy.ProcessQueryOrQueries(hc)
		}
	}
	pieces, err := sqlparser.SplitStatementToPieces(hc.GetRawQuery())
	if err != nil {
		return []internaldto.ExecutorOutput{internaldto.NewErroneousExecutorOutput(err)}, true
	}
	var output []internaldto.ExecutorOutput
	for _, sql := range pieces {
		if sql == "" {
			continue
		}
		var result []internaldto.ExecutorOutput
		_, _, active := host.ProviderTransaction()
		statement, parseErr := sqlparser.Parse(sql)
		if parseErr != nil {
			output = append(output, internaldto.NewErroneousExecutorOutput(parseErr))
			break
		}
		_, begin := statement.(*sqlparser.Begin)
		if !active && (!begin || !host.ProviderTransactionConfigured()) {
			result, _ = o.legacyQuery(hc, sql)
		} else {
			result = o.transactionQuery(host, sql, statement)
		}
		output = append(output, result...)
		failed := false
		for _, r := range result {
			if r.GetError() != nil {
				failed = true
			}
		}
		if failed {
			break
		}
	}
	return output, len(output) > 0
}
func transactionCaller(host providerTransactionHost) (context.Context, func(), error) {
	originalCtx, originalCaller := host.GetExecutionContext(), host.GetCaller()
	manager := host.GetAccessManager()
	if manager == nil {
		return originalCtx, func() {}, nil
	}
	caller, ctx, cancel, err := manager.Establish(originalCtx)
	if err != nil {
		return nil, func() {}, err
	}
	host.SetCaller(caller)
	host.SetExecutionContext(ctx)
	return ctx, func() { host.SetExecutionContext(originalCtx); host.SetCaller(originalCaller); cancel() }, nil
}
func executeProviderSQL(host providerTransactionHost, ctx context.Context, sql string) internaldto.ExecutorOutput {
	local := host.Clone()
	local.SetQuery(sql)
	local.SetExecutionContext(ctx)
	stmt := NewStatement(sql, local, txn_context.NewTransactionContext(0))
	if err := stmt.Prepare(); err != nil {
		return internaldto.NewErroneousExecutorOutput(err)
	}
	return stmt.Execute()
}
func (o *omniOrchestrator) transactionQuery(host providerTransactionHost, sql string, statement sqlparser.Statement) []internaldto.ExecutorOutput {
	requestCtx := host.GetExecutionContext()
	ctx, finish, err := transactionCaller(host)
	if err != nil {
		return []internaldto.ExecutorOutput{internaldto.NewErroneousExecutorOutput(err)}
	}
	defer finish()
	refusal := func(err error) []internaldto.ExecutorOutput {
		return []internaldto.ExecutorOutput{internaldto.NewErroneousExecutorOutput(err)}
	}
	if _, begin := statement.(*sqlparser.Begin); begin {
		if err = host.BeginProviderTransaction(ctx); err != nil {
			return refusal(err)
		}
		_, handle, _ := host.ProviderTransaction()
		return []internaldto.ExecutorOutput{internaldto.NewNopEmptyExecutorOutput([]string{"transaction begun: " + handle.ID()})}
	}
	session, handle, active := host.ProviderTransaction()
	if !active {
		return refusal(fmt.Errorf("no active provider transaction"))
	}
	if intrinsic.ProviderTransactionControlQuery(statement) {
		return []internaldto.ExecutorOutput{executeProviderSQL(host, requestCtx, sql)}
	}
	var output []internaldto.ExecutorOutput
	drain := func() error {
		return session.Drain(ctx, handle, func(runCtx context.Context, payload json.RawMessage) error {
			var queued queuedSQL
			if err := json.Unmarshal(payload, &queued); err != nil {
				return err
			}
			result := executeProviderSQL(host, runCtx, queued.SQL)
			output = append(output, result)
			return result.GetError()
		})
	}
	switch statement.(type) {
	case *sqlparser.Rollback:
		if err = intrinsic.RollbackProviderTransaction(host, session, handle); err != nil {
			return refusal(err)
		}
		host.ClearProviderTransaction()
		return []internaldto.ExecutorOutput{internaldto.NewNopEmptyExecutorOutput([]string{"Rollback OK"})}
	case *sqlparser.Commit:
		if err = drain(); err == nil {
			err = session.Commit(ctx, handle)
		}
		if err != nil {
			return append(output, internaldto.NewErroneousExecutorOutput(err))
		}
		host.ClearProviderTransaction()
		return append(output, internaldto.NewNopEmptyExecutorOutput([]string{"transaction committed"}))
	case *sqlparser.Insert, *sqlparser.Update, *sqlparser.Delete:
		if err = intrinsic.ProviderTransactionMutation(host, statement); err != nil {
			_ = session.Abort(ctx, handle, err)
			return refusal(err)
		}
		cfg, configErr := session.Configuration(ctx, handle)
		if configErr != nil {
			return refusal(configErr)
		}
		needsResult := false
		switch node := statement.(type) {
		case *sqlparser.Insert:
			needsResult = len(node.SelectExprs) > 0
		case *sqlparser.Update:
			needsResult = len(node.SelectExprs) > 0
		case *sqlparser.Delete:
			needsResult = len(node.SelectExprs) > 0
		}
		if cfg.Mode == omnisdk.TransactionLazy && !needsResult {
			if err = host.ResolveVariables(statement); err != nil {
				_ = session.Abort(ctx, handle, err)
				return refusal(err)
			}
			formatted := sqlparser.NewTrackedBuffer(formatQueuedSQL)
			formatted.Myprintf("%v", statement)
			payload, _ := json.Marshal(queuedSQL{SQL: formatted.String()})
			if err = session.Queue(ctx, handle, payload); err != nil {
				return refusal(err)
			}
			return []internaldto.ExecutorOutput{internaldto.NewNopEmptyExecutorOutput([]string{"mutating statement queued"})}
		}
	case *sqlparser.Select, *sqlparser.Union, *sqlparser.ParenSelect:
		// A read is a dependency barrier: preceding writes must be observed.
	default:
		err = fmt.Errorf("statement cannot execute inside a provider transaction")
		_ = session.Abort(ctx, handle, err)
		return refusal(err)
	}
	if err = drain(); err != nil {
		return append(output, internaldto.NewErroneousExecutorOutput(err))
	}
	// Normal submission owns its fresh authentication context and streaming lifetime.
	result := executeProviderSQL(host, requestCtx, sql)
	if result.GetError() != nil {
		_ = session.Abort(ctx, handle, result.GetError())
	}
	return append(output, result)
}

// Preserve identifier boundaries when persisting SQL for later parsing.
func formatQueuedSQL(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {
	if table, ok := node.(sqlparser.TableName); ok {
		first := true
		for _, part := range []sqlparser.TableIdent{table.QualifierThird, table.QualifierSecond, table.Qualifier, table.Name} {
			if part.IsEmpty() {
				continue
			}
			if !first {
				buf.WriteByte('.')
			}
			buf.WriteByte('"')
			buf.WriteString(strings.ReplaceAll(part.GetRawVal(), "\"", "\"\""))
			buf.WriteByte('"')
			first = false
		}
		return
	}
	node.Format(buf)
}
