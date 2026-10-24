package intrinsic

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/parserutil"
	"github.com/stackql/stackql/internal/stackql/sql_system"
	"github.com/stackql/stackql/internal/stackql/typing"
)

func materializedFunc(ctx queryContext, stmt sqlparser.Statement) (func() internaldto.ExecutorOutput, bool) {
	var selected sqlparser.SelectStatement
	var name, ddl string
	var replace, refresh bool
	switch node := stmt.(type) {
	case *sqlparser.DDL:
		if !parserutil.IsCreateMaterializedView(node) {
			return nil, false
		}
		selected, name, ddl, replace = node.SelectStatement, qualifiedName(node.Table), sqlparser.String(node), node.OrReplace
	case *sqlparser.RefreshMaterializedView:
		selected, name, refresh = node.ImplicitSelect, qualifiedName(node.ViewName), true
	default:
		return nil, false
	}
	source := selected
	host, ok := ctx.(interface{ GetSQLSystem() sql_system.SQLSystem })
	if !ok {
		return nil, false
	}
	return func() internaldto.ExecutorOutput {
		err := materialize(ctx, host.GetSQLSystem(), source, name, ddl, replace, refresh)
		if err != nil {
			return internaldto.NewErroneousExecutorOutput(err)
		}
		message := "DDL Execution Completed"
		if refresh {
			message = "refresh materialized view completed"
		}
		return internaldto.NewExecutorOutput(nil, nil, nil,
			internaldto.NewBackendMessages([]string{message}), nil)
	}, true
}

// Provider acquisition belongs to Omni; the host only persists the finished relation.
func materialize(
	ctx queryContext, system sql_system.SQLSystem, original sqlparser.SelectStatement,
	name, ddl string, replace, refresh bool,
) error {
	names, open, err := planStatementRead(ctx, original, ctx.GetCurrentProvider(), 0, nil)
	if err != nil {
		return err
	}
	bg := context.Background()
	rows, err := open(bg, nil)
	if err != nil {
		return err
	}
	defer rows.Close()
	manager, err := getStagingManager(ctx)
	if err != nil {
		return err
	}
	id, err := manager.Begin(bg)
	if err != nil {
		return err
	}
	defer manager.Release(bg, id) //nolint:errcheck // The statement outcome takes precedence over cleanup.
	batch := newOmniBatchSource(rows, names, previewCfg.getBatchSize())
	first, err := batch.Next(bg)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	batch.pending = first
	staged := stagedColumns(names, first)
	if _, err = manager.Stage(bg, id, staged, batch); err != nil {
		return err
	}
	if batch.terminalErr != nil {
		return batch.terminalErr
	}
	columns := make([]typing.RelationalColumn, len(staged))
	for i, column := range staged {
		columns[i] = typing.NewRelationalColumn(column.Name(), column.RelationalType())
	}
	query := fmt.Sprintf("SELECT * FROM %s", manager.TableName(id))
	if refresh {
		return system.RefreshMaterializedView(name, columns, query)
	}
	if _, exists := system.GetMaterializedViewByName(name); replace && exists {
		if err = system.DropMaterializedView(name); err != nil {
			return err
		}
	}
	return system.CreateMaterializedView(name, columns, ddl, replace, query)
}
