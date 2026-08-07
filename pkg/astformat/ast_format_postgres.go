package astformat

import (
	"fmt"
	"strings"

	"github.com/stackql/any-sdk/pkg/constants"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

//nolint:revive // Explicit type declaration removes any ambiguity
var (
	_ sqlparser.NodeFormatter = PostgresSelectExprsFormatter
	//nolint:gochecknoglobals // Convenient to have this as a global variable
	sqliteKeywords map[string]struct{} = map[string]struct{}{
		"commit": {},
		"key":    {},
		"select": {},
	}
)

func PostgresSelectExprsFormatter(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {
	switch node := node.(type) {
	case *sqlparser.SQLVal:
		if node.Type == sqlparser.StrVal {
			formatStrVal(node, buf)
			return
		}
		node.Format(buf)
		return
	case *sqlparser.ColName:
		formatColName(node, buf)
	case sqlparser.ColIdent:
		formatColIdent(node, buf)
		return
	case *sqlparser.FuncExpr:
		if strings.ToLower(node.Name.GetRawVal()) == constants.SQLFuncJSONExtractPostgres && len(node.Exprs) > 1 {
			sb := sqlparser.NewTrackedBuffer(PostgresSelectExprsFormatter)
			sb.AstPrintf(
				node,
				"%s(%v%s, %v",
				constants.SQLFuncJSONExtractPostgres,
				node.Exprs[0],
				constants.PostgresJSONCastSuffix,
				node.Exprs[1])
			for _, val := range node.Exprs[2:] {
				sb.AstPrintf(node, ", %v", val)
			}
			sb.AstPrintf(node, ")")
			buf.WriteString(sb.String())
			return
		}
		if strings.ToLower(node.Name.GetRawVal()) == sqlparser.JsonArrayElementsTextStr && len(node.Exprs) >= 1 {
			sb := sqlparser.NewTrackedBuffer(PostgresSelectExprsFormatter)
			sb.AstPrintf(
				node,
				"%s(%v%s",
				sqlparser.JsonArrayElementsTextStr,
				node.Exprs[0],
				constants.PostgresJSONCastSuffix,
			)
			if len(node.Exprs) > 1 {
				for _, val := range node.Exprs[1:] {
					sb.AstPrintf(node, ", %v", val)
				}
			}
			sb.AstPrintf(node, ")")
			buf.WriteString(sb.String())
			return
		}
		node.Format(buf)
		return
	case *sqlparser.GroupConcatExpr:
		sb := sqlparser.NewTrackedBuffer(PostgresSelectExprsFormatter)
		sb.AstPrintf(node, "%s(%v, %s)", constants.SQLFuncGroupConcatPostgres, node.Exprs[0], `','`)
		buf.WriteString(sb.String())
		return

	default:
		node.Format(buf)
		return
	}
}

// formatStrVal renders a string literal with SQL-standard escaping, doubling
// embedded single quotes. The parser's own Format renders MySQL-style
// backslash escapes, which neither SQLite nor Postgres accepts; a literal
// containing an apostrophe would otherwise be re-rendered into a syntax error
// on any round trip through the AST (view DDL, for instance).
func formatStrVal(node *sqlparser.SQLVal, buf *sqlparser.TrackedBuffer) {
	buf.WriteString("'")
	buf.WriteString(strings.ReplaceAll(string(node.Val), "'", "''"))
	buf.WriteString("'")
}

func formatColIdent(node sqlparser.ColIdent, buf *sqlparser.TrackedBuffer) {
	if node.AtCount() > 0 {
		sqlparser.FormatID(buf, node.String(), node.Lowered(), node.AtCount())
		return
	}
	buf.WriteString(fmt.Sprintf(`"%s"`, node.String()))
}

func isSqliteKeyword(term string) bool {
	_, ok := sqliteKeywords[strings.ToLower(term)]
	return ok
}

func formatColIdentCaseInsensitive(node sqlparser.ColIdent, buf *sqlparser.TrackedBuffer) {
	if node.AtCount() > 0 {
		sqlparser.FormatID(buf, node.String(), node.Lowered(), node.AtCount())
		return
	}
	if isSqliteKeyword(node.String()) {
		buf.WriteString(fmt.Sprintf(`"%s"`, node.String()))
		return
	}
	buf.WriteString(node.String())
}

func formatColName(node *sqlparser.ColName, buf *sqlparser.TrackedBuffer) {
	tableNameStr := node.Qualifier.GetRawVal()
	if tableNameStr != "" {
		buf.WriteString(fmt.Sprintf(`"%s".`, tableNameStr))
	}
	formatColIdent(node.Name, buf)
}

func String(node sqlparser.SQLNode, formatter sqlparser.NodeFormatter) string {
	return formattedString(node, formatter)
}

func formattedString(node sqlparser.SQLNode, formatter sqlparser.NodeFormatter) string {
	buf := sqlparser.NewTrackedBuffer(formatter)
	buf.Myprintf("%v", node)
	return buf.String()
}
