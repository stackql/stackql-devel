package sessionconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"strings"
)

// ApplySessionSet is the shared SET/UNSET value decoder; the host applies authorization.
//
//nolint:exhaustive // Unsupported literal kinds are explicitly rejected by the default branch.
func Apply(host interface {
	SetConfigAtPath(string, interface{}, string) error
}, setExpr *sqlparser.SetExpr) error {
	lhsTrimmed := setExpr.Name.GetRawVal()
	config := strings.HasPrefix(lhsTrimmed, "$")
	if config {
		lhsTrimmed = strings.TrimPrefix(strings.TrimPrefix(lhsTrimmed, "$"), ".")
		root, _, _ := strings.Cut(lhsTrimmed, ".")
		switch root {
		case "auth", "providers", "secret_store", "secrets", "document_merge", "transaction":
		default:
			return fmt.Errorf("unknown intrinsic session configuration %q", lhsTrimmed)
		}
	} else if !strings.HasPrefix(lhsTrimmed, "variables.") {
		lhsTrimmed = "variables." + lhsTrimmed
	}
	if _, reset := setExpr.Expr.(*sqlparser.Default); reset {
		return host.SetConfigAtPath(lhsTrimmed, nil, setExpr.Scope)
	}
	if !config {
		var raw json.RawMessage
		switch value := setExpr.Expr.(type) {
		case *sqlparser.NullVal:
			raw = json.RawMessage("null")
		case sqlparser.BoolVal:
			raw = json.RawMessage(fmt.Sprint(bool(value)))
		case *sqlparser.SQLVal:
			switch value.Type {
			case sqlparser.StrVal:
				raw, _ = json.Marshal(string(value.Val))
			case sqlparser.IntVal, sqlparser.FloatVal:
				raw = json.RawMessage(value.Val)
			default:
				return fmt.Errorf("variables accepts scalar literals only")
			}
		default:
			return fmt.Errorf("variables accepts scalar literals only")
		}
		return host.SetConfigAtPath(lhsTrimmed, raw, setExpr.Scope)
	}
	value, ok := setExpr.Expr.(*sqlparser.SQLVal)
	if !ok || value.Type != sqlparser.StrVal {
		return fmt.Errorf("session configuration requires a string literal or DEFAULT")
	}
	exprStr := string(value.Val)
	if strings.HasPrefix(lhsTrimmed, "secrets.") {
		return host.SetConfigAtPath(lhsTrimmed, exprStr, setExpr.Scope)
	}
	var exprObj map[string]interface{}
	if err := json.Unmarshal([]byte(exprStr), &exprObj); err == nil && exprObj != nil {
		return host.SetConfigAtPath(lhsTrimmed, exprObj, setExpr.Scope)
	}
	return host.SetConfigAtPath(lhsTrimmed, exprStr, setExpr.Scope)
}

var ErrSessionConfiguration = errors.New("session configuration failed")
