package intrinsic

import (
	"encoding/json"
	"fmt"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/util"
	"strconv"
	"strings"

	"github.com/stackql-labs/omnisdk/pkg/omnisdk"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

// ProviderTransactionMutation prevents a transaction from silently escaping to
// backend/local-table mutation paths that have no Omni inverse contract.
func ProviderTransactionMutation(ctx queryContext, stmt sqlparser.Statement) error {
	if _, create := stmt.(*sqlparser.Insert); !create {
		return fmt.Errorf("transaction has no supported inverse for this mutation")
	}
	tables, mutation := mutationTables(stmt)
	if !mutation {
		return fmt.Errorf("statement is not a supported provider mutation")
	}
	doc, err := fromDocProviders(tables, ctx.GetCurrentProvider())
	if err != nil {
		return err
	}
	if !doc {
		return fmt.Errorf("transaction mutation must use an Omni provider document")
	}
	return nil
}

func RollbackProviderTransaction(ctx queryContext, session *omnisdk.TransactionSession, handle omnisdk.TransactionHandle) error {
	cfg, err := session.Configuration(executionContext(ctx), handle)
	if err != nil {
		return err
	}
	args, err := providerTransactionArgs(ctx, cfg)
	if err != nil {
		return err
	}
	return session.Rollback(executionContext(ctx), handle, registryRoot(ctx), args)
}

// ProviderTransactionControlQuery identifies only the fixed, authorized control
// relations. They remain usable when ordinary execution is blocked for recovery.
func ProviderTransactionControlQuery(stmt sqlparser.Statement) bool {
	node, ok := stmt.(*sqlparser.Select)
	return ok && providerTransactionRelation(node) != ""
}
func providerTransactionRelation(node *sqlparser.Select) string {
	if len(node.From) != 1 {
		return ""
	}
	aliased, ok := node.From[0].(*sqlparser.AliasedTableExpr)
	if !ok {
		return ""
	}
	table, ok := aliased.Expr.(sqlparser.TableName)
	if !ok {
		return ""
	}
	if table.QualifierSecond.GetRawVal() != "stackql_preview" || table.Qualifier.GetRawVal() != "txn" {
		return ""
	}
	switch table.Name.GetRawVal() {
	case "status":
		return "status"
	case "recover":
		return "recover"
	}
	return ""
}
func providerTransactionControlFunc(ctx queryContext, node *sqlparser.Select) (func() internaldto.ExecutorOutput, bool) {
	relation := providerTransactionRelation(node)
	if relation == "" {
		return nil, false
	}
	host, ok := ctx.(interface {
		ProviderTransaction() (*omnisdk.TransactionSession, omnisdk.TransactionHandle, bool)
	})
	if !ok {
		return refuse(fmt.Errorf("provider transaction session is unavailable")), true
	}
	return func() internaldto.ExecutorOutput {
		session, handle, active := host.ProviderTransaction()
		if !active {
			return internaldto.NewErroneousExecutorOutput(fmt.Errorf("no active provider transaction; BEGIN or resume first"))
		}
		if relation == "recover" {
			values, unhandled := equalityPredicates(node.Where)
			if len(unhandled) > 0 {
				return internaldto.NewErroneousExecutorOutput(fmt.Errorf("transaction recovery requires equality predicates"))
			}
			recovery := omnisdk.TransactionRecovery{Version: values["version"]}
			if recovery.Version == "" {
				return internaldto.NewErroneousExecutorOutput(fmt.Errorf("transaction recovery requires the observed transaction version"))
			}
			if raw := values["resources"]; raw != "" {
				if err := json.Unmarshal([]byte(raw), &recovery.Resources); err != nil {
					return internaldto.NewErroneousExecutorOutput(fmt.Errorf("invalid recovery resources"))
				}
			}
			if raw := values["continue"]; raw != "" {
				value, err := strconv.ParseBool(raw)
				if err != nil {
					return internaldto.NewErroneousExecutorOutput(fmt.Errorf("continue must be true or false"))
				}
				recovery.Continue = value
			}
			for key := range values {
				if key != "version" && key != "resources" && key != "continue" {
					return internaldto.NewErroneousExecutorOutput(fmt.Errorf("unknown transaction recovery input %s", key))
				}
			}
			if err := session.Recover(executionContext(ctx), handle, recovery); err != nil {
				return internaldto.NewErroneousExecutorOutput(err)
			}
		} else if node.Where != nil {
			return internaldto.NewErroneousExecutorOutput(fmt.Errorf("transaction status takes no predicates"))
		}
		snapshot, err := session.Status(executionContext(ctx), handle)
		if err != nil {
			return internaldto.NewErroneousExecutorOutput(err)
		}
		pending, err := json.Marshal(snapshot.Pending)
		if err != nil {
			return internaldto.NewErroneousExecutorOutput(err)
		}
		return prepare(ctx, []string{"id", "status", "version", "pending"},
			map[string]map[string]interface{}{"0": {"id": snapshot.ID, "status": snapshot.Status, "version": snapshot.Version, "pending": string(pending)}}, util.DefaultRowSort)
	}, true
}

func providerTransactionArgs(ctx queryContext, cfg omnisdk.TransactionConfig) (omnisdk.Args, error) {
	bundles := []string{}
	for address := range cfg.Targets {
		bundle, _, _ := strings.Cut(address, ".")
		bundles = append(bundles, bundle)
	}
	cloud := ""
	if len(bundles) > 0 {
		cloud = bundles[0]
	}
	args, err := previewArgs(ctx, cloud, nil)
	if err != nil {
		return omnisdk.Args{}, err
	}
	args.AuthByProvider, err = queryProviderAuth(ctx, bundles)
	if err != nil {
		return omnisdk.Args{}, err
	}
	return args, nil
}
func OpenProviderTransaction(ctx queryContext, session *omnisdk.TransactionSession, cfg omnisdk.TransactionConfig) (omnisdk.TransactionHandle, error) {
	args, err := providerTransactionArgs(ctx, cfg)
	if err != nil {
		return omnisdk.TransactionHandle{}, err
	}
	return session.BeginWithRegistry(executionContext(ctx), cfg, registryRoot(ctx), args)
}
