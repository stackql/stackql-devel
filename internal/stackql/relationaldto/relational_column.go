package relationaldto

import (
	"fmt"
	"strings"
)

var (
	_ RelationalColumn = &standardRelationalColumn{}
)

type RelationalColumn interface {
	CanonicalSelectionString(string) string
	GetAlias() string
	GetDecorated() string
	GetName() string
	GetQualifier() string
	GetType() string
	GetWidth() int
	WithAlias(string) RelationalColumn
	WithDecorated(string) RelationalColumn
	WithQualifier(string) RelationalColumn
	WithWidth(int) RelationalColumn
}

func NewRelationalColumn(colName string, colType string) RelationalColumn {
	return &standardRelationalColumn{
		colType: colType,
		colName: colName,
	}
}

type standardRelationalColumn struct {
	alias     string
	colType   string
	colName   string
	decorated string
	qualifier string
	width     int
}

func (rc *standardRelationalColumn) CanonicalSelectionString(delim string) string {
	if rc.decorated != "" {
		// if !strings.ContainsAny(rc.decorated, " '`\t\n\"().") {
		// 	return fmt.Sprintf(`"%s" `, rc.decorated)
		// }
		return fmt.Sprintf("%s ", rc.decorated)
	}
	var colStringBuilder strings.Builder
	colStringBuilder.WriteString(fmt.Sprintf(`"%s" `, rc.colName))
	if rc.alias != "" {
		colStringBuilder.WriteString(fmt.Sprintf(` AS "%s"`, rc.alias))
	}
	return colStringBuilder.String()
}

func (rc *standardRelationalColumn) GetName() string {
	return rc.colName
}

func (rc *standardRelationalColumn) GetQualifier() string {
	return rc.qualifier
}

func (rc *standardRelationalColumn) GetType() string {
	return rc.colType
}

func (rc *standardRelationalColumn) GetWidth() int {
	return rc.width
}

func (rc *standardRelationalColumn) GetAlias() string {
	return rc.alias
}

func (rc *standardRelationalColumn) GetDecorated() string {
	return rc.decorated
}

func (rc *standardRelationalColumn) WithDecorated(decorated string) RelationalColumn {
	rc.decorated = decorated
	return rc
}

func (rc *standardRelationalColumn) WithQualifier(qualifier string) RelationalColumn {
	rc.qualifier = qualifier
	return rc
}

func (rc *standardRelationalColumn) WithAlias(alias string) RelationalColumn {
	rc.alias = alias
	return rc
}

func (rc *standardRelationalColumn) WithWidth(width int) RelationalColumn {
	rc.width = width
	return rc
}
